package fraud

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/common"
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/serialize"
	"github.com/RedisAI/aibench/inference"
	"github.com/mediocregopher/radix/v3"
	"io"
	"log"
	"os"
	"strconv"
	"sync/atomic"
)

type FTSSimulator struct {
	*commonaibenchSimulator
}

// Next advances a Transaction to the next state in the generator.
func (d *FTSSimulator) Next(p *serialize.Transaction) bool {
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

		_, err := file.Seek(0, io.SeekStart)
		if err != nil {
			log.Fatal(err)
		}
		br := bufio.NewReader(file)
		reader := csv.NewReader(br)

		//skip first line
		_, err = reader.Read()
		if err != nil {
			log.Fatal(err)
		}
		line, err := reader.Read()

		for err != io.EOF && (transactionCount < limit || (limit == 0 && seekCount < 1)) {
			qfloat := inference.ConvertSliceStringToFloat(line)
			qbytes := inference.Float32bytes(qfloat[0])
			for _, value := range qfloat[1:30] {
				qbytes = append(qbytes, inference.Float32bytes(value)...)
			}

			refFloats := inference.RandReferenceData(256)
			refBytes := inference.Float32bytes(refFloats[0])

			for _, value := range refFloats[1:256] {
				refBytes = append(refBytes, inference.Float32bytes(value)...)
			}
			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, transactionCount)
			crc := make([]byte, 2)
			binary.LittleEndian.PutUint16(crc, radix.CRC16(buf))

			transactions = append(transactions, serialize.Transaction{Id: buf, TransactionValues: qbytes, ReferenceValues: refBytes, Slot: crc})
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

	var maxPoints = uint64(len(transactions))
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
		fmt.Fprintf(os.Stderr, "finished reading %s, max transactions %d\n", inputFilename, maxPoints)
	}

	return sim
}
