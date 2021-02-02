package main

import "encoding/binary"

type ReadyForQuery byte

func (r *ReadyForQuery) isMessage() {}

func NewReadyForQuery(payload []byte) *ReadyForQuery {
	t := ReadyForQuery(payload[0])
	return &t
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
