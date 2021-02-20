package encoder

import (
	"fmt"
	"reflect"
	"strconv"
)

type Field struct {
	header  byte
	order   int
	preffix string
	size    int
	val     reflect.Value
	typ     reflect.Type
}

func analyzeFields(v interface{}) ([]Field, error) {
	strct := reflect.Indirect(reflect.ValueOf(v))
	if strct.Kind() != reflect.Struct {
		return nil, fmt.Errorf("support only struct not %v", strct.Kind())
	}

	fields := make([]Field, 0)
	for i := 0; i < strct.NumField(); i++ {
		field := strct.Type().Field(i)

		header, ok := field.Tag.Lookup("pg_header")
		if ok {
			fields = append(fields, Field{
				order:  0,
				header: header[0],
			})

			continue
		}

		pg, ok := field.Tag.Lookup("pg_order")
		if !ok {
			continue
		}

		preffix, _ := field.Tag.Lookup("pg_preffix")
		order, err := strconv.Atoi(pg)
		if err != nil {
			return nil, fmt.Errorf("incorrect order %s", pg)
		}

		realSize := 0
		size, ok := field.Tag.Lookup("pg_size")
		if ok {
			realSize, err = strconv.Atoi(size)
			if err != nil {
				return nil, fmt.Errorf("incorrect order")
			}
		}

		fields = append(fields, Field{
			order:   order,
			preffix: preffix,
			val:     strct.Field(i),
			typ:     field.Type,
			size:    realSize,
		})
	}

	return fields, nil
}

func checkSortedOrder(fields []Field) error {
	uniq := make(map[int]bool)

	orders := 1
	for i := range fields {
		if fields[i].header != 0 {
			uniq[0] = true
			continue
		}

		if uniq[fields[i].order] {
			return fmt.Errorf("duplicate %d", fields[i].order)
		}

		if orders != fields[i].order {
			return fmt.Errorf("broken order %d", fields[i].order)
		}

		uniq[fields[i].order] = true
		orders++
	}

	return nil
}
