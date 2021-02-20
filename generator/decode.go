package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"sort"
)

func Decode(data []byte, v interface{}) error {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr || ptr.IsNil() {
		return errors.New("not nil ptr is required")
	}

	fields, err := analyzeFields(v)
	if err != nil {
		return err
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].order < fields[j].order
	})

	header := data[0]
	data = data[5:] // skip header and size

	if header != fields[0].header {
		// handler postgres err and other async msg
		return errors.New("mismatch header type")
	}

	fields = fields[1:]
	for _, field := range fields {
		decode, err := decodeByType(field.typ)
		if err != nil {
			return err
		}

		offset, err := decode(data, field)
		if err != nil {
			return err
		}

		data = data[offset:]
	}

	return nil
}

func decodeByType(v reflect.Type) (func([]byte, Field) (int, error), error) {
	switch v.Kind() {
	case reflect.String:
		return decodeString, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return decodeInt, nil
	case reflect.Slice:
		return decodeSlice, nil
	default:
		return nil, fmt.Errorf("unsupported decode type %+v", v.Kind())
	}
}

func decodeString(data []byte, v Field) (int, error) {
	offset := elementaryOffset(data, v.typ)

	str := data[:offset]
	if v.val.CanSet() {
		v.val.SetString(string(str))
		return offset, nil
	}

	return 0, fmt.Errorf("unsettable field %+v", v)
}

func decodeInt(data []byte, v Field) (int, error) {
	offset := elementaryOffset(data, v.typ)

	if v.val.CanSet() {
		num := int64(binary.BigEndian.Uint32(data)) // change it can be uint16
		v.val.SetInt(num)
		return offset, nil
	}

	return 0, fmt.Errorf("unsettable field %+v", v)
}

func decodeSlice(data []byte, field Field) (int, error) {
	offset := 0
	if field.size == -1 {
		field.size = int(binary.BigEndian.Uint32(data))
		offset += 4
	}

	sl := reflect.MakeSlice(field.typ, field.size, field.size)
	for i := 0; i < field.size; i++ {
		val := sl.Index(i)

		decode, err := decodeByType(field.typ.Elem())
		if err != nil {
			return 0, err
		}

		elemOffset, err := decode(data[offset:], Field{val: val, typ: field.typ.Elem()})
		if err != nil {
			return 0, err
		}

		offset += elemOffset
	}

	if !field.val.CanSet() {
		return 0, fmt.Errorf("unsettable field %+v", field)
	}

	field.val.Set(sl)
	return offset, nil
}

func elementaryOffset(data []byte, typ reflect.Type) int {
	switch typ.Kind() {
	case reflect.String:
		return bytes.IndexByte(data, '\000') + 1
	case reflect.Int16:
		return 2
	case reflect.Int32, reflect.Int64, reflect.Int:
		return 4
	case reflect.Uint8:
		return 1
	default:
		return 0
	}
}
