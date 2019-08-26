//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"fmt"
	"github.com/filipecosta90/aibench/cmd/aibench_generate_data/fraud"
	"github.com/filipecosta90/aibench/inference"
	"github.com/filipecosta90/redisai-go/redisai"
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
	host  string
	model string
	showExplain bool
)

// Global vars:
var (
	runner *inference.BenchmarkRunner
	inferenceType = "RedisAI Query - with AI.TENSORSET transacation datatype BLOB"
	cpool *redis.Pool

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
	pclient *redisai.PipelinedClient
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
	p.pclient = redisai.ConnectPipelined(host, 3, cpool )
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
	p.pclient.TensorGet(classificationTensorName, redisai.TensorContentTypeValues)
	err := p.pclient.ForceFlush()
	if err != nil {
		log.Fatalf("Prediction failed:%v\n", err)
	}
	p.pclient.ActiveConn.Receive()
	p.pclient.ActiveConn.Receive()
	PredictResponse, err := p.pclient.ActiveConn.Receive()
	took = float64(time.Since(start).Nanoseconds()) / 1e6

	if p.opts.printResponse {
		fmt.Println("RESPONSE: ", PredictResponse)
	}
	// VALUES
	stat := inference.GetStat()
	stat.Init([]byte(inferenceType), took, uint64(0), false, "")

	return []*inference.Stat{stat}, nil
}
