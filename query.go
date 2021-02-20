package vpg

import (
	"encoding/binary"
	"fmt"
)

type readyForQuery byte

func (r readyForQuery) isMessage() {}

func newReadyForQuery(payload []byte) (readyForQuery, error) {
	if len(payload) < 0 {
		return readyForQuery('e'), fmt.Errorf("incorrect ready for queue %+v", payload)
	}

	return readyForQuery(payload[0]), nil
}

type Query struct {
	Text string
}

func (q *Query) Encode() []byte {
	buf := []byte{'Q', 0, 0, 0, 0}

	binary.BigEndian.PutUint32(buf[1:5], uint32(len(q.Text)+5))
	buf = append(buf, q.Text...)
	buf = append(buf, '\000')

	return buf
}
