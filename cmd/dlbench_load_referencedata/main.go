//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/filipecosta90/dlbench/inference"
	redisai "github.com/filipecosta90/dlbench/redisai-go"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
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

func randReferenceData(n int) []string {
	res := make([]string, n)
	for i := range res {
		res[i] = fmt.Sprintf("%f", float32(rand.Float64()))
	}
	return res
}

func (p *Loader) ProcessLoadQuery(q []string) ([]*inference.Stat, error) {

	referenceDataTensorName := "referenceTensor:" + q[0]
	referenceDataKeyName := "referenceKey:" + q[0]
	//referenceDataListName := "referenceList:" + q[0]
	refData := randReferenceData(256)
	tensorset_args := redisai.Generate_AI_TensorSet_Args(referenceDataTensorName, "FLOAT", []int{256}, refData)
	errTensorSet := client.Do(tensorset_args...).Err()
	if errTensorSet != nil {
		log.Fatalf("Command TensorSet:%v\n", errTensorSet)
	}

	buffer := &bytes.Buffer{}

	gob.NewEncoder(buffer).Encode(refData)
	refDataBytes := buffer.Bytes()
	//errLPush := client.RPush( referenceDataListName, refData ).Err()
	//if errLPush != nil {
	//	log.Fatalf("Command RPush:%v\n", errLPush)
	//}
	errSet := client.Set(referenceDataKeyName, refDataBytes, 0).Err()
	if errSet != nil {
		log.Fatalf("Command Set:%v\n", errSet)
	}
	return nil, nil
}
