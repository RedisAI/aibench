//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"fmt"
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/fraud"
	"github.com/RedisAI/aibench/inference"
	"github.com/RedisAI/redisai-go/redisai"
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
	setBlob      bool
	setTensor    bool
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
	flag.BoolVar(&setBlob, "set-blob", true, "Set reference data in plain binary safe Redis string format")
	flag.BoolVar(&setTensor, "set-tensor", true, "Set reference data in AI.TENSOR format")
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
	referenceValues := make([]byte, 16)
	copy(tmp, q[0:8])
	copy(referenceValues, q[128:144])

	idF := fraud.Uint64frombytes(tmp)
	id := "referenceTensor:{" + fmt.Sprintf("%d", int(idF)) + "}"
	idBlob := "referenceBLOB:{" + fmt.Sprintf("%d", int(idF)) + "}"
	p.pclient.ActiveConnNX()
	issuedCommands := 0
	if setBlob {
		errSet := p.pclient.ActiveConn.Send("SET", idBlob, referenceValues)
		if errSet != nil {
			log.Fatal(errSet)
		}
		issuedCommands++
	}
	if setTensor {
		err := p.pclient.TensorSet(id, redisai.TypeFloat, []int{1, 4}, referenceValues)
		if err != nil {
			log.Fatal(err)
		}
		issuedCommands++
	}

	return nil, uint64(issuedCommands), nil
}
