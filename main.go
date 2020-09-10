package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
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

	log.Printf("%s _ %v _ %v\n", typeMessage, binary.BigEndian.Uint32(response[1:5]), binary.BigEndian.Uint32(response[5:9]))

	md := md5.New()
	md5Pass := password + username
	_, err = md.Write([]byte(md5Pass))
	md5Pass = hex.EncodeToString(md.Sum(nil)) + string(response[9:13])
	_, err = md.Write([]byte(md5Pass))
	hash := md.Sum(nil)
	log.Printf("MD5 Hash %v", hash)

	_, err = conn.Write([]byte("md5" + string(hash)))
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Read(response)
	if err != nil {
		log.Fatal(err)
	}
	typeMessage = string(response[0])

	log.Printf("%s _ %s \n", typeMessage, response)

}
