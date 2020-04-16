//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"flag"
	"fmt"
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/fraud"
	"github.com/RedisAI/aibench/inference"
	"github.com/RedisAI/redisai-go/redisai"
	_ "github.com/lib/pq"
	"github.com/mediocregopher/radix"
	"log"
	"strings"
	"time"

	//ignoring until we get the correct model
	//"log"
	"sync"
)

// Global vars:
var (
	runner         *inference.BenchmarkRunner
	//cpool          *radix.Pool
	host           string
	port 	string
	model          string
	modelFilename  string
	useDag         bool
	showExplain    bool
	clusterMode    bool
	//nodes          []radix.ClusterNode
	//nodesAddresses []string
	//cluster 	*radix.Cluster
	//cpool map[int]*radix.Pool
	cpool *radix.Pool
	PoolPipelineConcurrency int
	PoolPipelineWindow time.Duration
)

var (
	inferenceType = "RedisAI Query - with AI.TENSORSET transacation datatype BLOB"
)


func getClusterNodesFromArgs(nodes []radix.ClusterNode, port string, host string, nodesAddresses []string) ([]radix.ClusterNode, []string) {
	nodes = []radix.ClusterNode{}
	ports := strings.Split(port, ",")
	for idx, nhost := range strings.Split(host, ",") {
		node := radix.ClusterNode{
			Addr:            fmt.Sprintf("%s:%s", nhost, ports[idx]),
			ID:              "",
			Slots:           nil,
			SecondaryOfAddr: "",
			SecondaryOfID:   "",
		}
		nodes = append(nodes, node)
		nodesAddresses = append(nodesAddresses, node.Addr)
	}
	return nodes, nodesAddresses
}

// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()
	flag.StringVar(&host, "host", "localhost", "Redis host address")
	flag.StringVar(&port, "port", "6379", "Redis host port")
	flag.StringVar(&model, "model", "", "model name")
	flag.StringVar(&modelFilename, "model-filename", "", "modelFilename")
	flag.BoolVar(&useDag, "use-dag", false, "use DAGRUN")
	flag.BoolVar(&clusterMode, "cluster-mode", false, "read cluster slots and distribute inferences among shards.")
flag.DurationVar(&PoolPipelineWindow,"pool-pipeline-window", 500*time.Microsecond,"If window is zero then implicit pipelining will be disabled")
	flag.IntVar(&PoolPipelineConcurrency,"pool-pipeline-concurrency", 10,"If limit is zero then no limit will be used and pipelines will only be limited by the specified time window")

	// If limit is zero then no limit will be used and pipelines will only be limited
		// by the specified time window.

	//	PoolPipelineConcurrency(size)
	//	PoolPipelineWindow(150 * time.Microsecond, 0)
	flag.Parse()
	//
	//cpool = &redis.Pool{
	//	MaxIdle:     3,
	//	IdleTimeout: 240 * time.Second,
	//	Dial:        func() (redis.Conn, error) { return redis.DialURL(fmt.Sprintf("%s:%d",host,port)) },
	//}
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
	pclient *redisai.Client
}

func (p *Processor) Close() {
	if p.pclient != nil {
		p.pclient.Close()
	}
}

func newProcessor() inference.Processor { return &Processor{} }

func (p *Processor) Init(numWorker int, wg *sync.WaitGroup, m chan uint64, rs chan uint64) {
	p.opts = &queryExecutorOptions{
		showExplain:   showExplain,
		debug:         runner.DebugLevel() > 0,
		printResponse: runner.DoPrintResponses(),
	}
	p.Wg = wg
	p.Metrics = m
	var err error  = nil
	//if cpool == nil {
	//	cpool = make(map[int]*radix.Pool)
	//}
	cpool, err = radix.NewPool("tcp",fmt.Sprintf("%s:%s",host,port),PoolPipelineConcurrency,radix.PoolPipelineWindow(PoolPipelineWindow,PoolPipelineConcurrency))
			if err != nil {
				log.Fatalf("Error preparing for ModelRun(), while issuing ModelSet. error = %v", err)
			}

	//p.pclient = redisai.Connect(fmt.Sprintf("%s:%d",host,port), cpool)
	//var autoFlushSize uint32 = 2
	//if runner.UseReferenceData() {
	//	autoFlushSize++
	//}

	//if clusterMode {
	//	nodes, nodesAddresses = getClusterNodesFromTopology(host, port, nodes, nodesAddresses)
	//} else {
	//	nodes, nodesAddresses = getClusterNodesFromArgs(nodes, port, host, nodesAddresses)
	//}
	//
	////p.pclient.Pipeline(autoFlushSize

	//if(numWorker==0){
	//	cluster, err := radix.NewCluster([]string{fmt.Sprintf("redis://%s:%s",host,port)})
	//	if err != nil {
	//		log.Fatalf("Error radix.NewCluster. error = %v", err)
	//	}
	//	fmt.Println(cluster.Topo().Map())
	//	//cluster.Do(radix.Cmd(nil,"PING"))
	//
	//	if modelFilename != "" {
	//		data, err := ioutil.ReadFile(modelFilename)
	//		if err != nil {
	//			log.Fatalf("Error preparing for ModelRun(), while issuing ModelSet. error = %v", err)
	//		}
	//
	//		err = p.pclient.ModelSet(model, redisai.BackendTF, redisai.DeviceCPU, data, []string{"transaction", "reference"}, []string{"output"})
	//		if err != nil {
	//			log.Fatalf("Error preparing for ModelRun(), while issuing ModelSet. error = %v", err)
	//		}
	//	}
	//}
}

