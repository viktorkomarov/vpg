package main

import (
	"log"
	"reflect"
)

type Test struct {
	R       byte   `pg_header:"R"`
	Payload string `pg_order:"1"`
	Numeric int    `pg_order:"2"`
}

func main() {
	var v byte

	val := reflect.ValueOf(v)
	log.Printf("%s", val.Kind())
}
