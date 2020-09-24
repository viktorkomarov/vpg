package main

import (
	"errors"
	"fmt"
	"net"
)

const (
	protocolVersion = uint32(196608)
)

type Conn struct {
	address string
	conn    net.Conn
	cfg     map[string]string
	writer  *Writer
	reader  *Reader
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
		reader: NewReader(conn),
		writer: NewWriter(conn),
	}

	if err = c.init(); err != nil {
		c.conn.Close()
		return nil, err
	}

	return c, nil
}

type Authorizationer interface {
	Authorize() error
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

	client := AuthClient(classificator, c.cfg["user"], c.cfg["password"], c)
	if err := client.Authorize(); err != nil {
		return err
	}

	return c.isAuthorized()
}

func (c *Conn) receiveAuthClassificator() (*ClassificatorAuth, error) {
	msg, err := c.reader.Receive()
	if err != nil {
		return &ClassificatorAuth{}, err
	}
	if c, ok := msg.(*ClassificatorAuth); ok {
		return c, nil
	}

	return &ClassificatorAuth{}, fmt.Errorf("unknown msg %+v", msg)
}

func (c *Conn) isAuthorized() error {
	msg, err := c.receiveAuthClassificator()
	if err != nil {
		return err
	}

	if msg.Type != AuthenticationOK {
		return errors.New("unathorized")
	}

	return nil
}
