//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/filipecosta90/aibench/inference"
	redisai "github.com/filipecosta90/aibench/redisai-go"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
	"strconv"
	math2 "math"

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

func convertSliceStringToFloat(transactionDataString []string) []float32 {
	res := make([]float32, len(transactionDataString))
	for i := range transactionDataString {
		value, _ := strconv.ParseFloat(transactionDataString[i], 64)
		res[i] = float32(value)
	}
	return res
}


func Float32bytes(float float32) []byte {
	bits := math2.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

func (p *Loader) ProcessLoadQuery(q []string) ([]*inference.Stat, error) {

	referenceDataTensorName := "referenceTensor:" + q[0]
	referenceDataKeyBLOBName := "referenceBLOB:" + q[0]
	refData := convertSliceStringToFloat(randReferenceData(256))

	qbytes := Float32bytes(refData[0])
	for _, value := range refData[1:256] {
		qbytes = append( qbytes, Float32bytes(value)... )
	}
	tensorset_args := redisai.Generate_AI_TensorSet_Args(referenceDataTensorName, "FLOAT",  []int{256}, "VALUES", refData)
	errTensorSet := client.Do(tensorset_args...).Err()
	if errTensorSet != nil {
		log.Fatalf("Command TensorSet:%v\n", errTensorSet)
	}

	buffer := &bytes.Buffer{}

	gob.NewEncoder(buffer).Encode(refData)
	errSet := client.Set(referenceDataKeyBLOBName, qbytes, 0).Err()
	if errSet != nil {
		log.Fatalf("Command Set:%v\n", errSet)
	}
	return nil, nil
}
