//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"fmt"
	"github.com/filipecosta90/aibench/cmd/aibench_generate_data/fraud"
	"github.com/filipecosta90/aibench/inference"
	"github.com/filipecosta90/redisai-go/redisai"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"log"
	"sync"
)

// Program option vars:
var (
	host string
)

// Global vars:
var (
	runner *inference.LoadRunner
)

var (
	_ *redis.Client
)

// Parse args:
func init() {
	runner = inference.NewLoadRunner()
	flag.StringVar(&host, "host", "redis://localhost:6379", "Redis host address and port")
	flag.Parse()
}

func main() {
	runner.RunLoad(&inference.RedisAIPool, newProcessor)
}

type Loader struct {
	Wg *sync.WaitGroup
	pclient *redisai.PipelinedClient
}

func (p *Loader) Close() {
	if p.pclient != nil {
		p.pclient.Close()
	}
}

func newProcessor() inference.Loader { return &Loader{} }

func (p *Loader) Init(numWorker int, wg *sync.WaitGroup) {
	p.Wg = wg
	p.pclient = redisai.ConnectPipelined(host, 10)
}

func (p *Loader) ProcessLoadQuery(q []byte) ([]*inference.Stat, error) {
	idF := fraud.Uint64frombytes(q[0:8])
	id := "referenceTensor:" + fmt.Sprintf("%d", idF)
	referenceValues := q[128:1152]
	err := p.pclient.TensorSet(id, redisai.TypeFloat, []int{1, 256}, referenceValues)
	if err != nil {
		log.Fatal(err)
	}
	return nil, nil
}
