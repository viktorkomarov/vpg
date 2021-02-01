package main

import (
	"io"
)

type Writer struct {
	sender io.Writer
}

func NewWriter(sender io.Writer) *Writer {
	return &Writer{
		sender: sender,
	}
}

func (w *Writer) Send(msg Encoder) error {
	if _, err := w.sender.Write(msg.Encode()); err != nil {
		return err
	}

	return nil
}