func (p *Processor) ProcessInferenceQuery(q []byte, isWarm bool, workerNum int, useReferenceData bool) ([]*inference.Stat, error) {

	//// No need to run again for EXPLAIN
	//if isWarm && p.opts.showExplain {
	//	return nil, nil
	//}
	idUint64 := fraud.Uint64frombytes(q[0:8])
	idS := fmt.Sprintf("%d", idUint64)
	referenceDataTensorName := "referenceTensor:{" + idS + "}"
	classificationTensorName := "classificationTensor:{" + idS + "}"
	transactionDataTensorName := "transactionTensor:{" + idS + "}"
	//transactionValues := q[8:128]
	//transactionValuesS := fmt.Sprintf("%x", q[8:128])

	args := []string{"LOAD", "1", referenceDataTensorName, "|>",
		"AI.TENSORSET", transactionDataTensorName, "FLOAT", "1", "30", "BLOB", string(q[8:128]), "|>",
		"AI.MODELRUN", model, "INPUTS", transactionDataTensorName, referenceDataTensorName, "OUTPUTS", classificationTensorName, "|>",
		"AI.TENSORGET", classificationTensorName, "BLOB",
	}
	//args[0]=args[0]
	start := time.Now()
	//cpool[workerNum].Do(radix.Cmd(nil, "AI.DAGRUN", args...))
	err := cpool.Do(radix.Cmd(nil, "AI.DAGRUN", args...))
	if err != nil {
		extendedError := fmt.Errorf("Prediction Receive() failed:%v\n", err)
			log.Fatal(extendedError)
	}
	//
	//p.pclient.ActiveConnNX()
	//var PredictResponse []interface{}
	//var err error = nil

	//if useDag == true {
	//	resp, err := redis.Values(p.pclient.ActiveConn.Do("AI.DAGRUN", args...))
	//	data, err := redisai.ProcessTensorReplyMeta(resp[2], err)
	//	PredictResponse, err = redisai.ProcessTensorReplyBlob(data, err)
	//} else {
	//	p.pclient.TensorSet(transactionDataTensorName, redisai.TypeFloat, []int{1, 30}, transactionValues)
	//	if useReferenceData == true {
	//		p.pclient.ModelRun(model, []string{transactionDataTensorName, referenceDataTensorName}, []string{classificationTensorName})
	//	} else {
	//		p.pclient.ModelRun(model, []string{transactionDataTensorName}, []string{classificationTensorName})
	//	}
	//	p.pclient.TensorGet(classificationTensorName, redisai.TensorContentTypeBlob)
	//	err = p.pclient.Flush()
	//}
	//
	//if err != nil {
	//	extendedError := fmt.Errorf("Prediction Receive() failed:%v\n", err)
	//	if runner.IgnoreErrors() {
	//		fmt.Println(extendedError)
	//	} else {
	//		log.Fatal(extendedError)
	//	}
	//}
	//
	//if useDag == false {
	//	p.pclient.Receive()
	//	p.pclient.Receive()
	//	resp, err := p.pclient.Receive()
	//	data, err := redisai.ProcessTensorReplyMeta(resp, err)
	//	PredictResponse, err = redisai.ProcessTensorReplyBlob(data, err)
	//}
	//
	took := time.Since(start).Microseconds()
	//if err != nil {
	//	extendedError := fmt.Errorf("ProcessTensorReplyBlob() failed:%v\n", err)
	//	if runner.IgnoreErrors() {
	//		fmt.Println(extendedError)
	//	} else {
	//		log.Fatal(extendedError)
	//	}
	//}
	//if p.opts.printResponse {
	//	if err != nil {
	//		extendedError := fmt.Errorf("Response parsing failed:%v\n", err)
	//		if runner.IgnoreErrors() {
	//			fmt.Println(extendedError)
	//		} else {
	//			log.Fatal(extendedError)
	//		}
	//	}
	//	fmt.Println("RESPONSE: ", PredictResponse[2])
	//}
	// VALUES
	stat := inference.GetStat()
	stat.Init([]byte(inferenceType), took, uint64(0), false, "")

	return []*inference.Stat{stat}, nil
}
