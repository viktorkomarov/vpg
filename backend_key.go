package main

import (
	"encoding/binary"
	"fmt"
)

type BackendKeyData struct {
	PID int32
	Key int32
}

func (b *BackendKeyData) isMessage() {}

func NewBackendKeyData(payload []byte) (*BackendKeyData, error) {
	if len(payload) < 8 {
		return nil, fmt.Errorf("uncorrect backend_key_data len %v", payload)
	}

	return &BackendKeyData{
		PID: int32(binary.BigEndian.Uint32(payload[:4])),
		Key: int32(binary.BigEndian.Uint32(payload[4:])),
	}, nil
}
