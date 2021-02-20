package main

import "log"

type Test struct {
	R       byte   `pg_header:"R"`
	Payload string `pg_order:"1"`
	Numeric int    `pg_order:"2"`
	Slices  []int  `pg_order:"3" pg_size:"5"`
}

func main() {
	t := Test{Payload: "Payload", Numeric: 124, Slices: []int{1, 2, 3, 4, 5}}
	data, err := Encode(t)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", data)

	var b Test
	err = Decode(data, &b)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", b)
}
