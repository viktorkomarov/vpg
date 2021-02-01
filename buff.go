package main

import (
	"bytes"
	"encoding/binary"
)

type WriteBuff struct {
	buffer *bytes.Buffer // change to bufio...
}

func NewBuff() WriteBuff {
	return WriteBuff{
		buffer: bytes.NewBuffer(make([]byte, 0, 1024)),
	}
}

// it means to set buff to 0 because we start new Message
func (w WriteBuff) Type(c byte) {
	w.buffer.Reset()
	w.buffer.WriteByte(c) // err is always nil
}

type Encoder interface {
	Encode() []byte
}

func (w WriteBuff) Payload(msg Encoder) {
	data := msg.Encode()
	size := len(data)
	_ = binary.Write(w.buffer, binary.BigEndian, uint32(size)+4) // can be error ?
	w.buffer.Write(msg.Encode())
}

func (w WriteBuff) Message() []byte {
	w.buffer.WriteByte('\000')
	return w.buffer.Bytes()
}
