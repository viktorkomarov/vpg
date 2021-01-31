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
	cfg    map[string]string
	conn   net.Conn
	writer *Writer
	reader *Reader

	parametersStatus map[string]string
	pid              int32
	key              int32
	status           byte
}

func New(cfg map[string]string) (*Conn, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", cfg["address"]) // read about SetDeadline
	if err != nil {
		return nil, fmt.Errorf("can't init connection %w", err)
	}

	c := &Conn{
		conn:             conn,
		cfg:              cfg,
		reader:           NewReader(conn),
		writer:           NewWriter(conn),
		parametersStatus: make(map[string]string),
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

	if err = c.isAuthorized(); err != nil {
		return err
	}

	for { // enrich conn
		msg, err := c.reader.Receive()
		if err != nil {
			return err
		}

		switch msg := msg.(type) {
		case *ParameterStatus:
			c.parametersStatus[msg.Name] = msg.Value
		case *BackendKeyData:
			c.pid = msg.PID
			c.key = msg.Key
		case *ReadyForQuery:
			c.status = byte(*msg)
			return nil
		}
	}
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

func (c *Conn) Close() error {
	return nil
}

func validateConfig(cfg map[string]string) error {
	requiredFields := []string{"address", "username", "password", "database"}

	for _, field := range requiredFields {
		if _, ok := cfg[field]; !ok {
			return fmt.Errorf("required field %s is empty", field)
		}
	}

	return nil
}
