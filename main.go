package main

import "log"

func main() {
	conn, err := New(map[string]string{
		"address":  "127.0.0.1:5432",
		"user":     "viktor",
		"password": "password",
		"database": "test",
	})
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()
}
