package inference

import (
	"io"
	"log"
	"os"
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

func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

// produce reads encoded inference queries and places them into a channel
func (s *producer) produce(pool *sync.Pool, c chan []byte, skipFirst bool) uint64 {
	nbytes := 8 + 120 + 1024
	bytes := make([]byte, nbytes)

	n := uint64(0)

	for {
		if *s.limit > 0 && n > *s.limit {
			// request queries limit reached, time to quit
			break
		}
		_, err := s.r.Read(bytes)
		if err != nil {
			break
		}



			c <- bytes



		n++
	}
	return n
}
