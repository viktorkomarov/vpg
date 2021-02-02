package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Conn struct {
	cfg map[string]string

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

	conn, err := net.DialTimeout("tcp", cfg["address"], time.Second*5)
	if err != nil {
		return nil, fmt.Errorf("can't open connection %w", err)
	}

	c := &Conn{
		conn:             conn,
		cfg:              cfg,
		reader:           newReader(conn),
		writer:           newWriter(conn),
		parametersStatus: make(map[string]string),
	}

	if err = c.init(); err != nil {
		c.conn.Close()
		return nil, err
	}

	return c, nil
}

type authorizationer interface {
	authorize(writer *Writer, reader *Reader) error
}

func (c *Conn) init() error {
	if err := c.writer.sendMsg(newStartUpMessage(c.cfg)); err != nil {
		return err
	}

	log.Printf("send start up msg")
	classificator, err := c.receiveAuthClassificator()
	if err != nil {
		return err
	}
	log.Printf("recieve %+v", classificator)

	client := authClient(classificator, c.cfg["user"], c.cfg["password"])
	if err = client.authorize(c.writer, c.reader); err != nil {
		return err
	}

	log.Fatal("AUTHORIZED")
	for { // enrich conn
		msg, err := c.reader.receive()
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
		default:
			//
		}
	}
}

func (c *Conn) receiveAuthClassificator() (auth, error) {
	msg, err := c.reader.receive()
	if err != nil {
		return auth{}, err
	}

	if c, ok := msg.(auth); ok {
		return c, nil
	}

	return auth{}, fmt.Errorf("unknown msg %+v", msg)
}

func (c *Conn) Close() error {
	return nil
}

func validateConfig(cfg map[string]string) error {
	requiredFields := []string{"address", "user", "password", "database"}

	for _, field := range requiredFields {
		if _, ok := cfg[field]; !ok {
			return fmt.Errorf("required field %s is empty", field)
		}
	}

	return nil
}
