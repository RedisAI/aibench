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
)

// A FTSSimulator generates data similar to telemetry from Telegraf for only CPU metrics.
// It fulfills the Simulator interface.
type FTSSimulator struct {
	*commonAIBenchSimulator
}

// Next advances a Transaction to the next state in the generator.
func (d *FTSSimulator) Next(p *serialize.Transaction) bool {
	// Switch to the next document
	if d.recordIndex >= uint64(len(d.records)) {
		d.recordIndex = 0
	}
	return d.populateTransaction(p)
}

func (d *FTSSimulator) populateTransaction(p *serialize.Transaction) bool {
	record := &d.records[d.recordIndex]

	p.Id = record.Id
	p.TransactionValues = record.TransactionValues
	p.ReferenceValues = record.ReferenceValues

	ret := d.recordIndex < uint64(len(d.records))
	d.recordIndex = d.recordIndex + 1
	d.madeTransactions = d.madeTransactions + 1
	return ret
}

// AIBenchSimulatorConfig is used to create a FTSSimulator.
type AIBenchSimulatorConfig commonAIBenchSimulatorConfig

// NewSimulator produces a Simulator that conforms to the given SimulatorConfig over the specified interval
func (c *AIBenchSimulatorConfig) NewSimulator(limit uint64, inputFilename string, debug int) common.Simulator {

	file, err := os.Open(inputFilename)
	var transactions []serialize.Transaction
	maxPoints := limit

	if err != nil {
		panic(fmt.Sprintf("cannot open file for read %s: %v", inputFilename, err))
	}
	br := bufio.NewReader(file)
	reader := csv.NewReader(br)

	if debug > 0 {
		fmt.Fprintln(os.Stderr, "started reading "+inputFilename)
	}
	transactionCount := uint64(0)
	//skip first line
	_, error := reader.Read()
	if error != nil {
		log.Fatal(error)
	}
	for err != io.EOF && (transactionCount < limit || limit == 0) {

		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
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
		transactionCount++

	}

	if debug > 0 {
		fmt.Fprintln(os.Stderr, "finished reading "+inputFilename)
	}

	maxPoints = uint64(len(transactions))
	if limit > 0 && limit < uint64(len(transactions)) {
		// Set specified points number limit
		maxPoints = limit
	}
	sim := &FTSSimulator{&commonAIBenchSimulator{
		madeTransactions: 0,
		maxTransactions:  maxPoints,

		recordIndex: 0,
		records:     transactions,
	}}

	return sim
}
