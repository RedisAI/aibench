//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/filipecosta90/aibench/inference"
	"github.com/filipecosta90/redisai-go/redisai"
	_ "github.com/lib/pq"
	"log"
	math2 "math"
	"strconv"

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
	blob bool
)

// Global vars:
var (
	runner *inference.BenchmarkRunner
)


// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()

	flag.StringVar(&host, "host", "redis://localhost:6379", "Redis host address and port")
	flag.StringVar(&model, "model", "", "model name")
	flag.BoolVar(&blob, "use-blob", true, "Use blob as data format (time to convert from []float to []byte)")
	flag.Parse()

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
	if p.pclient !=nil {
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
	p.pclient = redisai.ConnectPipelined(host, 3)
}

func Float32bytes(float float32) []byte {
	bits := math2.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}


func convertSliceStringToFloat(transactionDataString []string) []float32 {
	res := make([]float32, len(transactionDataString))
	for i := range transactionDataString {
		value, _ := strconv.ParseFloat(transactionDataString[i], 64)
		res[i] = float32(value)
	}
	return res
}

func (p *Processor) ProcessInferenceQuery(q []string, isWarm bool) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}

	// COMMON
	referenceDataTensorName := "referenceTensor:" + q[0]
	classificationTensorName := "classification:" + q[0]
	took := 0.0
	queryType := "RedisAI Query"
	// BLOB
	qfloat := convertSliceStringToFloat( q[1:31] )

	if blob {
		transactionTensorBLOBName := "transacationBLOB:" + q[0]
		qbytes := Float32bytes(qfloat[0])
		for _, value := range qfloat[1:30] {
			qbytes = append(qbytes, Float32bytes(value)...)
		}
		start := time.Now()
		p.pclient.TensorSet(transactionTensorBLOBName, redisai.TypeFloat, []int{1, 30}, qbytes)
		p.pclient.ModelRun(model, []string{transactionTensorBLOBName, referenceDataTensorName}, []string{classificationTensorName})
		p.pclient.TensorGet(classificationTensorName, redisai.TensorContentTypeBlob)
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
		queryType = queryType + " - with AI.TENSORSET transacation datatype BLOB"
		// VALUES
	} else {
		transactionTensorName := "transacation:" + q[0]
		start := time.Now()
		p.pclient.TensorSet(transactionTensorName, redisai.TypeFloat, []int{1, 30}, qfloat)
		p.pclient.ModelRun(model, []string{transactionTensorName, referenceDataTensorName}, []string{classificationTensorName})
		p.pclient.TensorGet(classificationTensorName, redisai.TensorContentTypeBlob)

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
	}

	stat := inference.GetStat()
	stat.Init([]byte(queryType), took, uint64(0), false, "")

	return []*inference.Stat{stat}, nil
}
