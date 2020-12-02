//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"context"
	"flag"
	"fmt"
	triton "github.com/RedisAI/aibench/cmd/aibench_run_inference_triton_vision/nvidia_inferenceserver"
	"github.com/RedisAI/aibench/inference"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"log"
	"sync"
	"time"
)

// Global vars:
var (
	runner             *inference.BenchmarkRunner
	host               string
	model              string
	version            string
	showExplain        bool
	inferenceType      = "NVIDIA triton Query - mobilenet_v1_100_224 "
	rowBenchmarkNBytes = 4 * 1 * 224 * 224 * 3 // number of bytes per float * N x H x W x C
	outputSize         = 1001
	grpcClientConn     *grpc.ClientConn
)

// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()
	flag.StringVar(&host, "host", "127.0.0.1:8001", "NVidia triton host address and port")
	flag.StringVar(&model, "model", "mobilenet_v1_100_224_NxHxWxC", "Name of model being served. (Required)")
	flag.StringVar(&version, "model-version", "", "Model version. Default: Latest Version.")
	flag.Parse()
}

func ServerLiveRequest(client triton.GRPCInferenceServiceClient) *triton.ServerLiveResponse {
	// Create context for our request with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverLiveRequest := triton.ServerLiveRequest{}
	// Submit ServerLive request to server
	serverLiveResponse, err := client.ServerLive(ctx, &serverLiveRequest)
	if err != nil {
		log.Fatalf("Couldn't get server live: %v", err)
	}
	return serverLiveResponse
}

func ServerReadyRequest(client triton.GRPCInferenceServiceClient) *triton.ServerReadyResponse {
	// Create context for our request with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverReadyRequest := triton.ServerReadyRequest{}
	// Submit ServerReady request to server
	serverReadyResponse, err := client.ServerReady(ctx, &serverReadyRequest)
	if err != nil {
		log.Fatalf("Couldn't get server ready: %v", err)
	}
	return serverReadyResponse
}

func ModelMetadataRequest(client triton.GRPCInferenceServiceClient, modelName string, modelVersion string) *triton.ModelMetadataResponse {
	// Create context for our request with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create status request for a given model
	modelMetadataRequest := triton.ModelMetadataRequest{
		Name:    modelName,
		Version: modelVersion,
	}
	// Submit modelMetadata request to server
	modelMetadataResponse, err := client.ModelMetadata(ctx, &modelMetadataRequest)
	if err != nil {
		log.Fatalf("Couldn't get server model metadata: %v", err)
	}
	return modelMetadataResponse
}

func ModelInferRequest(client triton.GRPCInferenceServiceClient, rawInput []byte, modelName string, modelVersion string) *triton.ModelInferResponse {
	// Create context for our request with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request input tensors
	inferInputs := []*triton.ModelInferRequest_InferInputTensor{
		{
			Name:     "input",
			Datatype: "FP32",
			Shape:    []int64{1, 224, 224, 3},
			Contents: &triton.InferTensorContents{
				RawContents: rawInput,
			},
		},
	}

	// Create request input output tensors
	inferOutputs := []*triton.ModelInferRequest_InferRequestedOutputTensor{
		{
			Name: "MobilenetV1/Predictions/Reshape_1",
		},
	}

	// Create inference request for specific model/version
	modelInferRequest := triton.ModelInferRequest{
		ModelName:    modelName,
		ModelVersion: modelVersion,
		Inputs:       inferInputs,
		Outputs:      inferOutputs,
	}

	// Submit inference request to server
	modelInferResponse, err := client.ModelInfer(ctx, &modelInferRequest)
	if err != nil {
		log.Fatalf("Error processing InferRequest: %v", err)
	}
	return modelInferResponse
}

// Convert output's raw bytes into int32 data (assumes Little Endian)
func Postprocess(inferResponse *triton.ModelInferResponse) []float32 {
	outputBytes0 := make([]byte, 0, 0)
	if len(inferResponse.Outputs) > 0 {
		outputBytes0 = inferResponse.Outputs[0].Contents.RawContents
	}
	return inference.ConvertByteSliceToFloatSlice(outputBytes0)
}

func main() {
	runner.Run(&inference.RedisAIPool, newProcessor, rowBenchmarkNBytes, 1)
}

type queryExecutorOptions struct {
	showExplain   bool
	debug         bool
	printResponse bool
}

type Processor struct {
	opts           *queryExecutorOptions
	Metrics        chan uint64
	Wg             *sync.WaitGroup
	pclient        triton.GRPCInferenceServiceClient
	grpcClientConn *grpc.ClientConn
}

func (p *Processor) Close() {
	p.grpcClientConn.Close()

}

func (p *Processor) CollectRunTimeMetrics() (ts int64, stats interface{}, err error) {
	// TODO:
	return
}

func newProcessor() inference.Processor { return &Processor{} }

func (p *Processor) Init(numWorker int, totalWorkers int, wg *sync.WaitGroup, m chan uint64, rs chan uint64) {
	p.opts = &queryExecutorOptions{
		showExplain:   showExplain,
		debug:         runner.DebugLevel() > 0,
		printResponse: runner.DoPrintResponses(),
	}
	p.Wg = wg
	p.Metrics = m
	var err error

	// Connect to gRPC server
	p.grpcClientConn, err = grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldn't connect to endpoint %s: %v", host, err)
	}

	// Create client from gRPC server connection
	p.pclient = triton.NewGRPCInferenceServiceClient(p.grpcClientConn)

	serverLiveResponse := ServerLiveRequest(p.pclient)
	fmt.Printf("triton Health - Live: %v\n", serverLiveResponse.Live)

	serverReadyResponse := ServerReadyRequest(p.pclient)
	fmt.Printf("triton Health - Ready: %v\n", serverReadyResponse.Ready)

	modelMetadataResponse := ModelMetadataRequest(p.pclient, model, "")
	fmt.Println(modelMetadataResponse)
}

func (p *Processor) ProcessInferenceQuery(q []byte, isWarm bool, workerNum int, useReferenceDataRedis bool, useReferenceDataMysql bool, queryNumber int64) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}
	tensorValues := q
	start := time.Now()
	inferResponse := ModelInferRequest(p.pclient, tensorValues, model, version)
	took := time.Since(start).Microseconds()
	if p.opts.printResponse {
		fmt.Println("RAW RESPONSE: ", inferResponse)
		fmt.Println("RESPONSE: ", Postprocess(inferResponse))
	}

	stat := inference.GetStat()
	stat.Init([]byte(inferenceType), took, uint64(0), false, "")

	return []*inference.Stat{stat}, nil
}
