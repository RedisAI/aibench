package inference

import (
	"encoding/csv"
	"io"
	"log"
	"sync"
)

// scanner is used to read in TwoWordQueries from a Reader where they are
// Go-encoded and then distribute them to workers
type scanner struct {
	r     io.Reader
	limit *uint64
}

// newScanner returns a new scanner for a given Reader and its limit
func newScanner(limit *uint64) *scanner {
	return &scanner{limit: limit}
}

// setReader sets the source, an io.Reader, that the scanner reads/decodes from
func (s *scanner) setReader(r io.Reader) *scanner {
	s.r = r
	return s
}

// scan reads encoded inference queries and places them into a channel
func (s *scanner) scan(pool *sync.Pool, c chan []string) uint64 {
	reader := csv.NewReader(s.r)
	//decoder := gob.NewDecoder(s.r)
	n := uint64(0)

	for {
		if *s.limit > 0 && n >= *s.limit {
			// request queries limit reached, time to quit
			break
		}

		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		c <- line
		//people = append(people, Person{
		//	Firstname: line[0],
		//	Lastname:  line[1],
		//	Address: &Address{
		//		City:  line[2],
		//		State: line[3],
		//	},
		//})
		n++
	}



	//for {
	//	if *s.limit > 0 && n >= *s.limit {
	//		// request queries limit reached, time to quit
	//		break
	//	}
	//
	//	q := pool.Get().(Query)
	//	err := decoder.Decode(q)
	//	if err == io.EOF {
	//		// EOF, all done
	//		break
	//	}
	//	if err == nil {
	//		// We have a inference, send it to the runner
	//		q.SetID(n)
	//		c <- q
	//		// Can't read, time to quit
	//		//	log.Fatal("error decoding inference: " , err)
	//	}
	//	if err != nil {
	//		// We have a inference, send it to the runner
	//
	//		// Can't read, time to quit
	//		log.Fatal("error decoding inference: ", err)
	//	}
	//
	//	// TwoWordQueries counter
	//	n++
	//}
	return n
}
