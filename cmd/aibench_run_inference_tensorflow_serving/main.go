//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"flag"
	"fmt"

	//"fmt"
	"github.com/filipecosta90/aibench/inference"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	tfcoreframework "tensorflow/core/framework"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"sync"
	tensorflowserving "tensorflow_serving/apis"
	google_protobuf "github.com/golang/protobuf/ptypes/wrappers"

	"time"
)

// Program option vars:
var (
	redis_host              string
	tensorflow_serving_host string
	model                   string
	version int

	showExplain bool
)

// Global vars:
var (
	runner *inference.BenchmarkRunner
)

var (
	redisClient             *redis.Client
	//predictionServiceClient tensorflowserving.PredictionServiceClient
	//grpcClientConn grpc.ClientConn
)

// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()

	flag.StringVar(&redis_host, "redis-host", "127.0.0.1:6379", "Redis host address and port")
	flag.StringVar(&tensorflow_serving_host, "tensorflow-serving-host", "127.0.0.1:8500", "TensorFlow serving host address and port")
	flag.StringVar(&model, "model", "", "Model name")
	flag.IntVar(&version, "model-version", 1, "Model version")


	flag.Parse()
	redisClient = redis.NewClient(&redis.Options{
		Addr: redis_host,
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
	opts    *queryExecutorOptions
	Metrics chan uint64
	Wg      *sync.WaitGroup
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

func convertSliceStringToFloat(transactionDataString []string) []float32 {
	res := make([]float32, len(transactionDataString))
	for i := range transactionDataString {
		value, _ := strconv.ParseFloat(transactionDataString[i], 64)
		res[i] = float32(value)
	}
	return res
}

func (p *Processor) ProcessInferenceQuery(q []string, isWarm bool) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}

	referenceDataKeyName := "referenceKey:" + q[0]
	referenceKeySlice := []float32{}

	grpcClientConn, err := grpc.Dial(tensorflow_serving_host, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect to the grpc server: %v\n", err)
	}
	defer grpcClientConn.Close()

	predictionServiceClient := tensorflowserving.NewPredictionServiceClient(grpcClientConn)


	start := time.Now()
	redisRespBytes, redisErr := redisClient.Get(referenceDataKeyName).Bytes()
	if redisErr != nil {
		log.Fatalln(redisErr)
	}
	//fmt.Println(redisRespBytes)
	reader := bytes.NewReader(redisRespBytes)
	dec := gob.NewDecoder(reader)
	if err := dec.Decode(&referenceKeySlice); err != nil {
		log.Fatalf("Unable to decode %s. Error:%v\n", referenceDataKeyName, err)
	}
	request := &tensorflowserving.PredictRequest{
		ModelSpec: &tensorflowserving.ModelSpec{
			Name: model,
			Version: &google_protobuf.Int64Value{
				Value: int64(version),
			},

		},
		Inputs: map[string]*tfcoreframework.TensorProto{
			"transaction": {
				Dtype: tfcoreframework.DataType_DT_FLOAT,
				TensorShape: &tfcoreframework.TensorShapeProto{
					Dim: []*tfcoreframework.TensorShapeProto_Dim{
						{
							Size: int64(1),
						},
						{
							Size: int64(30),
						},
					},
				},
				FloatVal: convertSliceStringToFloat(q[1:31]),
			},
			"reference" : {
				Dtype: tfcoreframework.DataType_DT_FLOAT,
				TensorShape: &tfcoreframework.TensorShapeProto{
					Dim: []*tfcoreframework.TensorShapeProto_Dim{
						{
							Size: int64(256),
						},
					},
				},
				FloatVal: referenceKeySlice,
			},
		},
	}
	//fmt.Println(request)

	PredictResponse, err := predictionServiceClient.Predict(context.Background(), request)
	if err != nil {
		log.Fatalf("Prediction failed:%v\n", err)
	}
	took := float64(time.Since(start).Nanoseconds()) / 1e6

	if p.opts.printResponse {
		fmt.Println("RESPONSE: ", PredictResponse)
	}

	if err != nil {
		log.Fatalf("Command failed:%v\n", err)
	}
	stat := inference.GetStat()
	stat.Init([]byte("TensorFlow serving Query"), took, uint64(0), false, "")

	return []*inference.Stat{stat}, nil
}
