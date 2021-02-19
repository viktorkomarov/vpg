package main

import "log"

type Test struct {
	R       byte   `pg_header:"R"`
	Payload string `pg_order:"1"`
}

func main() {
	t := Test{Payload: "payload"}

	data, err := Encode(t)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", data)
	var d Test
	err = Decode(data, &d)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v\n", d.Payload)
}
