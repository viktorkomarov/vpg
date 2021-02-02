package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

type Reader struct {
	reader *bufio.Reader
}

func newReader(src io.Reader) *Reader {
	return &Reader{
		reader: bufio.NewReader(src),
	}
}

type message interface {
	isMessage()
}

var (
	ErrBreakingProtocol = errors.New("unknown postgres protocol action")
)

func (r *Reader) receive() (message, error) {
	log.Printf("try to read first byte")
	typeMessage, err := r.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("can't read msg type %w", err)
	}
	log.Printf("read first byte")
	sizeBuf := make([]byte, 4)
	if _, err = r.reader.Read(sizeBuf); err != nil {
		return nil, fmt.Errorf("can't read msg size %w", err)
	}

	size := binary.BigEndian.Uint32(sizeBuf)
	payload := make([]byte, size-4)
	if _, err := r.reader.Read(payload); err != nil {
		return nil, fmt.Errorf("can't read msg payload %w", err)
	}

	switch rune(typeMessage) {
	case 'E':
		return nil, NewErrPostgresResponse(payload)
	case 'R':
		return receiveAuth(payload)
	case 'T':
		return NewRowDescription(payload)
	case 'S':
		return NewParametrStatus(payload), nil
	case 'K':
		return NewBackendKeyData(payload)
	case 'Z':
		return NewReadyForQuery(payload), nil
	default:
		return nil, fmt.Errorf("unknown msg type %s %w", string(typeMessage), ErrBreakingProtocol)
	}
}

func receiveAuth(payload []byte) (message, error) {
	authType := authenticationResponseType(binary.BigEndian.Uint32(payload[:4]))

	switch authType {
	case authenticationOK, authenticationCleartextPassword:
		return auth{Type: authType}, nil
	case authenticationMD5Password, authenticationSASL:
		return auth{Type: authType, Payload: payload[4:]}, nil
	//case authenticationSASLContinue:
	//return NewSASSLContinue(payload[4:])
	//case authenticationSASLFinal:
	//return NewSASLFinal(payload[4:])
	default:
		return nil, fmt.Errorf("unknown autherozied msg  %d %s", authType, payload[5:])
	}
}
