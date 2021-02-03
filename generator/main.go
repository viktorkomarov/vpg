package main

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type test struct {
	a int    `pg_type:"1,5"`
	b string `pg_type:"6,7"`
}

func main() {
	var t test

	v := reflect.ValueOf(t)
	typeOfT := reflect.TypeOf(t)

	seril := make(map[[2]int]reflect.Value)
	for i := 0; i < v.NumField(); i++ {
		name := typeOfT.Field(i).Tag.Get("pg_type")
		if name == "" {
			continue
		}

		var (
			key [2]int
			err error
		)
		limits := strings.Split(name, ",")
		if len(limits) != 2 {
			log.Fatal("len(limits) !=2 ")
		}

		key[0], err = strconv.Atoi(limits[0])
		if err != nil {
			log.Fatal(err)
		}

		key[1], err = strconv.Atoi(limits[1])
		if err != nil {
			log.Fatal(err)
		}

		seril[key] = v.Field(i)
	}

	for key, val := range seril {
		fmt.Printf("%d %d\n", key[0], key[1])
		fmt.Printf("%s\n", val.Type().Name())
	}
}
