//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/RedisAI/aibench/inference"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
	"log"
	"mime/multipart"
	"net"
	"sync"
	"time"
)

// Program option vars:
var (
	redisHost          string
	restapiHost        string
	restapiRequestUri  string
	strPost            = []byte("POST")
	strRequestURI      = []byte("")
	strHost            = []byte("")
	showExplain        bool
	runner             *inference.BenchmarkRunner
	redisClient        *redis.Client
	restapiReadTimeout time.Duration
	rowBenchmarkNBytes = 8 + 120 + 1024
)

// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()
	flag.StringVar(&redisHost, "redis-host", "127.0.0.1:6379", "Redis host address and port")
	flag.StringVar(&restapiHost, "restapi-host", "127.0.0.1:8000", "REST API host address and port")
	flag.DurationVar(&restapiReadTimeout, "restapi-read-timeout", 5*time.Second, "REST API timeout")
	flag.StringVar(&restapiRequestUri, "restapi-request-uri", "/v2/predict", "REST API request URI")
	flag.Parse()
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

}

func main() {
	strRequestURI = []byte(restapiRequestUri)
	strHost = []byte(restapiHost)
	runner.Run(&inference.RedisAIPool, newProcessor, rowBenchmarkNBytes, 1)
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

func (p *Processor) Close() {

}

func (p *Processor) CollectRunTimeMetrics() (ts int64, stats interface{}, err error) {
	// TODO:
	return
}

func newProcessor() inference.Processor { return &Processor{} }

func (p *Processor) Init(numWorker int, totalWorkers int, wg *sync.WaitGroup, m chan uint64, rs chan uint64) {
	p.Wg = wg
	p.Metrics = m
	p.opts = &queryExecutorOptions{
		showExplain:   showExplain,
		debug:         runner.DebugLevel() > 0,
		printResponse: runner.DoPrintResponses(),
	}

	p.httpclient = &fasthttp.HostClient{
		Addr:                      restapiHost,
		ReadTimeout:               restapiReadTimeout,
		MaxIdleConnDuration:       restapiReadTimeout,
		MaxIdemponentCallAttempts: 10,
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, restapiReadTimeout)
		},
	}

}

func (p *Processor) ProcessInferenceQuery(q []byte, isWarm bool, workerNum int, useReferenceDataRedis bool, useReferenceDataMysql bool, queryNumber int64) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}
	idUint64 := inference.Uint64frombytes(q[0:8])
	idS := fmt.Sprintf("%d", idUint64)
	transactionValues := q[8:128]
	referenceDataKeyName := "referenceBLOB:{" + idS + "}"
	req := fasthttp.AcquireRequest()
	req.Header.SetMethodBytes(strPost)

	req.SetRequestURIBytes(strRequestURI)
	req.SetHostBytes(strHost)
	res := fasthttp.AcquireResponse()
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	transPart, err := writer.CreateFormFile("transaction", "transaction")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = transPart.Write(transactionValues)
	if err != nil {
		log.Fatalln(err)
	}
	start := time.Now()
	if useReferenceDataRedis {
		redisRespReferenceBytes, redisErr := redisClient.Get(redisClient.Context(), referenceDataKeyName).Bytes()
		if redisErr != nil {
			log.Fatalln("Error on redisClient.Get", redisErr)
		}
		refPart, err := writer.CreateFormFile("reference", "reference")
		if err != nil {
			log.Fatalln(err)
		}
		_, err = refPart.Write(redisRespReferenceBytes)
		if err != nil {
			log.Fatalln(err)
		}
	}

	writer.Close()
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.SetBody(body.Bytes())
	err = p.httpclient.DoTimeout(req, res, restapiReadTimeout)
	if err != nil {
		fasthttp.ReleaseResponse(res)
		log.Fatalln("Error on httpclient.DoTimeout", err)
	}
	took := time.Since(start).Microseconds()
	fasthttp.ReleaseRequest(req)
	if res.StatusCode() != 200 {
		log.Fatalln(fmt.Sprintf("Wrong status inference response code. expected %v, got %d", 200, res.StatusCode()))
	}
	if p.opts.printResponse {
		body := res.Body()
		fmt.Println("RESPONSE: ", string(body))
	}
	fasthttp.ReleaseResponse(res)
	stat := inference.GetStat()

	stat.Init([]byte("DL REST API Query"), took, uint64(0), false, "")
	return []*inference.Stat{stat}, nil
}
