package main

import "log"

func main() {
	_, err := New("127.0.0.1:5432")
	if err != nil {
		log.Fatal(err)
	}
}
