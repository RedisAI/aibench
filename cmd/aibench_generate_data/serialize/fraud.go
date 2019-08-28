package serialize

import (
	"io"
)

// FraudSerializer writes a Transaction in a serialized form for RediSearch
type FraudSerializer struct{}

// Serialize writes Transaction data to the given writer, in a format that will be easy to create a RedisAI command
func (s *FraudSerializer) Serialize(p *Transaction, w io.Writer) (err error) {
	var buf []byte
	buf = append(buf, p.Id...)
	buf = append(buf, p.TransactionValues...)
	buf = append(buf, p.ReferenceValues...)
	_, err = w.Write(buf)
	return err
}
