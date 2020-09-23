package main

import "encoding/binary"

type StartUpMsg struct {
	Payload map[string]string
}

func (s StartUpMsg) Encode() []byte {
	dst := make([]byte, 9)

	dst[0] = 'p'
	binary.BigEndian.PutUint32(dst[4:8], protocolVersion)
	for key, val := range s.Payload {
		dst = append(dst, key...)
		dst = append(dst, '\000')
		dst = append(dst, val...)
		dst = append(dst, '\000')
	}
	binary.BigEndian.PutUint32(dst[0:4], uint32(len(dst)))

	return dst
}
