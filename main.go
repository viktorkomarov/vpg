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

func HexString(needHash string) string {
	m := md5.New()
	m.Write([]byte(needHash))
	return hex.EncodeToString(m.Sum(nil))
}

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

	hash := "md5" + HexString(HexString(password+username)+string(response[9:13]))
	b := make([]byte, 5)
	b[0] = 'p'
	binary.BigEndian.PutUint32(b[1:5], uint32(len(hash)+5))
	b = append(b, hash...)
	b = append(b, '\000')
	_, err = conn.Write(b)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Read(response)
	if err != nil {
		log.Fatal(err)
	}
	typeMessage := string(response[0])

	log.Printf("%s _ %d _ %d\n", typeMessage, binary.BigEndian.Uint32(response[1:5]), binary.BigEndian.Uint32(response[5:9]))

}
