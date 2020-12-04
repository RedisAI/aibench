package inference

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/HdrHistogram/hdrhistogram-go"
	"golang.org/x/time/rate"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	labelAllQueries = "All queries"
	defaultReadSize = 4 << 20 // 4 MB
	Inf             = rate.Limit(math.MaxFloat64)
)

// LoadRunner contains the common components for running a inference benchmarking
// program against a database.
type BenchmarkRunner struct {
	// flag fields
	dbName                             string
	limit                              uint64
	limitrps                           uint64
	memProfile                         string
	cpuProfile                         string
	workers                            uint
	repetitions                        uint
	printResponses                     bool
	ignoreErrors                       bool
	debug                              int
	enableReferenceDataRedis           bool
	fileName                           string
	seed                               int64
	reportingPeriod                    time.Duration
	outputFileStatsResponseLatencyHist string

	// non-flag fields
	br             *bufio.Reader
	sp             *statProcessor
	scanner        *producer
	ch             chan []byte
	inferenceCount uint64

	testResult           TestResult
	JsonOutFile          string
	MetadataAutobatching int64
}

// NewLoadRunner creates a new instance of LoadRunner which is
// common functionality to be used by inference benchmarker programs
func NewBenchmarkRunner() *BenchmarkRunner {
	runner := &BenchmarkRunner{}
	runner.scanner = newScanner(&runner.limit)
	runner.sp = &statProcessor{
		limit: &runner.limit,
	}
	flag.Uint64Var(&runner.sp.burnIn, "burn-in", 0, "Number of queries to ignore before collecting statistics.")
	flag.Uint64Var(&runner.limit, "max-queries", 0, "Limit the number of queries to send, 0 = no limit")
	flag.StringVar(&runner.memProfile, "memprofile", "", "Write a memory profile to this file.")
	flag.StringVar(&runner.cpuProfile, "cpuprofile", "", "Write a cpu profile to this file.")
	flag.Uint64Var(&runner.limitrps, "limit-rps", 0, "Limit overall RPS. 0 disables limit.")
	flag.UintVar(&runner.workers, "workers", 8, "Number of concurrent requests to make.")
	flag.BoolVar(&runner.printResponses, "print-responses", false, "Pretty print response bodies for correctness checking (default false).")
	flag.BoolVar(&runner.ignoreErrors, "ignore-errors", false, "Whether to ignore the inference errors and continue. By default on error the benchmark stops (default false).")
	flag.BoolVar(&runner.enableReferenceDataRedis, "enable-reference-data-redis", false, "Whether to enable benchmarking inference with a model with reference data on Redis or not (default false).")
	flag.IntVar(&runner.debug, "debug", 0, "Whether to print debug messages.")
	flag.Int64Var(&runner.seed, "seed", 0, "PRNG seed (default, or 0, uses the current timestamp).")
	flag.StringVar(&runner.fileName, "file", "", "File name to read queries from")
	flag.DurationVar(&runner.reportingPeriod, "reporting-period", 1*time.Second, "Period to report write stats")
	flag.StringVar(&runner.JsonOutFile, "json-out-file", "", "Name of json output file to output benchmark results. If not set, will not print to json.")
	flag.Int64Var(&runner.MetadataAutobatching, "metadata-autobatching", -1, "Metadata string containing autobatching on the server side info.")
	flag.StringVar(&runner.outputFileStatsResponseLatencyHist, "output-file-stats-hdr-response-latency-hist", "", "File name to output the hdr response latency histogram to")

	return runner
}

// SetLimit changes the number of queries to run, with 0 being all of them
func (b *BenchmarkRunner) SetLimit(limit uint64) {
	b.limit = limit
}

// DoPrintResponses indicates whether responses for queries should be printed
func (b *BenchmarkRunner) DoPrintResponses() bool {
	return b.printResponses
}

// DebugLevel returns the level of debug messages for this benchmark
func (b *BenchmarkRunner) DebugLevel() int {
	return b.debug
}

// ModelName returns the name of the database to run queries against
func (b *BenchmarkRunner) DatabaseName() string {
	return b.dbName
}

