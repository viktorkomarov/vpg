package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
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
	IDTable      int32
	NumColumn    int16
	IDDataType   int32
	SizeDataType int16
	ModDataType  int32
	CodeFormat   int16
}

func (r *RowDescription) IsMessage() {}

func NewRowDescription(data []byte) (*RowDescription, error) {
	rows := &RowDescription{}

	rows.Count = int16(binary.BigEndian.Uint16(data[:2]))
	data = data[2:]
	log.Fatalf("%s\n", data)
	return nil, nil
}
