//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"fmt"
	"github.com/RedisAI/redisai-go/redisai"
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/fraud"
	"github.com/RedisAI/aibench/inference"
	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
	"log"

	//ignoring until we get the correct model
	//"log"
	"sync"
	"time"
)

// Program option vars:
var (
	host        string
	model       string
	showExplain bool
)

// Global vars:
var (
	runner        *inference.BenchmarkRunner
	inferenceType = "RedisAI Query - with AI.TENSORSET transacation datatype BLOB"
	cpool         *redis.Pool
)

// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()
	flag.StringVar(&host, "host", "redis://localhost:6379", "Redis host address and port")
	flag.StringVar(&model, "model", "", "model name")
	flag.Parse()
	cpool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(host) },
	}
}

func main() {
	runner.Run(&inference.RedisAIPool, newProcessor)
}

type queryExecutorOptions struct {
	showExplain   bool
	debug         bool
	printResponse bool
}

type Processor struct {
	opts    *queryExecutorOptions
	Metrics chan uint64
	Wg      *sync.WaitGroup
	pclient *redisai.Client
}

func (p *Processor) Close() {
	if p.pclient != nil {
		p.pclient.Close()
	}
}

func newProcessor() inference.Processor { return &Processor{} }

func (p *Processor) Init(numWorker int, wg *sync.WaitGroup, m chan uint64, rs chan uint64) {
	p.opts = &queryExecutorOptions{
		showExplain:   showExplain,
		debug:         runner.DebugLevel() > 0,
		printResponse: runner.DoPrintResponses(),
	}
	p.Wg = wg
	p.Metrics = m
	p.pclient = redisai.Connect(host, cpool)
	p.pclient.Pipeline(3)
}

func (p *Processor) ProcessInferenceQuery(q []byte, isWarm bool) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}
	idUint64 := fraud.Uint64frombytes(q[0:8])
	idS := fmt.Sprintf("%d", idUint64)
	referenceDataTensorName := "referenceTensor:" + idS
	classificationTensorName := "classificationTensor:" + idS
	transactionDataTensorName := "transactionTensor:" + idS
	transactionValues := q[8:128]

	took := 0.0
	start := time.Now()
	p.pclient.TensorSet(transactionDataTensorName, redisai.TypeFloat, []int{1, 30}, transactionValues)
	p.pclient.ModelRun(model, []string{transactionDataTensorName, referenceDataTensorName}, []string{classificationTensorName})
	p.pclient.TensorGet(classificationTensorName, redisai.TensorContentTypeBlob)
	err := p.pclient.Flush()
	if err != nil {
		extendedError := fmt.Errorf("Prediction Receive() failed:%v\n", err)
		if runner.IgnoreErrors(){
			fmt.Println(extendedError)
		} else{
			log.Fatal(extendedError )
		}
	}
	p.pclient.Receive()
	p.pclient.Receive()
	resp, err := p.pclient.Receive()
	data, err := redisai.ProcessTensorReplyMeta(resp, err)
	PredictResponse, err := redisai.ProcessTensorReplyBlob(data, err)
	took = float64(time.Since(start).Nanoseconds()) / 1e6
	if err != nil {
		extendedError := fmt.Errorf("Prediction Receive() failed:%v\n", err)
		if runner.IgnoreErrors(){
			fmt.Println(extendedError)
		} else{
			log.Fatal(extendedError )
		}
	}
	if p.opts.printResponse {
		if err != nil {
			extendedError := fmt.Errorf("Response parsing failed:%v\n", err)
			if runner.IgnoreErrors(){
				fmt.Println(extendedError)
			} else{
				log.Fatal(extendedError )
			}
		}
		fmt.Println("RESPONSE: ", PredictResponse[2])
	}
	// VALUES
	stat := inference.GetStat()
	stat.Init([]byte(inferenceType), took, uint64(0), false, "")

	return []*inference.Stat{stat}, nil
}
