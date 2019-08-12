package inference

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"
	"sync"
)

// producer is used to read in TwoWordQueries from a Reader where they are
// Go-encoded and then distribute them to workers
type producer struct {
	r     io.Reader
	limit *uint64
}

// newScanner returns a new producer for a given Reader and its limit
func newScanner(limit *uint64) *producer {
	return &producer{limit: limit}
}

// setReader sets the source, an io.Reader, that the producer reads/decodes from
func (s *producer) setReader(r io.Reader) *producer {
	s.r = r
	return s
}

// produce reads encoded inference queries and places them into a channel
func (s *producer) produce(pool *sync.Pool, c chan []string, skip_first bool) uint64 {
	reader := csv.NewReader(s.r)
	n := uint64(0)

	for {
		if *s.limit > 0 && n > *s.limit {
			// request queries limit reached, time to quit
			break
		}

		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		if n > 0 && skip_first == true || skip_first == false {
			ns := []string{strconv.FormatUint(n, 10)}
			ns = append(ns, line...)
			c <- ns

		}

		n++
	}
	return n
}
