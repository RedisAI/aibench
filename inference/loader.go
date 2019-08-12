package inference

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	defaultLoadSize = 4 << 20 // 4 MB
)

// LoadRunner contains the common components for running a inference benchmarking
// program against a database.
type LoadRunner struct {
	// flag fields
	limit         uint64
	workers       uint
	fileName      string
	skipFirstLine bool
	seed          int64

	// non-flag fields
	br      *bufio.Reader
	sp      *statProcessor
	scanner *producer
	ch      chan []string
}

// NewLoadRunner creates a new instance of LoadRunner which is
// common functionality to be used by inference benchmarker programs
func NewLoadRunner() *LoadRunner {
	runner := &LoadRunner{}
	runner.scanner = newScanner(&runner.limit)
	runner.sp = &statProcessor{
		limit: &runner.limit,
	}
	flag.Uint64Var(&runner.limit, "max-queries", 0, "Limit the number of queries to send, 0 = no limit")
	flag.UintVar(&runner.workers, "workers", 1, "Number of concurrent requests to make.")
	flag.BoolVar(&runner.skipFirstLine, "skip-first-line", true, "Skip first line of csv (default true).")
	flag.Int64Var(&runner.seed, "seed", 0, "PRNG seed (default, or 0, uses the current timestamp).")
	flag.StringVar(&runner.fileName, "file", "", "File name to read queries from")
	return runner
}

// SetLimit changes the number of queries to run, with 0 being all of them
func (b *LoadRunner) SetLimit(limit uint64) {
	b.limit = limit
}

// LoaderCreate is a function that creates a new Loader (called in Run)
type LoaderCreate func() Loader

// Loader is an interface that handles the setup of a inference processing worker and executes queries one at a time
type Loader interface {
	// Init initializes at global state for the Loader, possibly based on its worker number / ID
	Init(workerNum int, wg *sync.WaitGroup)

	// ProcessInferenceQuery handles a given inference and reports its stats
	ProcessLoadQuery(q []string) ([]*Stat, error)
}

// GetBufferedReader returns the buffered Reader that should be used by the loader
func (b *LoadRunner) GetBufferedReader() *bufio.Reader {
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
func (b *LoadRunner) RunLoad(queryPool *sync.Pool, LoaderCreateFn LoaderCreate) {

	rand.Seed(b.seed)

	if b.workers == 0 {
		panic("must have at least one worker")
	}
	b.ch = make(chan []string, b.workers)

	// Launch the stats processor:
	go b.sp.process(b.workers)

	// Launch inference processors
	var wg sync.WaitGroup
	for i := 0; i < int(b.workers); i++ {
		wg.Add(1)
		go b.loadHandler(&wg, queryPool, LoaderCreateFn(), i)
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

}

func (b *LoadRunner) loadHandler(wg *sync.WaitGroup, queryPool *sync.Pool, processor Loader, workerNum int) {
	pwg := &sync.WaitGroup{}
	pwg.Add(1)

	processor.Init(workerNum, pwg)

	for query := range b.ch {
		_, err := processor.ProcessLoadQuery(query)
		if err != nil {
			panic(err)
		}
	}

	//pwg.Wait()

	wg.Done()
}
