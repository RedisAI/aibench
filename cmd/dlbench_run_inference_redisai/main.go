//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"github.com/filipecosta90/dlbench/inference"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"log"
	"strings"
	"sync"
	"time"
)

// Program option vars:
var (
	host  string
	index string

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
	opts          *queryExecutorOptions
	Metrics       chan uint64
	Wg            *sync.WaitGroup
}

func newProcessor() inference.Processor { return &Processor{} }

func (p *Processor) Init(numWorker int, wg *sync.WaitGroup, m chan uint64, rs chan uint64) {
	p.Wg = wg
	p.Metrics = m

	p.opts = &queryExecutorOptions{
		showExplain:   showExplain,
		debug:         runner.DebugLevel() > 0,
		printResponse: runner.DoPrintResponses(),
	}
}

func (p *Processor) ProcessQuery(q inference.Query, isWarm bool) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}
	tq := q.(*inference.RedisAI)

	qry := string(tq.RedisQuery)

	t := strings.Split(qry, ",")
	//if len(t) < 2 {
	//	log.Fatalf("The inference has not the correct format ", qry)
	//}
	//command := t[0]
	//if command != "FT.SEARCH" {
	//	log.Fatalf("Command not supported yet. Only FT.SEARCH. ", command)
	//}
	//rediSearchQuery := redisearch.NewQuery(t[1])
	start := time.Now()
	//client.Do("AI.TENSORSET", "transactionX", "FLOAT32", 1,  37, "BLOB", imgbuf.Bytes())
	pipe := client.Pipeline()
	pipe.Do("PING")
	pipe.Do("PING")
	pipe.Do("PING")
	// Execute
	//
	//     INCR pipeline_counter
	//     EXPIRE pipeline_counts 3600
	//
	// using one redisdb-server roundtrip.
	_, err := pipe.Exec()
	took := float64(time.Since(start).Nanoseconds()) / 1e6

	if err != nil {

		log.Fatalf("Command failed:%v\tError message:%v\tString Error message:|%s|\n", err, err.Error())

	}
	stat := inference.GetStat()
	stat.Init(q.HumanLabelName(), took, uint64(0), false, t[1])

	return []*inference.Stat{stat}, nil
}
