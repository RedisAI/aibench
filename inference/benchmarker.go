package inference

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bradfitz/iter"
	"golang.org/x/time/rate"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
)

const (
	labelAllQueries    = "All queries"
	labelColdQueries   = "Cold queries"
	labelWarmQueries   = "Warm queries"
	rowBenchmarkNBytes = 8 + 120 + 1024
	defaultReadSize    = 4 << 20 // 4 MB
	Inf                = rate.Limit(math.MaxFloat64)
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
	enableReferenceData                bool
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
	opsCount       uint64
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
	flag.Uint64Var(&runner.sp.printInterval, "print-interval", 100, "Print timing stats to stderr after this many queries (0 to disable)")
	flag.StringVar(&runner.memProfile, "memprofile", "", "Write a memory profile to this file.")
	flag.StringVar(&runner.cpuProfile, "cpuprofile", "", "Write a cpu profile to this file.")
	flag.Uint64Var(&runner.limitrps, "limit-rps", 0, "Limit overall RPS. 0 disables limit.")
	flag.UintVar(&runner.workers, "workers", 1, "Number of concurrent requests to make.")
	flag.UintVar(&runner.repetitions, "repetitions", 10, "Number of repetitions of requests per dataset ( will round robin ).")
	flag.BoolVar(&runner.sp.prewarmQueries, "prewarm-queries", false, "Run each inference twice in a row so the warm inference is guaranteed to be a cache hit")
	flag.BoolVar(&runner.printResponses, "print-responses", false, "Pretty print response bodies for correctness checking (default false).")
	flag.BoolVar(&runner.ignoreErrors, "ignore-errors", false, "Whether to ignore the inference errors and continue. By default on error the benchmark stops (default false).")
	flag.BoolVar(&runner.enableReferenceData, "enable-reference-data", true, "Whether to enable benchmarking inference with a model with reference data or not (default true).")
	flag.IntVar(&runner.debug, "debug", 0, "Whether to print debug messages.")
	flag.Int64Var(&runner.seed, "seed", 0, "PRNG seed (default, or 0, uses the current timestamp).")
	flag.StringVar(&runner.fileName, "file", "", "File name to read queries from")
	flag.DurationVar(&runner.reportingPeriod, "reporting-period", 1*time.Second, "Period to report write stats")
	flag.StringVar(&runner.outputFileStatsResponseLatencyHist, "output-file-stats-hdr-response-latency-hist", "stats-response-latency-hist.txt", "File name to output the hdr response latency histogram to")

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

func (b *BenchmarkRunner) UseReferenceData() bool {
	return b.enableReferenceData
}

// LoaderCreate is a function that creates a new Loader (called in Run)
type ProcessorCreate func() Processor

// Loader is an interface that handles the setup of a inference processing worker and executes queries one at a time
type Processor interface {
	// Init initializes at global state for the Loader, possibly based on its worker number / ID
	Init(workerNum int, totalWorkers int, wg *sync.WaitGroup, m chan uint64, rs chan uint64)

	// ProcessInferenceQuery handles a given inference and reports its stats
	ProcessInferenceQuery(q []byte, isWarm bool, workerNum int, useReferenceData bool) ([]*Stat, error)

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
			b.br = bufio.NewReaderSize(file, defaultReadSize)
		} else {
			// Read from STDIN
			b.br = bufio.NewReaderSize(os.Stdin, defaultReadSize)
		}
	}
	return b.br
}

