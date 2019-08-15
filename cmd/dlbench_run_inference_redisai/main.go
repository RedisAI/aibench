//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"encoding/binary"
	"flag"
	"github.com/filipecosta90/dlbench/inference"
	redisai "github.com/filipecosta90/dlbench/redisai-go"
	"github.com/go-redis/redis"
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

var (
	client *redis.Client
)

// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()

	flag.StringVar(&host, "host", "localhost:6379", "Redis host address and port")
	flag.StringVar(&model, "model", "", "model name")
	flag.BoolVar(&blob, "use-blob", false, "Use blob as data format (time to convert from []float to []byte)")
	flag.Parse()
	client = redis.NewClient(&redis.Options{
		Addr: host,
	})
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
}

func newProcessor() inference.Processor { return &Processor{} }

func (p *Processor) Init(numWorker int, wg *sync.WaitGroup, m chan uint64, rs chan uint64) {
	p.Wg = wg
	p.Metrics = m
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
		modelrunBLOB_args := redisai.Generate_AI_ModelRun_Args(model, []string{transactionTensorBLOBName, referenceDataTensorName}, []string{classificationTensorName})
		tensorget_args := redisai.Generate_AI_TensorGet_Args(classificationTensorName, "VALUES")
		start := time.Now()
		qbytes := Float32bytes(qfloat[0])
		for _, value := range qfloat[1:30] {
			qbytes = append( qbytes, Float32bytes(value)... )
		}
		tensorsetBLOB_args := redisai.Generate_AI_TensorSetBLOB_Args(transactionTensorBLOBName, "FLOAT", []int{1, 30}, "BLOB", qbytes)
		pipe := client.Pipeline()
		pipe.Do(tensorsetBLOB_args...)
		pipe.Do(modelrunBLOB_args...)
		pipe.Do(tensorget_args...)
		_, err := pipe.Exec()
		took = float64(time.Since(start).Nanoseconds()) / 1e6
		if err != nil {
			log.Fatalf("Command failed:%v\n", err)
		}
		queryType = queryType + " - with AI.TENSORSET transacation datatype BLOB"
		// VALUES
	} else {
		transactionTensorName := "transacation:" + q[0]
		tensorset_args := redisai.Generate_AI_TensorSet_Args(transactionTensorName, "FLOAT", []int{1, 30}, "VALUES", qfloat)
		tensorget_args := redisai.Generate_AI_TensorGet_Args(classificationTensorName, "VALUES")
		modelrun_args := redisai.Generate_AI_ModelRun_Args(model, []string{transactionTensorName, referenceDataTensorName}, []string{classificationTensorName})
		start := time.Now()
		pipe := client.Pipeline()
		pipe.Do(tensorset_args...)
		pipe.Do(modelrun_args...)
		pipe.Do(tensorget_args...)
		_, err := pipe.Exec()
		took = float64(time.Since(start).Nanoseconds()) / 1e6
		if err != nil {
			log.Fatalf("Command failed:%v\n", err)
		}
		queryType = queryType + " - with AI.TENSORSET transacation datatype VALUES"
	}

	stat := inference.GetStat()
	stat.Init([]byte(queryType), took, uint64(0), false, "")

	return []*inference.Stat{stat}, nil
}
