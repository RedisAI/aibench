//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/fraud"
	"github.com/RedisAI/aibench/inference"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"sync"
	"time"
)

// Program option vars:
var (
	redisHost             string
	torchserveHost        string
	mysqlHost             string
	torchserveRequestUri  string
	strPost               = []byte("POST")
	strRequestURI         = []byte("")
	strHost               = []byte("")
	showExplain           bool
	runner                *inference.BenchmarkRunner
	redisClient           *redis.Client
	torchserveReadTimeout time.Duration
	mysqlClient           *sql.DB
	mysqlMaxIdle          int
	mysqlMaxOpen          int
	mysqlConnMaxLifetime  time.Duration
)

// Parse args:
func init() {
	runner = inference.NewBenchmarkRunner()
	flag.StringVar(&redisHost, "redis-host", "127.0.0.1:6379", "Redis host address and port")
	flag.StringVar(&torchserveHost, "torchserve-host", "127.0.0.1:8080", "REST API host address and port")
	flag.DurationVar(&torchserveReadTimeout, "torchserve-read-timeout", 5*time.Second, "REST API timeout")
	flag.StringVar(&torchserveRequestUri, "torchserve-request-uri", "/predictions/financial", "torchserve REST API request URI")
	flag.StringVar(&mysqlHost, "mysql-host", "perf:perf@tcp(127.0.0.1:3306)/", "MySql host address and port")
	flag.IntVar(&mysqlMaxIdle, "mysql-max-idle", 256, "MySql max idle")
	flag.IntVar(&mysqlMaxOpen, "mysql-max-open", 512, "MySql max open")
	flag.DurationVar(&mysqlConnMaxLifetime, "mysql-conn-max-lifetime", time.Minute*10, "MySql ConnMaxLifetime")
	flag.Parse()
	if runner.UseReferenceDataRedis() {
		redisClient = redis.NewClient(&redis.Options{
			Addr: redisHost,
		})
	}
	if runner.UseReferenceDataMysql() {
		var err error
		mysqlClient, err = sql.Open("mysql", mysqlHost)
		if err != nil {
			log.Fatalf(fmt.Sprintf("Error connection to MySql %v", err))
		}
		mysqlClient.SetMaxIdleConns(mysqlMaxIdle)
		mysqlClient.SetMaxOpenConns(mysqlMaxOpen)
		mysqlClient.SetConnMaxLifetime(mysqlConnMaxLifetime)
	}
}

func main() {
	strRequestURI = []byte(torchserveRequestUri)
	strHost = []byte(torchserveHost)
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

func (p *Processor) Close() {

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
		Addr:                      torchserveHost,
		ReadTimeout:               torchserveReadTimeout,
		MaxIdleConnDuration:       torchserveReadTimeout,
		MaxIdemponentCallAttempts: 10,
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, torchserveReadTimeout)
		},
	}
}

func (p *Processor) ProcessInferenceQuery(q []byte, isWarm bool, workerNum int, useReferenceDataRedis bool, useReferenceDataMysql bool) ([]*inference.Stat, error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}
	idUint64 := fraud.Uint64frombytes(q[0:8])
	idS := fmt.Sprintf("%d", idUint64)
	transactionValues := q[8:128]
	transactionValuesFloats := fraud.ConvertStringToFloatSlice(transactionValues)
	referenceDataKeyName := "referenceBLOB:{" + idS + "}"
	req := fasthttp.AcquireRequest()
	req.Header.SetMethodBytes(strPost)
	var redisRespReference []byte
	var redisRespReferenceFloats []float32
	//var redisErr error
	var body map[string][]float32

	req.SetRequestURIBytes(strRequestURI)
	req.SetHostBytes(strHost)
	req.Header.SetContentType("application/json")
	res := fasthttp.AcquireResponse()
	start := time.Now()
	if useReferenceDataRedis {
		//redisRespReference, redisErr = redisClient.Get(redisClient.Context(), referenceDataKeyName).Bytes()
		//if redisErr != nil {
		//	log.Fatalln("Error on redisClient.Get", redisErr)
		//}
		redisRespReferenceFloats = fraud.ConvertStringToFloatSlice(redisRespReference)
		body = map[string][]float32{"transaction": transactionValuesFloats, "reference": redisRespReferenceFloats}
	}

	if useReferenceDataMysql {
		statement := mysqlClient.QueryRow("select blobtensor from test.tbltensorblobs where id=?", referenceDataKeyName)
		var mysqlResult []byte
		err := statement.Scan(&mysqlResult)
		if err != nil {
			log.Fatalln("Error on MySqlClient", err)
		}
		redisRespReferenceFloats = fraud.ConvertStringToFloatSlice(redisRespReference)
		body = map[string][]float32{"transaction": transactionValuesFloats, "reference": redisRespReferenceFloats}
	}

	if !useReferenceDataRedis && !useReferenceDataMysql {
		body = map[string][]float32{"transaction": transactionValuesFloats}
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		log.Fatalln(err)
	}

	req.SetBody(bytes.NewBuffer(bodyJSON).Bytes())
	err = p.httpclient.DoTimeout(req, res, torchserveReadTimeout)
	if err != nil {
		fasthttp.ReleaseResponse(res)
		log.Fatalln("Error on httpclient.DoTimeout", err)
	}
	took := time.Since(start).Microseconds()
	if p.opts.printResponse {
		fmt.Printf("REQUEST BODY: %v RESPONSE %v", body, res.String())
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("Wrong status inference response code. expected %v, got %d", 200, res.StatusCode())
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
	stat := inference.GetStat()

	stat.Init([]byte("DL REST API Query"), took, uint64(0), false, "")
	return []*inference.Stat{stat}, nil
}