// Run does the bulk of the benchmark execution.
// It launches a gorountine to track stats, creates workers to process queries,
// read in the input, execute the queries, and then does cleanup.
func (b *BenchmarkRunner) Run(queryPool *sync.Pool, processorCreateFn ProcessorCreate) {

	if b.cpuProfile != "" {
		fmt.Println(fmt.Sprintf("starting cpu profile. Saving into :%s", b.cpuProfile))
		f, err := os.Create(b.cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
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
		go b.processorHandler(rateLimiter, &wg, queryPool, processorCreateFn(), i)
	}

	// Read in jobs, closing the job channel when done:
	// Wall clock start time
	wallStart := time.Now()

	// Start background reporting process
	if b.reportingPeriod.Nanoseconds() > 0 {
		go b.report(b.reportingPeriod, wallStart)
	}

	br := b.scanner.setReader(b.GetBufferedReader())
	_ = br.produce(queryPool, b.ch, rowBenchmarkNBytes, b.debug)
	close(b.ch)

	// Block for workers to finish sending requests, closing the stats channel when done:
	wg.Wait()
	b.sp.CloseAndWait()

	// Wall clock end time
	wallEnd := time.Now()
	wallTook := wallEnd.Sub(wallStart)
	_, err := fmt.Printf("Took: %8.3f sec\n", float64(wallTook.Nanoseconds())/1e9)
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

func (b *BenchmarkRunner) processorHandler(rateLimiter *rate.Limiter, wg *sync.WaitGroup, queryPool *sync.Pool, processor Processor, workerNum int) {
	buflen := uint64(len(b.ch))
	metricsChan := make(chan uint64, buflen)
	pwg := &sync.WaitGroup{}
	responseSizesChan := make(chan uint64, buflen)
	pwg.Add(1)

	processor.Init(workerNum, int(b.workers), pwg, metricsChan, responseSizesChan)

	for i := range iter.N(int(b.repetitions)) {
		for query := range b.ch {
			r := rateLimiter.ReserveN(time.Now(), 1)
			time.Sleep(r.Delay())
			stats, err := processor.ProcessInferenceQuery(query, false, workerNum, b.enableReferenceData)
			if err != nil {
				if b.IgnoreErrors() {
					fmt.Printf("On iteration %d Ignoring inference error: %v\n", i, err)
				} else {
					panic(err)
				}
			} else {
				atomic.AddUint64(&b.inferenceCount, 1)
				b.sp.sendStats(stats)
			}

			// If PrewarmQueries is set, we run the inference as 'cold' first (see above),
			// then we immediately run it a second time and report that as the 'warm' stat.
			// This guarantees that the warm stat will reflect optimal cache performance.
			if b.sp.prewarmQueries {
				// Warm run
				stats, err = processor.ProcessInferenceQuery(query, true, workerNum, b.enableReferenceData)
				if err != nil {
					if b.IgnoreErrors() {
						fmt.Printf("Ignoring inference error: %v\n", err)
					} else {
						panic(err)
					}
				} else {
					b.sp.sendStats(stats)
					atomic.AddUint64(&b.inferenceCount, 1)
				}
			}
			queryPool.Put(query)
		}
	}
	processor.Close()

	//pwg.Wait()
	close(metricsChan)
	close(responseSizesChan)

	wg.Done()
}

// report handles periodic reporting of loading stats
func (b *BenchmarkRunner) report(period time.Duration, start time.Time) {
	prevTime := start
	prevOpsCount := uint64(0)

	fmt.Printf("time (ms),total queries,instantaneous inferences/s,overall inferences/s,overall q50 lat(ms),overall q90 lat(ms),overall q95 lat(ms),overall q99 lat(ms),overall q99.999 lat(ms)\n")
	for now := range time.NewTicker(period).C {
		opsCount := atomic.LoadUint64(&b.inferenceCount)

		sinceStart := now.Sub(start)
		took := now.Sub(prevTime)
		instantRate := float64(opsCount-prevOpsCount) / float64(took.Seconds())
		overallRate := float64(opsCount) / float64(sinceStart.Seconds())
		statHist := b.sp.StatsMapping[labelAllQueries].latencyHDRHistogram

		fmt.Printf("%d,%d,%0.0f,%0.0f,%0.2f,%0.2f,%0.2f,%0.2f,%0.2f\n",
			now.UnixNano()/10e6,
			opsCount,
			instantRate,
			overallRate,
			float64(statHist.ValueAtQuantile(50.0))/10e2,
			float64(statHist.ValueAtQuantile(90.00))/10e2,
			float64(statHist.ValueAtQuantile(95.00))/10e2,
			float64(statHist.ValueAtQuantile(99.00))/10e2,
			float64(statHist.ValueAtQuantile(99.999))/10e2,
		)

		prevOpsCount = opsCount
		prevTime = now
	}
}
