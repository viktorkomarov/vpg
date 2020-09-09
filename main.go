package main

import (
	"encoding/binary"
	"log"
	"net"
)

const (
	username = "viktor"
	password = "password"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:5432")
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 8, 256)
	buf = append(buf, "user"...)
	buf = append(buf, '\000')
	buf = append(buf, username...)
	buf = append(buf, '\000')
	binary.BigEndian.PutUint32(buf[4:8], uint32(196608))
	buf = append(buf, '\000')
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(buf)))
	_, err = conn.Write(buf)
	if err != nil {
		log.Fatal("can't send ", err)
	}

	response := make([]byte, 128)
	_, err = conn.Read(response)
	if err != nil {
		log.Fatal("can't read ", err)
	}
	typeMessage := string(response[0])
	responseLen, _ := binary.Uvarint(response[1:5])

	log.Printf("%s _ %d\n", typeMessage, responseLen)
}
