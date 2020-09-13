//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"fmt"
	aibench "github.com/RedisAI/aibench/inference"
	"github.com/RedisAI/redisai-go/redisai"
	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
	"log"
	"sync"
	"time"
)

// Program option vars:
var (
	host               string
	pipelineSize       uint
	setBlob            bool
	setTensor          bool
	runner             *aibench.LoadRunner
	cpool              *redis.Pool
	rowBenchmarkNBytes = 8 + 120 + 1024
)

// Parse args:
func init() {
	runner = aibench.NewLoadRunner()
	flag.StringVar(&host, "redis-host", "redis://localhost:6379", "Redis host address and port")
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
	runner.RunLoad(&aibench.RedisAIPool, newProcessor, 0)
}

type Loader struct {
	Wg      *sync.WaitGroup
	pclient *redisai.Client
}

func (p *Loader) Close() {
	p.pclient.Close()
}

func newProcessor() aibench.Loader { return &Loader{} }

func (p *Loader) Init(numWorker int, wg *sync.WaitGroup) {
	p.Wg = wg
	p.pclient = redisai.Connect(host, cpool)
	p.pclient.Pipeline(uint32(pipelineSize))
}

func (p *Loader) ProcessLoadQuery(q []byte, debug int) ([]*aibench.Stat, uint64, error) {
	if(debug>0){
		fmt.Println("adad")
	}
	if len(q) != (1024 + 8 + 120) {
		log.Fatalf("wrong Row lenght. Expected Set:%d got %d\n", 1024+8+120, len(q))
	}
	tmp := make([]byte, 8)
	referenceValues := make([]byte, 1024)
	copy(tmp, q[0:8])
	copy(referenceValues, q[128:1152])

	idF := aibench.Uint64frombytes(tmp)
	id := "referenceTensor:{" + fmt.Sprintf("%d", int(idF)) + "}"
	idBlob := "referenceBLOB:{" + fmt.Sprintf("%d", int(idF)) + "}"
	issuedCommands := 0
	p.pclient.ActiveConnNX()
	if setBlob {
		errSet := p.pclient.ActiveConn.Send("SET", idBlob, referenceValues)
		if errSet != nil {
			log.Fatal(errSet)
		}
		issuedCommands++
	}
	if setTensor {
		err := p.pclient.TensorSet(id, redisai.TypeFloat, []int64{1, 256}, referenceValues)
		if err != nil {
			log.Fatal(err)
		}
		issuedCommands++
	}

	return nil, uint64(issuedCommands), nil
}
