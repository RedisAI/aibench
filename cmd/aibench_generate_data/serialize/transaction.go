package serialize

// Transaction wraps a single document. It stores database-agnostic data
import (
	"io"
)

// Transaction wraps a single document. It stores database-agnostic data
// representing one Transaction
//
// Internally, Transaction uses byte slices instead of strings to try to minimize
// overhead.
type Transaction struct {
	Id, TransactionValues, ReferenceValues []byte
}

// NewTransaction returns a new empty Transaction
func NewTransaction() *Transaction {
	return &Transaction{
		Id:                make([]byte, 8),
		TransactionValues: make([]byte, 120),
		ReferenceValues:   make([]byte, 1028),
	}
}

// Reset clears all information from this Transaction so it can be reused.
func (p *Transaction) Reset() {
	p.Id = p.Id[:0]
	p.TransactionValues = p.TransactionValues[:0]
	p.ReferenceValues = p.ReferenceValues[:0]
}

// TransactionSerializer serializes a Transaction for writing
type TransactionSerializer interface {
	Serialize(p *Transaction, w io.Writer) error
}
