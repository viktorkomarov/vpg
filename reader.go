package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Reader struct {
	reader  *bufio.Reader
	sizeBuf []byte
}

func NewReader(src io.Reader) *Reader {
	return &Reader{
		reader:  bufio.NewReader(src),
		sizeBuf: make([]byte, 4),
	}
}

type Message interface {
	IsMessage()
}

var (
	ErrBreakingProtocol = errors.New("unknown postgres protocol action")
)

func (r *Reader) Receive() (Message, error) {
	t, err := r.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("can't read msg type %w", err)
	}

	if _, err = r.reader.Read(r.sizeBuf); err != nil {
		return nil, fmt.Errorf("can't read msg size %w", err)
	}

	size := binary.BigEndian.Uint32(r.sizeBuf)

	payload := make([]byte, size-4)
	if _, err := r.reader.Read(payload); err != nil {
		return nil, fmt.Errorf("can't read msg payload %w", err)
	}

	switch rune(t) {
	case 'E':
		return nil, fmt.Errorf("%s", payload)
	case 'R':
		return receiveAuth(payload)
	default:
		return nil, fmt.Errorf("unknown msg type %s %w", t, ErrBreakingProtocol)
	}
}

func receiveAuth(payload []byte) (Message, error) {
	authType := AuthenticationResponseType(binary.BigEndian.Uint32(payload[:4]))

	switch authType {
	case AuthenticationOK, AuthenticationCleartextPassword:
		return &ClassificatorAuth{Type: authType}, nil
	case AuthenticationMD5Password, AuthenticationSASL:
		return &ClassificatorAuth{Type: authType, Payload: payload[4:]}, nil
	case AuthenticationSASLContinue:
		return NewSASSLContinue(payload[4:])
	case AuthenticationSASLFinal:
		return NewSASLFinal(payload[4:])
	default:
		return nil, fmt.Errorf("unknown autherozied msg  %d %s", authType, payload[5:])
	}
}
