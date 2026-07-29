package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Value struct {
	Typ   string
	Str   string
	Num   int
	Array []Value
}

type Reader struct {
	reader *bufio.Reader
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{reader: bufio.NewReader(rd)}
}

func (r *Reader) readLine() (line []byte, err error) {
	b, err := r.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return b[:len(b)-2], nil
}

func (r *Reader) Read() (Value, error) {
	b, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch b {
	case '*':
		return r.readArray()
	case '$':
		return r.readBulk()
	default:
		fmt.Printf("unknown type :%v", string(b))
		return Value{}, nil
	}
}

func (r *Reader) readArray() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	count, _ := strconv.Atoi(string(line))

	var array []Value

	for i := 0; i < count; i++ {
		val, err := r.Read()
		if err != nil {
			return Value{}, err
		}

		array = append(array, val)
	}

	return Value{
		Typ:   "array",
		Array: array,
	}, nil
}

func (r *Reader) readBulk() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	count, _ := strconv.Atoi(string(line))

	buf := make([]byte, count)

	_, err = io.ReadFull(r.reader, buf)
	if err != nil {
		return Value{}, err
	}

	r.readLine()

	return Value{
		Typ: "bulk",
		Str: string(buf),
	}, nil
}
