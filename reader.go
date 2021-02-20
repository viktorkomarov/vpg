package vpg

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/viktorkomarov/vpg/encoder"
)

type Reader struct {
	reader *bufio.Reader
	conn   *Conn // for async msg update info
}

func newReader(src io.Reader, conn *Conn) Reader {
	return Reader{
		reader: bufio.NewReader(src),
		conn:   conn,
	}
}

var (
	ErrBreakingProtocol = errors.New("unknown postgres protocol action")
)

func (r Reader) recieveMsg(v interface{}) error {
	for {
		full, err := r.readBytes()
		if err != nil {
			return err
		}

		asyncMsg, err := encoder.Decode(full, v)
		if err != nil {
			return err
		}

		if asyncMsg != nil {
			// r.conn.updateInfo(asyncMsg)
			continue
		}

		return nil
	}
}

func (r Reader) readBytes() ([]byte, error) {
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

	return append(append([]byte{typeMessage}, sizeBuf...), payload...), nil
}
