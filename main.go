package main

import "log"

func main() {
	conn, err := New("127.0.0.1:5432")
	if err != nil {
		log.Fatal(err)
	}

	conn.Query("SELECT id, info FROM test")
}