func (b *BenchmarkRunner) IgnoreErrors() bool {
	return b.ignoreErrors
}

func (b *BenchmarkRunner) UseReferenceDataRedis() bool {
	return b.enableReferenceDataRedis
}

// LoaderCreate is a function that creates a new Loader (called in Run)
type ProcessorCreate func() Processor

type MetricCollectorCreate func() MetricCollector

// MetricCollector is an interface that handles the metrics collection from the model server
type MetricCollector interface {
	// CollectRunTimeMetrics asks the specific runner to fetch runtime stats that will then be stored on the results file.
	// Returns the collection timestamp and an interface with all fetched data
	CollectRunTimeMetrics() (int64, interface{}, error)
}

// Processor is an interface that handles the setup of a inference processing worker and executes queries one at a time
type Processor interface {
	// Init initializes at global state for the Loader, possibly based on its worker number / ID
	Init(workerNum int, totalWorkers int, wg *sync.WaitGroup, m chan uint64, rs chan uint64)

	// ProcessInferenceQuery handles a given inference and reports its stats
	ProcessInferenceQuery(q []byte, isWarm bool, workerNum int, useReferenceDataRedis bool, useReferenceDataMysql bool, queryNumber int64) ([]*Stat, error)

	// Close forces any work buffered to be sent to the DB being tested prior to going further
	Close()
}

// GetBufferedReader returns the buffered Reader that should be used by the loader
func (b *BenchmarkRunner) GetBufferedReader() *bufio.Reader {
	if b.br == nil {
		if len(b.fileName) > 0 {
			// Read from specified file
			file, err := os.Open(b.fileName)
			if err != nil {
				panic(fmt.Sprintf("cannot open file for read %s: %v", b.fileName, err))
			}
			b.br = bufio.NewReader(file)
			//bufio.
		} else {
			// Read from STDIN
			b.br = bufio.NewReader(os.Stdin)
		}
	}
	return b.br
}

