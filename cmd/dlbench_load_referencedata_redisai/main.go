//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"github.com/filipecosta90/dlbench/inference"
	redisai "github.com/filipecosta90/dlbench/redisai-go"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	//ignoring until we get the correct model
	//"log"
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
	client *redis.Client
)

// Parse args:
func init() {
	runner = inference.NewLoadRunner()

	flag.StringVar(&host, "host", "localhost:6379", "Redis host address and port")

	flag.Parse()
	client = redis.NewClient(&redis.Options{
		Addr: host,
	})
}

func main() {
	runner.RunLoad(&inference.RedisAIPool, newProcessor)
}

type Loader struct {
	Wg *sync.WaitGroup
}

func newProcessor() inference.Loader { return &Loader{} }

func (p *Loader) Init(numWorker int, wg *sync.WaitGroup) {
	p.Wg = wg
}

func (p *Loader) ProcessLoadQuery(q []string) ([]*inference.Stat, error) {

	referenceDataTensorName := "reference:" + q[0]
	tensorset_args := redisai.Generate_AI_TensorSet_Args(referenceDataTensorName, "FLOAT", []int{30}, q[1:31])
	client.Do(tensorset_args...)
	//ignoring until we get the correct model
	//_, err := pipe.Exec()
	//ignoring until we get the correct model
	//
	//if err != nil {
	//	log.Fatalf("Command failed:%v\n", err)
	//
	//}

	return nil, nil
}
