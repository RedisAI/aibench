//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"github.com/filipecosta90/dlbench/inference"
	redisai "github.com/filipecosta90/dlbench/redisai-go"
	"github.com/go-redis/redis"
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
)

var (
	client *redis.Client
)

// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()

	flag.StringVar(&host, "host", "localhost:6379", "Redis host address and port")
	flag.StringVar(&model, "model", "", "model name")

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

func (p *Processor) ProcessInferenceQuery(q []string, isWarm bool) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}
	transactionTensorName := "transacation:" + q[0]
	referenceDataTensorName := "referenceTensor:" + q[0]
	classificationTensorName := "classification:" + q[0]
	tensorset_args := redisai.Generate_AI_TensorSet_Args(transactionTensorName, "FLOAT", []int{1, 30}, q[1:31])
	modelrun_args := redisai.Generate_AI_ModelRun_Args(model, []string{transactionTensorName, referenceDataTensorName}, []string{classificationTensorName})
	tensorget_args := redisai.Generate_AI_TensorGet_Args(classificationTensorName)
	pipe := client.Pipeline()
	start := time.Now()
	pipe.Do(tensorset_args...)
	pipe.Do(modelrun_args...)
	pipe.Do(tensorget_args...)
	_, err := pipe.Exec()
	took := float64(time.Since(start).Nanoseconds()) / 1e6
	if err != nil {
		log.Fatalf("Command failed:%v\n", err)
	}
	stat := inference.GetStat()
	stat.Init([]byte("RedisAI Query"), took, uint64(0), false, "")

	return []*inference.Stat{stat}, nil
}
