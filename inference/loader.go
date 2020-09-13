package inference

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// LoadRunner contains the common components for running a inference benchmarking
// program against a database.
type LoadRunner struct {
	// flag fields
	limit    uint64
	workers  uint
	fileName string
	debug    int

	// non-flag fields
	br              *bufio.Reader
	sp              *statProcessor
	scanner         *producer
	ch              chan []byte
	reportingPeriod time.Duration
	commandCount    uint64
}

// NewLoadRunner creates a new instance of LoadRunner which is
// common functionality to be used by inference benchmarker programs
func NewLoadRunner() *LoadRunner {
	runner := &LoadRunner{}
	runner.scanner = newScanner(&runner.limit)
	runner.sp = &statProcessor{
		limit: &runner.limit,
	}
	flag.Uint64Var(&runner.limit, "max-inserts", 0, "Limit the number of inserts, 0 = no limit")
	flag.UintVar(&runner.workers, "workers", 1, "Number of concurrent requests to make.")
	flag.StringVar(&runner.fileName, "file", "", "File name to read queries from")
	flag.IntVar(&runner.debug, "debug", 0, "Whether to print debug messages.")
	flag.DurationVar(&runner.reportingPeriod, "reporting-period", 1*time.Second, "Period to report write stats")

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
	ProcessLoadQuery(q []byte, debug int) ([]*Stat, uint64, error)
	Close()
}

// GetBufferedReader returns the buffered Reader that should be used by the loader
func (b *LoadRunner) GetBufferedReader() *bufio.Reader {
	if b.br == nil {
		if len(b.fileName) > 0 {
			// Read from specified file
			file, err := os.Open(b.fileName)
			log.Printf("Reading %s\n", b.fileName )
			if err != nil {
				panic(fmt.Sprintf("cannot open file for read %s: %v", b.fileName, err))
			}
			b.br = bufio.NewReaderSize(file, defaultReadSize)
		} else {
			// Read from STDIN
			log.Printf("Reading from STDIN\n" )
			b.br = bufio.NewReaderSize(os.Stdin, defaultReadSize)
		}
	}
	return b.br
}

// Run does the bulk of the benchmark execution.
// It launches a gorountine to track stats, creates workers to process queries,
// read in the input, execute the queries, and then does cleanup.
func (b *LoadRunner) RunLoad(queryPool *sync.Pool, LoaderCreateFn LoaderCreate, rowBenchmarkNBytes int) {

	if b.workers == 0 {
		panic("must have at least one worker")
	}
	b.ch = make(chan []byte, b.workers)

	// Launch the stats processor:
	go b.sp.process(b.workers, false)

	// Launch inference processors
	var wg sync.WaitGroup
	for i := 0; i < int(b.workers); i++ {
		wg.Add(1)
		go b.loadHandler(&wg, queryPool, LoaderCreateFn(), i)
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

}

func (b *LoadRunner) loadHandler(wg *sync.WaitGroup, queryPool *sync.Pool, processor Loader, workerNum int) {
	pwg := &sync.WaitGroup{}
	pwg.Add(1)

	processor.Init(workerNum, pwg)

	for query := range b.ch {
		_, ncommands, err := processor.ProcessLoadQuery(query, b.debug)
		if err != nil {
			panic(err)
		}
		atomic.AddUint64(&b.commandCount, ncommands)
	}

	processor.Close()

	//pwg.Wait()

	wg.Done()
}

// report handles periodic reporting of loading stats
func (b *LoadRunner) report(period time.Duration, start time.Time) {
	prevTime := start
	prevInfCount := uint64(0)

	fmt.Printf("time (ns),total commands,instantaneous commands/s,overall commands/s\n")
	for now := range time.NewTicker(period).C {
		infCount := atomic.LoadUint64(&b.commandCount)

		sinceStart := now.Sub(start)
		took := now.Sub(prevTime)
		instantInfRate := float64(infCount-prevInfCount) / float64(took.Seconds())
		overallInfRate := float64(infCount) / float64(sinceStart.Seconds())

		fmt.Printf("%d,%d,%0.2f,%0.2f\n", now.UnixNano(), infCount, instantInfRate, overallInfRate)

		prevInfCount = infCount
		prevTime = now
	}
}
