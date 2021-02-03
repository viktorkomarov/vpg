package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
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

func (w *Writer) mType(c byte) *Writer {
	w.buffer.Reset()
	w.buffer.WriteByte(c)

	return w
}

func (w *Writer) sendMsg(msg encoder) error {
	data := msg.encode()
	size := len(data) + 4
	binary.Write(w.buffer, binary.BigEndian, int32(size))
	w.buffer.Write(data)
	log.Printf("%+v", w.buffer.Bytes())

	sender := w.buffer.Bytes()
	_, err := w.writer.Write(sender)
	return err
}
