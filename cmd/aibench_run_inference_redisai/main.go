//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/RedisAI/aibench/inference"
	_ "github.com/lib/pq"
	"github.com/mediocregopher/radix/v3"
	"sync"
)

// Global vars:
var (
	runner                  *inference.BenchmarkRunner
	host                    string
	port                    string
	model                   string
	modelFilename           string
	useDag                  bool
	showExplain             bool
	clusterMode             bool
	PoolPipelineConcurrency int
	PoolPipelineWindow      time.Duration
	rowBenchmarkNBytes      = 8 + 120 + 1024
	inferenceType           = "RedisAI Query - with AI.TENSORSET transacation datatype BLOB"
)

// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()
	flag.StringVar(&host, "host", "localhost", "Redis host address, if more than one is passed will round robin requests")
	flag.StringVar(&port, "port", "6379", "Redis host port, if more than one is passed will round robin requests")
	flag.StringVar(&model, "model", "", "model name")
	flag.StringVar(&modelFilename, "model-filename", "", "modelFilename")
	flag.BoolVar(&useDag, "use-dag", false, "use DAGRUN")
	flag.BoolVar(&clusterMode, "cluster-mode", false, "read cluster slots and distribute inferences among shards.")
	flag.DurationVar(&PoolPipelineWindow, "pool-pipeline-window", 500*time.Microsecond, "If window is zero then implicit pipelining will be disabled")
	flag.IntVar(&PoolPipelineConcurrency, "pool-pipeline-concurrency", 0, "If limit is zero then no limit will be used and pipelines will only be limited by the specified time window")
	flag.Parse()
}

func main() {
	runner.Run(&inference.RedisAIPool, newProcessor, rowBenchmarkNBytes, 1)
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
	pclient []*radix.Pool
}

func (p *Processor) CollectRunTimeMetrics() (ts int64, stats interface{}, err error) {
	// TODO:
	return
}

func (p *Processor) Close() {
	if p.pclient != nil {
		for _, client := range p.pclient {
			client.Close()
		}
	}
}

func newProcessor() inference.Processor { return &Processor{} }

func (p *Processor) Init(numWorker int, totalWorkers int, wg *sync.WaitGroup, m chan uint64, rs chan uint64) {
	p.opts = &queryExecutorOptions{
		showExplain:   showExplain,
		debug:         runner.DebugLevel() > 0,
		printResponse: runner.DoPrintResponses(),
	}
	p.Wg = wg
	p.Metrics = m
	var err error

	hosts := strings.Split(host, ",")
	ports := strings.Split(port, ",")

	// if we have more hosts than workers lets connect to them all
	if len(hosts) > totalWorkers {
		p.pclient = make([]*radix.Pool, len(hosts))
		for idx, h := range hosts {
			p.pclient[idx], err = radix.NewPool("tcp", fmt.Sprintf("%s:%s", h, ports[idx]), 1, radix.PoolPipelineWindow(0, 0))
			if err != nil {
				log.Fatalf("Error preparing for DAGRUN(), while creating new pool. error = %v", err)
			}
		}

	} else {
		pos := (numWorker + 1) % len(hosts)
		p.pclient = make([]*radix.Pool, 1)
		p.pclient[0], err = radix.NewPool("tcp", fmt.Sprintf("%s:%s", hosts[pos], ports[pos]), 1, radix.PoolPipelineWindow(0, 0))
		if err != nil {
			log.Fatalf("Error preparing for DAGRUN(), while creating new pool. error = %v", err)
		}
	}

}

func (p *Processor) ProcessInferenceQuery(q []byte, isWarm bool, workerNum int, useReferenceDataRedis bool, useReferenceDataMysql bool, queryNumber int64) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}
	idUint64 := inference.Uint64frombytes(q[0:8])
	idS := fmt.Sprintf("%d", idUint64)
	referenceDataTensorName := "referenceTensor:{" + idS + "}"
	classificationTensorName := "classificationTensor:{" + idS + "}"
	transactionDataTensorName := "transactionTensor:{" + idS + "}"
	transactionValues := q[8:128]
	var args []string
	if useReferenceDataRedis {
		args = []string{"LOAD", "1", referenceDataTensorName, "|>",
			"AI.TENSORSET", transactionDataTensorName, "FLOAT", "1", "30", "BLOB", string(transactionValues), "|>",
			"AI.MODELRUN", model, "INPUTS", transactionDataTensorName, referenceDataTensorName, "OUTPUTS", classificationTensorName, "|>",
			"AI.TENSORGET", classificationTensorName, "BLOB",
		}
	} else {
		args = []string{"|>",
			"AI.TENSORSET", transactionDataTensorName, "FLOAT", "1", "30", "BLOB", string(transactionValues), "|>",
			"AI.MODELRUN", model, "INPUTS", transactionDataTensorName, "OUTPUTS", classificationTensorName, "|>",
			"AI.TENSORGET", classificationTensorName, "BLOB",
		}
	}
	pos := rand.Int31n(int32(len(p.pclient)))
	start := time.Now()

	err := p.pclient[pos].Do(radix.Cmd(nil, "AI.DAGRUN", args...))
	if err != nil {
		extendedError := fmt.Errorf("Prediction Receive() failed:%v\n", err)
		log.Fatal(extendedError)
	}

	took := time.Since(start).Microseconds()

	stat := inference.GetStat()
	stat.Init([]byte(inferenceType), took, uint64(0), false, "")

	return []*inference.Stat{stat}, nil
}
