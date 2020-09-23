package main

import (
	"fmt"
	"net"

	"gitlab.com/VictorKomarov/vpg/postgres"
)

const (
	protocolVersion = uint32(196608)
)

type Conn struct {
	conn    net.Dial
	address string
	cfg     map[string]string
	writer  *postgres.Writer
	reader  *postgres.Reader
}

func New(address string) (*Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("can't init connection %w", err)
	}

	c := &Conn{
		conn: conn,
		cfg: map[string]string{
			"user":     "viktor",
			"password": "password",
			"database": "viktor",
		},
		reader: postgres.NewReader(conn),
		writer: postgres.NewWriter(conn),
	}

	if err = c.init(); err != nil {
		c.conn.Close()
		return nil, err
	}

	return c, nil
}

type Authorizationer interface {
	Authorize(conn *Conn) error
}

func (c *Conn) init() error {
	start := StartUpMsg{
		Payload: c.cfg,
	}

	if err := c.writer.Send(start); err != nil {
		return err
	}

	classificator, err := c.receiveAuthClassificator()
	if err != nil {
		return err
	}

	client := AuthClient(classificator, c.user, c.password)
	if err := client.Authorize(c); err != nil {
		return err
	}

	return nil
}

func (c *Conn) receiveAuthClassificator() (ClassificatorAuth, error) {
	msg, err := c.reader.Receive()
	if err != nil {
		return ClassificatorAuth{}, err
	}

	if c, ok := msg.(ClassificatorAuth); !ok {
		return c, nil
	}

	return ClassificatorAuth{}, fmt.Errorf("unknown msg %+v", msg)
}
