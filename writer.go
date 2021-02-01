package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
)

type Writer struct {
	buffer *bytes.Buffer
	writer *bufio.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		buffer: bytes.NewBuffer(make([]byte, 0, 1024)),
		writer: bufio.NewWriter(writer),
	}
}

// it means to set buff to 0 because we start new Message
func (w *Writer) Type(c byte) {
	w.buffer.Reset()
	w.buffer.WriteByte(c) // err is always nil
}

type Encoder interface {
	Encode() []byte
}

func (w *Writer) Payload(msg Encoder) {
	data := msg.Encode()
	size := len(data)
	_ = binary.Write(w.buffer, binary.BigEndian, uint32(size)+4) // can be error ?
	w.buffer.Write(msg.Encode())
	w.buffer.WriteByte('\000')
}

func (w *Writer) Send() error {
	_, err := w.buffer.WriteTo(w.writer)
	return err
}

func (w *Writer) SendMsg(msg Encoder) error {
	w.Payload(msg)
	return w.Send()
}
