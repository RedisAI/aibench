//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"fmt"
	aibench "github.com/RedisAI/aibench/inference"
	"github.com/RedisAI/redisai-go/redisai"
	_ "github.com/lib/pq"
	"log"
	"sync"
)

// Program option vars:
var (
	host               string
	pipelineSize       uint
	setBlob            bool
	setTensor          bool
	runner             *aibench.LoadRunner
	rowBenchmarkNBytes = 8 + 120 + 1024
)

// Parse args:
func init() {
	runner = aibench.NewLoadRunner()
	flag.StringVar(&host, "redis-host", "redis://localhost:6379", "Redis host address and port")
	flag.UintVar(&pipelineSize, "pipeline", 1, "Redis pipeline size")
	flag.BoolVar(&setBlob, "set-blob", true, "Set reference data in plain binary safe Redis string format")
	flag.BoolVar(&setTensor, "set-tensor", true, "Set reference data in AI.TENSOR format")
	flag.Parse()
}

func main() {
	runner.RunLoad(&aibench.RedisAIPool, newProcessor, rowBenchmarkNBytes)
}

type Loader struct {
	Wg       *sync.WaitGroup
	aiClient *redisai.Client
}

func (p *Loader) Close() {
	p.aiClient.Close()
}

func newProcessor() aibench.Loader { return &Loader{} }

func (p *Loader) Init(numWorker int, wg *sync.WaitGroup) {
	p.Wg = wg
	p.aiClient = redisai.Connect(host, nil)
	p.aiClient.Pipeline(uint32(pipelineSize))
}

func (p *Loader) ProcessLoadQuery(q []byte, debug int) ([]*aibench.Stat, uint64, error) {
	if len(q) != rowBenchmarkNBytes {
		log.Fatalf("wrong Row lenght. Expected Set:%d got %d\n", rowBenchmarkNBytes, len(q))
	}
	tmp := make([]byte, 8)
	referenceValues := make([]byte, 1024)
	copy(tmp, q[0:8])
	copy(referenceValues, q[128:1152])

	idF := aibench.Uint64frombytes(tmp)
	id := "referenceTensor:{" + fmt.Sprintf("%d", int(idF)) + "}"
	idBlob := "referenceBLOB:{" + fmt.Sprintf("%d", int(idF)) + "}"
	issuedCommands := 0
	p.aiClient.ActiveConnNX()
	if setBlob {
		errSet := p.aiClient.ActiveConn.Send("SET", idBlob, referenceValues)
		if errSet != nil {
			log.Fatal(errSet)
		}
		issuedCommands++
	}
	if setTensor {
		err := p.aiClient.TensorSet(id, redisai.TypeFloat, []int64{1, 256}, referenceValues)
		if err != nil {
			log.Fatal(err)
		}
		issuedCommands++
	}

	return nil, uint64(issuedCommands), nil
}
