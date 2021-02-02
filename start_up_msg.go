package main

import (
	"bytes"
	"encoding/binary"
)

type startUpMsg struct {
	fields []byte
}

const (
	protocolVersion = int32(196608)
)

func newStartUpMessage(cfg map[string]string) startUpMsg {
	fields := make([]byte, 0)
	for _, field := range []string{"user", "database"} { // replication
		fields = append(fields, field...)
		fields = append(fields, '\000')
		fields = append(fields, cfg[field]...)
		fields = append(fields, '\000')
	}
	fields = append(fields, '\000')

	return startUpMsg{
		fields: fields,
	}
}

func (s startUpMsg) encode() []byte {
	var b bytes.Buffer

	binary.Write(&b, binary.BigEndian, protocolVersion) // error
	b.Write(s.fields)

	return b.Bytes()
}
