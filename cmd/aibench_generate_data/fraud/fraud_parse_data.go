package fraud

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"github.com/filipecosta90/aibench/cmd/aibench_generate_data/common"
	"github.com/filipecosta90/aibench/cmd/aibench_generate_data/serialize"
	"io"
	"log"
	"os"
	"strconv"
	"sync/atomic"
)

// A FTSSimulator generates data similar to telemetry from Telegraf for only CPU metrics.
// It fulfills the Simulator interface.
type FTSSimulator struct {
	*commonaibenchSimulator
}

// Next advances a Transaction to the next state in the generator.
func (d *FTSSimulator) Next(p *serialize.Transaction) bool {
	//// Switch to the next document
	//if d.recordIndex >= uint64(len(d.records)) {
	//	d.recordIndex = 0
	//}
	return d.populateTransaction(p)
}

func (d *FTSSimulator) populateTransaction(p *serialize.Transaction) bool {
	record := &d.records[d.recordIndex]

	p.Id = record.Id
	p.TransactionValues = record.TransactionValues
	p.ReferenceValues = record.ReferenceValues

	ret := d.recordIndex < d.maxTransactions
	atomic.AddUint64(&d.recordIndex, 1)
	return ret
}

// aibenchSimulatorConfig is used to create a FTSSimulator.
type AibenchSimulatorConfig commonaibenchSimulatorConfig

// NewSimulator produces a Simulator that conforms to the given SimulatorConfig over the specified interval
func (c *AibenchSimulatorConfig) NewSimulator(limit uint64, inputFilename string, debug int) common.Simulator {

	file, err := os.Open(inputFilename)
	var transactions []serialize.Transaction
	maxPoints := limit

	if err != nil {
		panic(fmt.Sprintf("cannot open file for read %s: %v", inputFilename, err))
	}

	if debug > 0 {
		fmt.Fprintln(os.Stderr, "started reading "+inputFilename)
	}
	transactionCount := uint64(0)
	seekCount := uint64(0)
	// if the transaction count is lower than the limit or if there is no limit and we havent read the file one
	for transactionCount < limit || (limit == 0 && seekCount < 1) {

		file.Seek(0, io.SeekStart)
		br := bufio.NewReader(file)
		reader := csv.NewReader(br)

		//skip first line
		line, err := reader.Read()
		if err != nil {
			log.Fatal(err)
		}
		line, err = reader.Read()

		for err != io.EOF && (transactionCount < limit || (limit == 0 && seekCount < 1)) {
			qfloat := ConvertSliceStringToFloat(line)
			qbytes := Float32bytes(qfloat[0])
			for _, value := range qfloat[1:30] {
				qbytes = append(qbytes, Float32bytes(value)...)
			}

			refFloats := randReferenceData(256)
			refBytes := Float32bytes(refFloats[0])

			for _, value := range refFloats[1:256] {
				refBytes = append(refBytes, Float32bytes(value)...)
			}
			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, transactionCount)
			transactions = append(transactions, serialize.Transaction{Id: buf, TransactionValues: qbytes, ReferenceValues: refBytes})
			if debug > 0 {
				if transactionCount%1000 == 0 {
					fmt.Fprintln(os.Stderr, "At transaction "+strconv.Itoa(int(transactionCount)))
				}
			}
			atomic.AddUint64(&transactionCount, 1)
			line, err = reader.Read()
		}
		atomic.AddUint64(&seekCount, 1)

	}

	maxPoints = uint64(len(transactions))
	if limit > 0 && limit < uint64(len(transactions)) {
		// Set specified points number limit
		maxPoints = limit
	}
	sim := &FTSSimulator{&commonaibenchSimulator{
		maxTransactions: maxPoints,
		recordIndex:     0,
		records:         transactions,
	}}

	if debug > 0 {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("finished reading %s, max transactions %d", inputFilename, maxPoints))
	}

	return sim
}
