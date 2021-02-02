package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

type Writer struct {
	buffer *bytes.Buffer
	writer *bufio.Writer
}

func newWriter(writer io.Writer) *Writer {
	return &Writer{
		buffer: bytes.NewBuffer(make([]byte, 0, 1024)),
		writer: bufio.NewWriter(writer),
	}
}

type encoder interface {
	encode() []byte
}

func (w *Writer) payload(msg encoder) {
	data := msg.encode()
	size := len(data)
	_ = binary.Write(w.buffer, binary.BigEndian, int32(size)) // can be error ?
	w.buffer.Write(data)
	w.buffer.WriteByte('\000')
}

func (w *Writer) send() error {
	data := w.buffer.Bytes()
	log.Printf("%+v", data)
	_, err := w.buffer.WriteTo(w.writer)
	return err
}

func (w *Writer) sendMsg(msg encoder) error {
	w.payload(msg)
	return w.send()
}
