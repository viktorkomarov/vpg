package main

import (
	"log"
	"net"
	"time"
)

func main() {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:5432", time.Second*5)
	if err != nil {
		log.Fatal(err)
	}

	w := NewBuff()

	//	w.Type(0)
	w.Payload(NewStartUpMessage(map[string]string{
		"user":     "viktor",
		"database": "test",
	}))
	data := w.Message()

	_, err = conn.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	reader := NewReader(conn)
	msg, _ := reader.Receive()
	log.Printf("%T\n", msg)
}
