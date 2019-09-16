package common

import (
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/serialize"
)

// SimulatorConfig is an interface to create a Simulator
type SimulatorConfig interface {
	NewSimulator(uint64, string, int) Simulator
}

// Simulator simulates a use case.
type Simulator interface {
	Finished() bool
	Next(transaction *serialize.Transaction) bool
}
