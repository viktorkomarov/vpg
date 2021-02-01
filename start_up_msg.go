package main

import (
	"bytes"
	"encoding/binary"
)

type StartUpMsg struct {
	fields []byte
}

const (
	protocolVersion = uint32(196608)
)

// no need to check because we validate config earler
func NewStartUpMessage(cfg map[string]string) StartUpMsg {
	fields := make([]byte, 0)
	for _, field := range []string{"user", "database"} { // replication
		fields = append(fields, field...)
		fields = append(fields, '\000')
		fields = append(fields, cfg[field]...)
		fields = append(fields, '\000')
	}
	fields = append(fields, '\000')

	return StartUpMsg{
		fields: fields,
	}
}

func (s StartUpMsg) Encode() []byte {
	var b bytes.Buffer

	binary.Write(&b, binary.BigEndian, protocolVersion) // error
	b.Write(s.fields)

	return b.Bytes()
}
