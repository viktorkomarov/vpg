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
	binary.Write(&out, binary.BigEndian, w.Numeric)
	for i := range w.Slices {
		out.Grow(4)
		binary.Write(&out, binary.BigEndian, w.Slices[i])
	}

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
			require.EqualError(t, tC.expectedErr, err, tC.desc)
			if tC.expectedErr != nil {
				require.Equal(t, tC.v.Bytes(), data)
			}
		})
	}
}
