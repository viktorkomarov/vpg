package main

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Writer struct {
	buffer *bytes.Buffer
	writer io.Writer
}

func newWriter(writer io.Writer) *Writer {
	return &Writer{
		buffer: bytes.NewBuffer(make([]byte, 0, 1024)),
		writer: writer,
	}
}

type encoder interface {
	encode() []byte
}

func (w *Writer) payload(msg encoder) {
	data := msg.encode()
	size := len(data)
	_ = binary.Write(w.buffer, binary.BigEndian, int32(size)+4) // can be error ?
	w.buffer.Write(data)
	w.buffer.WriteByte('\000')
}

func (w *Writer) send() error {
	_, err := w.buffer.WriteTo(w.writer)
	return err
}

func (w *Writer) sendMsg(msg encoder) error {
	w.payload(msg)
	return w.send()
}
