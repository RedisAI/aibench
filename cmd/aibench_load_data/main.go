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
	"sync"
	"time"
)

// Program option vars:
var (
	host         string
	pipelineSize uint
)

// Global vars:
var (
	runner *inference.LoadRunner
	cpool  *redis.Pool
)

// Parse args:
func init() {
	runner = inference.NewLoadRunner()
	flag.StringVar(&host, "host", "redis://localhost:6379", "Redis host address and port")
	flag.UintVar(&pipelineSize, "pipeline", 10, "Redis pipeline size")

	flag.Parse()

	cpool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(host) },
	}
}

func main() {
	runner.RunLoad(&inference.RedisAIPool, newProcessor)
}

type Loader struct {
	Wg      *sync.WaitGroup
	pclient *redisai.Client
}

func (p *Loader) Close() {
	p.pclient.Close()
}

func newProcessor() inference.Loader { return &Loader{} }

func (p *Loader) Init(numWorker int, wg *sync.WaitGroup) {
	p.Wg = wg
	p.pclient = redisai.Connect(host, cpool)
	p.pclient.Pipeline(uint32(pipelineSize))
}

func (p *Loader) ProcessLoadQuery(q []byte, debug int) ([]*inference.Stat, uint64, error) {
	if len(q) != (1024 + 8 + 120) {
		log.Fatalf("wrong Row lenght. Expected Set:%d got %d\n", (1024 + 8 + 120), len(q))
	}
	tmp := make([]byte, 8)
	referenceValues := make([]byte, 1024)
	copy(tmp, q[0:8])
	copy(referenceValues, q[128:1152])

	idF := fraud.Uint64frombytes(tmp)
	if debug > 0 {
		//fmt.Printf("On Row: %d\n", idF )
	}
	id := "referenceTensor:" + fmt.Sprintf("%d", int(idF))
	idBlob := "referenceBLOB:" + fmt.Sprintf("%d", int(idF))
	p.pclient.ActiveConnNX()
	errSet := p.pclient.ActiveConn.Send("SET", idBlob, referenceValues)
	if errSet != nil {
		log.Fatal(errSet)
	}
	err := p.pclient.TensorSet(id, redisai.TypeFloat, []int{1, 256}, referenceValues)
	if err != nil {
		log.Fatal(err)
	}
	return nil, uint64(2), nil
}
