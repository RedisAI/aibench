package inference

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// statProcessor is used to collect, analyze, and print inference execution statistics.
type statProcessor struct {
	prewarmQueries     bool       // PrewarmQueries tells the StatProcessor whether we're running each inference twice to prewarm the cache
	c                  chan *Stat // c is the channel for Stats to be sent for processing
	limit              *uint64    // limit is the number of statistics to analyze before stopping
	burnIn             uint64     // burnIn is the number of statistics to ignore before analyzing
	printInterval      uint64     // printInterval is how often print intermediate stats (number of queries)
	wg                 sync.WaitGroup
	StatsMapping       map[string]*statGroup
	InstantaneousStats *statGroup
	opsCount           uint64
}

func (sp *statProcessor) sendStats(stats []*Stat) {
	if stats == nil {
		return
	}

	for _, s := range stats {
		sp.c <- s
	}
}

// process collects latency results, aggregating them into summary
// statistics. Optionally, they are printed to stderr at regular intervals.
func (sp *statProcessor) process(workers uint, printStats bool) {
	sp.c = make(chan *Stat, workers)
	sp.wg.Add(1)
	const allQueriesLabel = labelAllQueries
	sp.StatsMapping = map[string]*statGroup{
		allQueriesLabel: newStatGroup(*sp.limit),
	}
	sp.InstantaneousStats = newStatGroup(*sp.limit)

	i := uint64(0)
	start := time.Now()
	for stat := range sp.c {
		atomic.AddUint64(&sp.opsCount, stat.totalResults)
		if sp.opsCount < sp.burnIn {
			i++
			statPool.Put(stat)
			continue
		} else if i == sp.burnIn && sp.burnIn > 0 {
			_, err := fmt.Fprintf(os.Stderr, "burn-in complete after %d queries with %d workers\n", sp.burnIn, workers)
			if err != nil {
				log.Fatal(err)
			}
		}
		if _, ok := sp.StatsMapping[string(stat.label)]; !ok {
			sp.StatsMapping[string(stat.label)] = newStatGroup(*sp.limit)
		}

		sp.StatsMapping[string(stat.label)].push(stat.value, stat.totalResults, stat.timedOut, stat.query)

		if !stat.isPartial {
			sp.StatsMapping[allQueriesLabel].push(stat.value, stat.totalResults, stat.timedOut, stat.query)
			sp.InstantaneousStats.push(stat.value, stat.totalResults, stat.timedOut, stat.query)

			// If we're prewarming queries (i.e., running them twice in a row),
			// only increment the counter for the first (cold) inference. Otherwise,
			// increment for every inference.
			if !sp.prewarmQueries || !stat.isWarm {
				i++
			}
		}

		statPool.Put(stat)
	}

	if printStats {
		sinceStart := time.Since(start)
		overallQueryRate := float64(sp.opsCount) / float64(sinceStart.Seconds())
		// the final stats output goes to stdout:
		_, err := fmt.Printf("Run complete after %d inferences with %d workers (Overall inference rate %0.2f inferences/sec):\n",
			sp.opsCount,
			workers,
			overallQueryRate)
		if err != nil {
			log.Fatal(err)
		}
		err = writeStatGroupMap(os.Stdout, sp.StatsMapping)
		if err != nil {
			log.Fatal(err)
		}
	}

	sp.wg.Done()
}

// CloseAndWait closes the stats channel and blocks until the StatProcessor has finished all the stats on its channel.
func (sp *statProcessor) CloseAndWait() {
	close(sp.c)
	sp.wg.Wait()
}
