package inference

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

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
func (s *producer) produce(pool *sync.Pool, c chan []byte, nbytes int, debug int) uint64 {
	n := uint64(0)
	for {
		bytes := make([]byte, nbytes)

		if *s.limit > 0 && n >= *s.limit {
			fmt.Println(fmt.Sprintf("Reached produce limit %d", *s.limit))
			// request queries limit reached, time to quit
			break
		}
		readBytes, err := io.ReadFull(s.r, bytes)
		if readBytes == 0 {
			break
		}
		if err != nil {
			panic(fmt.Sprintf("expected to read %d bytes but got %d on row %d", nbytes, readBytes, n))
		}
		if debug > 0 {
			fmt.Fprintf(os.Stderr, "Sending Row: %d with %d bytes. \n", n, readBytes)
		}
		c <- bytes
		atomic.AddUint64(&n, 1)
	}
	return n
}
