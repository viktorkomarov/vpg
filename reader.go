package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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
	typeMessage, err := r.reader.ReadByte() // maybe use ioutil
	if err != nil {
		return nil, fmt.Errorf("can't read msg type %w", err)
	}

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
		return newParametrStatus(payload), nil
	case 'K':
		return NewBackendKeyData(payload)
	case 'Z':
		return newReadyForQuery(payload)
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
		return nil, fmt.Errorf("unknown autherization msg  %d %s", authType, payload[5:])
	}
}
