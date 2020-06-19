//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/RedisAI/aibench/inference"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	triton "nvidia_inferenceserver"

	//ignoring until we get the correct model
	//"log"
	"sync"
)

// Global vars:
var (
	runner                  *inference.BenchmarkRunner
	host                    string
	model                   string
	version					int
	showExplain             bool
	inferenceType           = "NVIDIA Triton Query - mobilenet_v1_100_224 "
	rowBenchmarkNBytes      = 4 * 1 * 224 * 224 * 3 // number of bytes per float * N x H x W x C
	grpcClientConn          *grpc.ClientConn
)



// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()
	flag.StringVar(&host, "host", "127.0.0.1:8500", "NVidia Triton host address and port")
	flag.StringVar(&model, "model", "mobilenet_v1_100_224_NxHxWxC", "model name")
	flag.IntVar(&version, "model-version", 1, "Model version")
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
		Name: modelName,
		Version: modelVersion,
	}
	// Submit modelMetadata request to server
	modelMetadataResponse, err := client.ModelMetadata(ctx, &modelMetadataRequest)
	if err != nil {
		log.Fatalf("Couldn't get server model metadata: %v", err)
	}
	return modelMetadataResponse
}

func ModelInferRequest(client triton.GRPCInferenceServiceClient, rawInput [][]byte, modelName string, modelVersion string) *triton.ModelInferResponse {
	// Create context for our request with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request input tensors
	inferInputs := []*triton.ModelInferRequest_InferInputTensor {
		&triton.ModelInferRequest_InferInputTensor{
			Name: "INPUT0",
			Datatype: "INT32",
			Shape: []int64{1, 16},
			Contents: &triton.InferTensorContents{
				RawContents: rawInput[0],
			},
		},
	}

	// Create request input output tensors
	inferOutputs := []*triton.ModelInferRequest_InferRequestedOutputTensor {
		&triton.ModelInferRequest_InferRequestedOutputTensor{
			Name: "OUTPUT0",
		},
	}

	// Create inference request for specific model/version
	modelInferRequest := triton.ModelInferRequest{
		ModelName:    modelName,
		ModelVersion: modelVersion,
		Inputs:    	  inferInputs,
		Outputs:      inferOutputs,
	}

	// Submit inference request to server
	modelInferResponse, err := client.ModelInfer(ctx, &modelInferRequest)
	if err != nil {
		log.Fatalf("Error processing InferRequest: %v", err)
	}
	return modelInferResponse
}
//
//// Convert int32 input data into raw bytes (assumes Little Endian)
//func Preprocess(inputs [][]int32) [][]byte {
//	inputData0 := inputs[0]
//	inputData1 := inputs[1]
//
//	var inputBytes0 []byte
//	var inputBytes1 []byte
//	// Temp variable to hold our converted int32 -> []byte
//	bs := make([]byte, 4)
//	for i := 0; i < inputSize; i++ {
//		binary.LittleEndian.PutUint32(bs, uint32(inputData0[i]))
//		inputBytes0 = append(inputBytes0, bs...)
//		binary.LittleEndian.PutUint32(bs, uint32(inputData1[i]))
//		inputBytes1 = append(inputBytes1, bs...)
//	}
//
//	return [][]byte{inputBytes0, inputBytes1}
//}
//
//// Convert slice of 4 bytes to int32 (assumes Little Endian)
//func readInt32(fourBytes []byte) int32 {
//	buf := bytes.NewBuffer(fourBytes)
//	var retval int32
//	binary.Read(buf, binary.LittleEndian, &retval)
//	return retval
//}
//
//// Convert output's raw bytes into int32 data (assumes Little Endian)
//func Postprocess(inferResponse *triton.ModelInferResponse) [][]int32 {
//	var outputs []*triton.ModelInferResponse_InferOutputTensor
//	outputs = inferResponse.Outputs
//	output0 := outputs[0]
//	output1 := outputs[1]
//	outputBytes0 := output0.Contents.RawContents
//	outputBytes1 := output1.Contents.RawContents
//
//	outputData0 := make([]int32, outputSize)
//	outputData1 := make([]int32, outputSize)
//	for i := 0; i < outputSize; i++ {
//		outputData0[i] = readInt32(outputBytes0[i*4 : i*4+4])
//		outputData1[i] = readInt32(outputBytes1[i*4 : i*4+4])
//	}
//	return [][]int32{outputData0, outputData1}
//}


func main() {
	runner.Run(&inference.RedisAIPool, newProcessor, rowBenchmarkNBytes)
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
	pclient triton.GRPCInferenceServiceClient
	grpcClientConn          *grpc.ClientConn

}

func (p *Processor) Close() {
	p.grpcClientConn.Close()

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
	fmt.Printf("Triton Health - Live: %v\n", serverLiveResponse.Live)

	serverReadyResponse := ServerReadyRequest(p.pclient)
	fmt.Printf("Triton Health - Ready: %v\n", serverReadyResponse.Ready)

	modelMetadataResponse := ModelMetadataRequest(p.pclient, model, "")
	fmt.Println(modelMetadataResponse)

}

func (p *Processor) ProcessInferenceQuery(q []byte, isWarm bool, workerNum int, useReferenceDataRedis bool, useReferenceDataMysql bool, queryNumber int64) ([]*inference.Stat, error) {

	//// No need to run again for EXPLAIN
	//if isWarm && p.opts.showExplain {
	//	return nil, nil
	//}
	//tensorName := fmt.Sprintf("imageTensor:{w%d-i%d}", workerNum, queryNumber)
	//outputTensorName := fmt.Sprintf("classificationTensor:{w%d-i%d}", workerNum, queryNumber)
	//tensorValues := q
	//var args []string
	//if persistOutputs {
	//	args = []string{"PERSIST", "1", outputTensorName, "|>"}
	//
	//} else {
	//	args = []string{"|>"}
	//}
	//args = append(args, //                                           N x H  x W  x C
	//	//"AI.TENSORSET" "000000019042.jpg" "UINT8" "1" "224" "224" "3" "BLOB" ...
	//	"AI.TENSORSET", tensorName, "FLOAT", "1", "224", "224", "3", "BLOB", string(tensorValues), "|>",
	//	"AI.MODELRUN", model, "INPUTS", tensorName, "OUTPUTS", outputTensorName, "|>",
	//	"AI.TENSORGET", outputTensorName, "BLOB")
	//
	//pos := rand.Int31n(int32(len(p.pclient)))
	//start := time.Now()
	//
	//err := p.pclient[pos].Do(radix.Cmd(nil, "AI.DAGRUN", args...))
	//if err != nil {
	//	extendedError := fmt.Errorf("Prediction Receive() failed:%v\n", err)
	//	log.Fatal(extendedError)
	//}
	//
	//took := time.Since(start).Microseconds()

	stat := inference.GetStat()
	stat.Init([]byte(inferenceType), 0, uint64(0), false, "")

	return []*inference.Stat{stat}, nil
}
