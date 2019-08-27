package fraud

import (
	"github.com/filipecosta90/aibench/cmd/aibench_generate_data/serialize"
)

type commonaibenchSimulatorConfig struct {
	InputFilename string
	// Start is the beginning time for the Simulator
}

type commonaibenchSimulator struct {
	madeTransactions uint64
	maxTransactions  uint64
	recordIndex      uint64
	records          []serialize.Transaction
}

// Finished tells whether we have simulated all the necessary documents
func (s *commonaibenchSimulator) Finished() bool {
	return s.madeTransactions >= s.maxTransactions
}
