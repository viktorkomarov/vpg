package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

/*
Сервер может передавать в этой фазе следующие ответные сообщения:

CommandComplete (Команда завершена)
Команда SQL выполнена нормально.

RowDescription (Описание строк)
Показывает, что в ответ на запрос SELECT, FETCH и т. п. будут возвращены строки. В содержимом этого сообщения описывается структура столбцов этих строк. За ним для каждой строки, возвращаемой клиенту, следует сообщение DataRow.

DataRow (Строка данных)
Одна строка из набора, возвращаемого запросом SELECT, FETCH и т. п.

EmptyQueryResponse (Ответ на пустой запрос)
Была принята пустая строка запроса.

ErrorResponse (Ошибочный ответ)
Произошла ошибка.

ReadyForQuery (Готов к запросам)
Обработка строки запроса завершена. Чтобы отметить это, отправляется отдельное сообщение, так как строка запроса может содержать несколько команд SQL. (Сообщение CommandComplete говорит о завершении обработки одной команды SQL, а не всей строки.) ReadyForQuery передаётся всегда, и при успешном завершении обработки, и при ошибке.

NoticeResponse (Ответ с замечанием)
Выдаётся предупреждение, связанное с запросом. Эти замечания дополняют другие ответы, то есть сервер, выдавая их, продолжает обрабатывать команду.
*/
var validCmds = map[string]bool{
	"INSERT": true,
	"DELETE": true,
	"UPDATE": true,
	"SELECT": true,
	"MOVE":   true,
	"FETCH":  true,
	"COPY":   true,
}

type CommandComplete struct {
	Command string
	Count   int
	OID     int //only for INSERT
}

func (c *CommandComplete) IsMessage() {}

func NewCommandComplete(data []byte) (*CommandComplete, error) {
	buf := bytes.Split(data, []byte(" "))
	if len(buf) < 2 {
		return nil, fmt.Errorf("uncorrect command complete len %s", data)
	}
	c := &CommandComplete{}
	cmd := string(buf[0])

	if !validCmds[cmd] {
		return nil, fmt.Errorf("unknown CommandComplete tag %s", cmd)
	}

	c.Command = cmd
	count := buf[2]
	if cmd == "INSERT" {
		if len(buf) != 3 {
			return nil, fmt.Errorf("uncorrect command complete len %s", data)
		}

		c.OID = int(binary.BigEndian.Uint32(buf[2]))
		count = buf[3]
	}
	c.Count = int(binary.BigEndian.Uint32(count))

	return c, nil
}

type RowDescription struct {
	Count        int16
	Descriptions []Description
}

type Description struct {
	Name         string
	IDTable      uint32
	NumColumn    uint16
	IDDataType   uint32
	SizeDataType uint16
	ModDataType  uint32
	CodeFormat   uint16
}

func (r *RowDescription) isMessage() {}

func NewRowDescription(data []byte) (*RowDescription, error) {
	rows := &RowDescription{}

	rows.Count = int16(binary.BigEndian.Uint16(data[:2]))
	rows.Descriptions = make([]Description, 0, rows.Count)
	data = data[2:]
	for i := 0; int16(i) < rows.Count; i++ {
		var desc Description
		idx := bytes.Index(data, []byte{'\000'})
		if idx == -1 {
			return nil, fmt.Errorf("uncorrect rows description %+v", data)
		}

		desc.Name = string(data[:idx])
		data = data[idx+1:] //TODO::errcheck
		idx++
		desc.IDTable = binary.BigEndian.Uint32(data[idx : idx+4])
		idx += 4
		desc.NumColumn = binary.BigEndian.Uint16(data[idx : idx+2])
		idx += 2
		desc.IDDataType = binary.BigEndian.Uint32(data[idx : idx+4])
		idx += 4
		desc.SizeDataType = binary.BigEndian.Uint16(data[idx : idx+2])
		idx += 2
		desc.ModDataType = binary.BigEndian.Uint32(data[idx : idx+4])
		idx += 4
		desc.CodeFormat = binary.BigEndian.Uint16(data[idx : idx+2])
		data = data[idx:]
	}
	return rows, nil
}

type DataRows struct {
	rows [][]byte
}

func (d *DataRows) IsMessage() {}

func NewDataRow(payload []byte) (*DataRows, error) {
	return nil, nil
}