// Run does the bulk of the benchmark execution.
// It launches a gorountine to track stats, creates workers to process queries,
// read in the input, execute the queries, and then does cleanup.
func (b *BenchmarkRunner) Run(queryPool *sync.Pool, processorCreateFn ProcessorCreate, rowSizeBytes int, inferencesPerRow int64, metricCollectorFn MetricCollectorCreate) {

	if b.cpuProfile != "" {
		fmt.Printf("starting cpu profile. Saving into :%s", b.cpuProfile)
		f, err := os.Create(b.cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

	rand.Seed(b.seed)

	if b.workers == 0 {
		panic("must have at least one worker")
	}
	if b.sp.burnIn > b.limit && b.limit > 0 {
		panic("burn-in is larger than limit")
	}
	b.ch = make(chan []byte, b.workers)

	// Launch the stats processor:
	go b.sp.process(b.workers, true)

	// Launch inference processors

	var requestRate = Inf
	var requestBurst = 1
	if b.limitrps != 0 {
		requestRate = rate.Limit(b.limitrps)
		requestBurst = 1 //int(b.workers)
	}
	var rateLimiter = rate.NewLimiter(requestRate, requestBurst)

	var wg sync.WaitGroup
	for i := 0; i < int(b.workers); i++ {
		wg.Add(1)
		go b.processorHandler(rateLimiter, &wg, queryPool, processorCreateFn(), i, inferencesPerRow, b.limitrps != 0)
	}
	b.testResult.ServerRunTimeStats = make(map[int64]interface{})
	b.testResult.ClientRunTimeStats = make(map[int64]interface{})

	// Read in jobs, closing the job channel when done:
	// Wall clock start time
	wallStart := time.Now()

	// Start background reporting process
	if b.reportingPeriod.Nanoseconds() > 0 {
		go b.report(b.reportingPeriod, wallStart, b.testResult.ClientRunTimeStats)
	}

	if metricCollectorFn != nil {
		go b.collectRunTimeStats(b.reportingPeriod, metricCollectorFn(), b.testResult.ServerRunTimeStats)
	}

	br := b.scanner.setReader(b.GetBufferedReader())
	totalRows := br.produce(queryPool, b.ch, rowSizeBytes, inferencesPerRow, b.debug)
	_, err := fmt.Printf("Read a total of :%d rows\n", totalRows)

	close(b.ch)

	// Block for workers to finish sending requests, closing the stats channel when done:
	wg.Wait()
	b.sp.CloseAndWait()

	// Wall clock end time
	wallEnd := time.Now()
	wallTook := wallEnd.Sub(wallStart)

	b.testResult.StartTime = wallStart.Unix()
	b.testResult.EndTime = wallEnd.Unix()
	b.testResult.DurationMillis = wallTook.Milliseconds()
	b.testResult.TensorBatchSize = uint64(inferencesPerRow)
	b.testResult.MetadataAutobatching = b.MetadataAutobatching
	b.testResult.OverallRates = b.GetOverallRatesMap(b.limit, wallTook)
	b.testResult.OverallQuantiles = b.GetOverallQuantiles(b.sp.StatsMapping[labelAllQueries].latencyHDRHistogram)
	b.testResult.Limit = b.limit
	b.testResult.Workers = b.workers
	b.testResult.MaxRps = b.limitrps

	if strings.Compare(b.JsonOutFile, "") != 0 {
		_, _ = fmt.Printf("Saving JSON results to %s\n", b.JsonOutFile)
		file, err := json.MarshalIndent(b.testResult, "", " ")
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(b.JsonOutFile, file, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = fmt.Printf("Took: %8.3f sec\n", float64(wallTook.Nanoseconds())/1e9)
	if err != nil {
		log.Fatal(err)
	}

	if len(b.outputFileStatsResponseLatencyHist) > 0 {
		_, _ = fmt.Printf("Saving Query Latencies HDR Histogram to %s\n", b.outputFileStatsResponseLatencyHist)

		d1 := []byte(b.sp.StatsMapping[labelAllQueries].stringQueryLatencyFullHistogram())
		err = ioutil.WriteFile(b.outputFileStatsResponseLatencyHist, d1, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	// (Optional) create a memory profile:
	if len(b.memProfile) > 0 {
		f, err := os.Create(b.memProfile)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.WriteHeapProfile(f)
		_ = f.Close()
	}

}

func calculateRateMetrics(current, prev int64, took time.Duration) (rate float64) {
	rate = float64(current-prev) / float64(took.Seconds())
	return
}

func (l *BenchmarkRunner) GetOverallRatesMap(totalOps uint64, took time.Duration) map[string]interface{} {
	/////////
	// Overall Rates
	/////////
	configs := map[string]interface{}{}
	overallOpsRate := calculateRateMetrics(int64(totalOps), 0, took)
	configs["overallOpsRate"] = overallOpsRate
	return configs
}

func generateQuantileMap(hist *hdrhistogram.Histogram) (int64, map[string]float64) {
	ops := hist.TotalCount()
	q0 := 0.0
	q50 := 0.0
	q95 := 0.0
	q99 := 0.0
	q999 := 0.0
	q100 := 0.0
	if ops > 0 {
		q0 = float64(hist.ValueAtQuantile(0.0)) / 10e2
		q50 = float64(hist.ValueAtQuantile(50.0)) / 10e2
		q95 = float64(hist.ValueAtQuantile(95.0)) / 10e2
		q99 = float64(hist.ValueAtQuantile(99.0)) / 10e2
		q999 = float64(hist.ValueAtQuantile(99.90)) / 10e2
		q100 = float64(hist.ValueAtQuantile(100.0)) / 10e2
	}

	mp := map[string]float64{"q0": q0, "q50": q50, "q95": q95, "q99": q99, "q999": q999, "q100": q100}
	return ops, mp
}

func (b *BenchmarkRunner) GetOverallQuantiles(histogram *hdrhistogram.Histogram) map[string]interface{} {
	configs := map[string]interface{}{}
	_, all := generateQuantileMap(histogram)
	configs["AllQueries"] = all
	configs["EncodedHistogram"] = nil
	encodedHist, err := histogram.Encode(hdrhistogram.V2CompressedEncodingCookieBase)
	if err == nil {
		configs["EncodedHistogram"] = encodedHist
	}
	return configs
}

func (b *BenchmarkRunner) processorHandler(rateLimiter *rate.Limiter, wg *sync.WaitGroup, queryPool *sync.Pool, processor Processor, workerNum int, inferencesPerRow int64, limitRps bool) {
	buflen := uint64(len(b.ch))
	metricsChan := make(chan uint64, buflen)
	pwg := &sync.WaitGroup{}
	responseSizesChan := make(chan uint64, buflen)
	pwg.Add(1)
	var workerInferences int64 = 0

	processor.Init(workerNum, int(b.workers), pwg, metricsChan, responseSizesChan)

	for query := range b.ch {
		if limitRps {
			r := rateLimiter.ReserveN(time.Now(), int(inferencesPerRow))
			time.Sleep(r.Delay())
		}
		stats, err := processor.ProcessInferenceQuery(query, false, workerNum, b.enableReferenceDataRedis, false, workerInferences)
		if err != nil {
			if b.IgnoreErrors() {
				fmt.Printf("Ignoring inference error: %v\n", err)
			} else {
				panic(err)
			}
		} else {
			workerInferences++
			totalStatInferences := stats[0].totalResults
			workerInferences = workerInferences + int64(totalStatInferences)
			atomic.AddUint64(&b.inferenceCount, totalStatInferences)
			b.sp.sendStats(stats)
		}
		queryPool.Put(&query)
	}

	processor.Close()

	//pwg.Wait()
	close(metricsChan)
	close(responseSizesChan)

	wg.Done()
}

// report handles periodic reporting of loading stats
func (b *BenchmarkRunner) report(period time.Duration, start time.Time, quantileStats map[int64]interface{}) {
	prevTime := start
	fmt.Printf("%26s %25s %25s %26s %26s %26s\n", "Test time", "Inference Rate", "Total Inferences", "p50 lat. (msec)", "p95 lat. (msec)", "p99 lat. (msec)")
	for now := range time.NewTicker(period).C {
		opsCount := atomic.LoadUint64(&b.inferenceCount)
		took := now.Sub(prevTime)
		statHist := b.sp.InstantaneousStats.latencyHDRHistogram
		testTime := time.Since(start).Seconds()
		instantRate := float64(statHist.TotalCount()) / float64(took.Seconds())
		p50 := float64(statHist.ValueAtQuantile(50.0)) / 10e2
		p95 := float64(statHist.ValueAtQuantile(95.0)) / 10e2
		p99 := float64(statHist.ValueAtQuantile(99.0)) / 10e2
		fmt.Printf("%25.0fs %25.0f %25d %25.3f %25.3f %25.3f\t", testTime, instantRate, opsCount, p50, p95, p99)
		fmt.Printf("\n")
		prevTime = now

		var currentClientStats = make(map[string]interface{})
		currentClientStats["InferenceRate"] = instantRate
		currentClientStats["TestTime"] = testTime
		_, qm := generateQuantileMap(statHist)
		encodedHist, err := statHist.Encode(hdrhistogram.V2CompressedEncodingCookieBase)
		if err == nil {
			currentClientStats["EncodedHistogram"] = encodedHist
		}
		currentClientStats["Quantiles"] = qm
		quantileStats[now.UnixNano()] = currentClientStats
		b.sp.InstantaneousStats.reset()
	}
}

// report handles periodic reporting of loading stats
func (b *BenchmarkRunner) collectRunTimeStats(period time.Duration, collector MetricCollector, runtimeStats map[int64]interface{}) {

	for now := range time.NewTicker(period).C {
		_, metrics, err := collector.CollectRunTimeMetrics()
		if err != nil {
			if b.IgnoreErrors() {
				fmt.Printf("Ignoring runtime stats error: %v\n", err)
			} else {
				log.Fatalf("Runtime stats error: %v\n", err)
			}
		}
		runtimeStats[now.UnixNano()] = metrics
	}
}
