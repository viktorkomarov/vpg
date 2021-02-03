package main

import (
	"encoding/binary"
	"fmt"
)

type backendKeyData struct {
	PID int32
	Key int32
}

func (b backendKeyData) isMessage() {}

func NewBackendKeyData(payload []byte) (backendKeyData, error) {
	if len(payload) < 8 {
		return backendKeyData{}, fmt.Errorf("incorrect backend_key_data len %v", payload)
	}

	return backendKeyData{
		PID: int32(binary.BigEndian.Uint32(payload[:4])),
		Key: int32(binary.BigEndian.Uint32(payload[4:])),
	}, nil
}
