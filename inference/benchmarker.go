package inference

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

const (
	labelAllQueries  = "All queries"
	labelColdQueries = "Cold queries"
	labelWarmQueries = "Warm queries"

	defaultReadSize = 4 << 20 // 4 MB
)

// LoadRunner contains the common components for running a inference benchmarking
// program against a database.
type BenchmarkRunner struct {
	// flag fields
	dbName         string
	limit          uint64
	memProfile     string
	workers        uint
	printResponses bool
	debug          int
	fileName       string
	skipFirstLine  bool
	seed           int64

	// non-flag fields
	br      *bufio.Reader
	sp      *statProcessor
	scanner *producer
	ch      chan []string
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
	flag.UintVar(&runner.workers, "workers", 1, "Number of concurrent requests to make.")
	flag.BoolVar(&runner.sp.prewarmQueries, "prewarm-queries", false, "Run each inference twice in a row so the warm inference is guaranteed to be a cache hit")
	flag.BoolVar(&runner.printResponses, "print-responses", false, "Pretty print response bodies for correctness checking (default false).")
	flag.BoolVar(&runner.skipFirstLine, "skip-first-line", true, "Skip first line of csv (default true).")
	flag.IntVar(&runner.debug, "debug", 0, "Whether to print debug messages.")
	flag.Int64Var(&runner.seed, "seed", 0, "PRNG seed (default, or 0, uses the current timestamp).")
	flag.StringVar(&runner.fileName, "file", "", "File name to read queries from")
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

// LoaderCreate is a function that creates a new Loader (called in Run)
type ProcessorCreate func() Processor

// Loader is an interface that handles the setup of a inference processing worker and executes queries one at a time
type Processor interface {
	// Init initializes at global state for the Loader, possibly based on its worker number / ID
	Init(workerNum int, wg *sync.WaitGroup, m chan uint64, rs chan uint64)

	// ProcessInferenceQuery handles a given inference and reports its stats
	ProcessInferenceQuery(q []string, isWarm bool) ([]*Stat, error)
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

	rand.Seed(b.seed)

	if b.workers == 0 {
		panic("must have at least one worker")
	}
	if b.sp.burnIn > b.limit {
		panic("burn-in is larger than limit")
	}
	b.ch = make(chan []string, b.workers)

	// Launch the stats processor:
	go b.sp.process(b.workers)

	// Launch inference processors
	var wg sync.WaitGroup
	for i := 0; i < int(b.workers); i++ {
		wg.Add(1)
		go b.processorHandler(&wg, queryPool, processorCreateFn(), i)
	}

	// Read in jobs, closing the job channel when done:
	// Wall clock start time
	wallStart := time.Now()
	br := b.scanner.setReader(b.GetBufferedReader())
	_ = br.produce(queryPool, b.ch, b.skipFirstLine)
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

	// (Optional) create a memory profile:
	if len(b.memProfile) > 0 {
		f, err := os.Create(b.memProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
	}
}

func (b *BenchmarkRunner) processorHandler(wg *sync.WaitGroup, queryPool *sync.Pool, processor Processor, workerNum int) {
	buflen := uint64(len(b.ch))
	metricsChan := make(chan uint64, buflen)
	pwg := &sync.WaitGroup{}
	responseSizesChan := make(chan uint64, buflen)
	pwg.Add(1)

	processor.Init(workerNum, pwg, metricsChan, responseSizesChan)

	for query := range b.ch {
		stats, err := processor.ProcessInferenceQuery(query, false)
		if err != nil {
			panic(err)
		}
		b.sp.sendStats(stats)

		// If PrewarmQueries is set, we run the inference as 'cold' first (see above),
		// then we immediately run it a second time and report that as the 'warm' stat.
		// This guarantees that the warm stat will reflect optimal cache performance.
		if b.sp.prewarmQueries {
			// Warm run
			stats, err = processor.ProcessInferenceQuery(query, true)
			if err != nil {
				panic(err)
			}
			b.sp.sendStatsWarm(stats)
		}
		queryPool.Put(query)
	}

	//pwg.Wait()
	close(metricsChan)
	close(responseSizesChan)

	wg.Done()
}
