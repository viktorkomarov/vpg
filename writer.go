package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

type Writer struct {
	buffer         *bytes.Buffer
	setMessageType bool
	writer         io.Writer
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
	w.setMessageType = true

	return w
}

func (w *Writer) sendMsg(msg encoder) error {
	data := msg.encode()
	size := len(data)
	if w.setMessageType {
		size++
		w.setMessageType = false
	}

	binary.Write(w.buffer, binary.BigEndian, int32(size)+4)
	w.buffer.Write(data)
	w.buffer.WriteByte('\000')
	log.Printf("%+v", w.buffer.Bytes())
	_, err := w.buffer.WriteTo(w.writer)
	return err
}
