package vpg

import (
	"net"

	"github.com/viktorkomarov/vpg/encoder"
)

type Writer struct {
	conn net.Conn
}

func newWriter(conn net.Conn) Writer {
	return Writer{
		conn: conn,
	}
}

func (w Writer) sendMsg(v interface{}) error {
	data, err := encoder.Encode(v)
	if err != nil {
		return err
	}

	_, err = w.conn.Write(data)
	return err
}
