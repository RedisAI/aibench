package fraud

import (
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/serialize"
)

type commonaibenchSimulatorConfig struct {
	InputFilename string
	// Start is the beginning time for the Simulator
}

type commonaibenchSimulator struct {
	maxTransactions uint64
	recordIndex     uint64
	records         []serialize.Transaction
}

// Finished tells whether we have simulated all the necessary documents
func (s *commonaibenchSimulator) Finished() bool {
	return s.recordIndex >= s.maxTransactions
}
