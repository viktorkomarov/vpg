package main

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

type Byter interface {
	Bytes() []byte
}

type WithoutHeader struct {
	Payload string `pg_order:"1" pg_preffix:"name"`
	Numeric int    `pg_order:"2"`
	Slices  []int  `pg_order:"3" pg_preffix:"slice"`
}

func (w WithoutHeader) Bytes() []byte {
	var out bytes.Buffer

	out.Write([]byte{0, 0, 0, 0})
	out.Write(append([]byte("name"), '\000'))
	out.Write(append([]byte(w.Payload), '\000'))
	out.Grow(4)
	binary.Write(&out, binary.BigEndian, int32(w.Numeric))
	out.Write(append([]byte("slice"), '\000'))
	for _, n := range w.Slices {
		out.Grow(4)
		binary.Write(&out, binary.BigEndian, int32(n))
	}

	data := out.Bytes()
	binary.BigEndian.PutUint32(data[0:4], uint32(len(data)))
	return out.Bytes()
}

func TestEncode(t *testing.T) {
	testCases := []struct {
		desc        string
		v           Byter
		expectedErr error
	}{
		{
			desc: "WithouHeader",
			v:    WithoutHeader{Payload: "payload", Numeric: 5, Slices: []int{1, 2, 3, 1404}},
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			data, err := Encode(tC.v)
			if tC.expectedErr == nil {
				require.Equal(t, tC.v.Bytes(), data)
			} else {
				require.Error(t, tC.expectedErr, err.Error(), tC.desc)
			}
		})
	}
}
