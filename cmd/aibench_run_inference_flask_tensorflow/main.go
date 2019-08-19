//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/filipecosta90/aibench/inference"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Program option vars:
var (
	redis_host   string
	restapi_host string
	restapi_request_uri string
	model        string
	version      int
 strPost = []byte("POST")
	strContentType = []byte("application/json")

	strRequestURI = []byte("")
	strHost = []byte("")
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
	flag.StringVar(&redis_host, "redis-host", "127.0.0.1:6379", "Redis host address and port")
	flag.StringVar(&restapi_host, "restapi-host", "127.0.0.1:8000", "REST API host address and port")
	flag.StringVar(&restapi_request_uri, "restapi-request-uri", "/predict", "REST API request URI")
	flag.Parse()
	redisClient = redis.NewClient(&redis.Options{
		Addr: redis_host,
	})
}

func main() {
	strRequestURI = []byte(restapi_request_uri)
	strHost = []byte(restapi_host)

	runner.Run(&inference.RedisAIPool, newProcessor)
}

type queryExecutorOptions struct {
	showExplain   bool
	debug         bool
	printResponse bool
}

type Processor struct {
	opts       *queryExecutorOptions
	Metrics    chan uint64
	Wg         *sync.WaitGroup
	httpclient *fasthttp.HostClient
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
	p.httpclient = &fasthttp.HostClient{
		Addr: restapi_host,
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

func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}

func (p *Processor) ProcessInferenceQuery(q []string, isWarm bool) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}

	referenceDataKeyName := "referenceBLOB:" + q[0]
	req := fasthttp.AcquireRequest()
	req.Header.SetMethodBytes(strPost)
	req.Header.SetContentTypeBytes(strContentType)
	req.SetRequestURIBytes(strRequestURI)
	req.SetHostBytes(strHost)
	res := fasthttp.AcquireResponse()
	transaction_string := strings.Join(q[1:31], ",")

	start := time.Now()
	redisRespReferenceBytes, redisErr := redisClient.Get(referenceDataKeyName).Bytes()
	if redisErr != nil {
		log.Fatalln(redisErr)
	}
	referenceFloats := []string{}
	for i := 0; i < 256; i++ {
		value := Float32frombytes(redisRespReferenceBytes[4*i:4*(i+1)])
		referenceFloats = append(referenceFloats,fmt.Sprintf("%f", value))
	}
	bodyJSON := []byte(fmt.Sprintf(`{"inputs":{"transaction":[[%s]],"reference":[%s]}}`,transaction_string, strings.Join(referenceFloats, ",") ))
	req.SetBody(bodyJSON)
	if err := p.httpclient.Do(req, res); err != nil {
		fasthttp.ReleaseResponse(res)
		log.Fatalln(err)
		}
	took := float64(time.Since(start).Nanoseconds()) / 1e6
	fasthttp.ReleaseRequest(req)
	if p.opts.printResponse {
		body := res.Body()
		fmt.Println("RESPONSE: ", string(body))
	}
	fasthttp.ReleaseResponse(res)
	stat := inference.GetStat()
	stat.Init([]byte("DL REST API Query"), took, uint64(0), false, "")
	return []*inference.Stat{stat}, nil
}
