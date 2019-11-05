//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/fraud"
	"github.com/RedisAI/aibench/inference"
	"github.com/go-redis/redis"
	googleprotobuf "github.com/golang/protobuf/ptypes/wrappers"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"log"
	"sync"
	tfcoreframework "tensorflow/core/framework"
	tensorflowserving "tensorflow_serving/apis"
	"time"
)

// Program option vars:
var (
	redisHost             string
	tensorflowServingHost string
	model                 string
	version               int

	showExplain bool
)

// Global vars:
var (
	runner *inference.BenchmarkRunner
)

var (
	redisClient *redis.Client
)

// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()
	flag.StringVar(&redisHost, "redis-host", "127.0.0.1:6379", "Redis host address and port")
	flag.StringVar(&tensorflowServingHost, "tensorflow-serving-host", "127.0.0.1:8500", "TensorFlow serving host address and port")
	flag.StringVar(&model, "model", "", "Model name")
	flag.IntVar(&version, "model-version", 1, "Model version")
	flag.Parse()
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisHost,
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
	opts                    *queryExecutorOptions
	Metrics                 chan uint64
	Wg                      *sync.WaitGroup
	predictionServiceClient tensorflowserving.PredictionServiceClient
	grpcClientConn          *grpc.ClientConn
}

func (p *Processor) Close() {
	p.grpcClientConn.Close()
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
	var err error = nil
	p.grpcClientConn, err = grpc.Dial(tensorflowServingHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect to the grpc server: %v\n", err)
	}
	//defer p.grpcClientConn.Close()
	p.predictionServiceClient = tensorflowserving.NewPredictionServiceClient(p.grpcClientConn)
}

func (p *Processor) ProcessInferenceQuery(q []byte, isWarm bool, workerNum int) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}
	// reconnect if the connection was shutdown
	if p.grpcClientConn.GetState() == connectivity.Shutdown {
		var err error = nil
		p.grpcClientConn, err = grpc.Dial(tensorflowServingHost, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Cannot connect to the grpc server: %v\n", err)
		}
		defer p.grpcClientConn.Close()
		p.predictionServiceClient = tensorflowserving.NewPredictionServiceClient(p.grpcClientConn)
	}

	idUint64 := fraud.Uint64frombytes(q[0:8])
	idS := fmt.Sprintf("%d", idUint64)
	transactionValues := q[8:128]

	referenceDataKeyName := "referenceBLOB:{" + idS + "}"

	start := time.Now()
	redisRespReferenceBytes, redisErr := redisClient.Get(referenceDataKeyName).Bytes()
	if redisErr != nil {
		log.Fatalln(redisErr)
	}
	request := &tensorflowserving.PredictRequest{
		ModelSpec: &tensorflowserving.ModelSpec{
			Name: model,
			Version: &googleprotobuf.Int64Value{
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
				TensorContent: transactionValues,
			},
			"reference": {
				Dtype: tfcoreframework.DataType_DT_FLOAT,
				TensorShape: &tfcoreframework.TensorShapeProto{
					Dim: []*tfcoreframework.TensorShapeProto_Dim{
						{
							Size: int64(256),
						},
					},
				},
				TensorContent: redisRespReferenceBytes,
			},
		},
	}
	PredictResponse, err := p.predictionServiceClient.Predict(context.Background(), request)
	took := time.Since(start).Microseconds()
	if err != nil {
		log.Fatalf("Prediction failed:%v\n", err)
	}
	if p.opts.printResponse {
		fmt.Println("RESPONSE: ", PredictResponse)
	}

	stat := inference.GetStat()
	stat.Init([]byte("TensorFlow serving Query"), took, uint64(0), false, "")
	return []*inference.Stat{stat}, nil
}
