package inference

import (
	"fmt"
	"sync"
)

// RedisAI encodes a RedisAI request. This will be serialized for use
// by the dlbench_run_inference_redisai program.
type RedisAI struct {
	HumanLabel       []byte
	HumanDescription []byte

	RedisQuery []byte
	id         uint64
}

// RedisAIPool is a sync.Pool of RedisAI Query types
var RedisAIPool = sync.Pool{
	New: func() interface{} {
		return &RedisAI{
			HumanLabel:       make([]byte, 0, 1024),
			HumanDescription: make([]byte, 0, 1024),
			RedisQuery:       make([]byte, 0, 1024),
		}
	},
}

// NewRediSearch returns a new RedisAI Query instance
func NewRediSearch() *RedisAI {
	return RedisAIPool.Get().(*RedisAI)
}

// GetID returns the ID of this Query
func (q *RedisAI) GetID() uint64 {
	return q.id
}

// SetID sets the ID for this Query
func (q *RedisAI) SetID(n uint64) {
	q.id = n
}

// String produces a debug-ready description of a Query.
func (q *RedisAI) String() string {
	return fmt.Sprintf("HumanLabel: %s, HumanDescription: %s, Query: %s", q.HumanLabel, q.HumanDescription, q.RedisQuery)
}

// HumanLabelName returns the human readable name of this Query
func (q *RedisAI) HumanLabelName() []byte {
	return q.HumanLabel
}

// HumanDescriptionName returns the human readable description of this Query
func (q *RedisAI) HumanDescriptionName() []byte {
	return q.HumanDescription
}

// Release resets and returns this Query to its pool
func (q *RedisAI) Release() {
	q.HumanLabel = q.HumanLabel[:0]
	q.HumanDescription = q.HumanDescription[:0]
	q.id = 0

	q.RedisQuery = q.RedisQuery[:0]

	RedisAIPool.Put(q)
}
