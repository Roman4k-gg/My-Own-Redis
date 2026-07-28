package resp

import (
	"io"
	"strconv"
)

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) WriteString(s string) error {
    _, err := w.writer.Write([]byte("+" + s + "\r\n"))
    return err
}

func (w *Writer) WriteError(err error) error {
    _, e := w.writer.Write([]byte("-" + err.Error() + "\r\n"))
    return e
}

func (w *Writer) WriteNull() error {
    _, err := w.writer.Write([]byte("$-1\r\n"))
    return err
}

func (w *Writer) WriteBulk(s string) error {
    strLen := strconv.Itoa(len(s))
    w.writer.Write([]byte("$" + strLen + "\r\n"))
    w.writer.Write([]byte(s + "\r\n"))
    
    return nil 
}

func (w *Writer) WriteInt(n int) error {
	str := strconv.Itoa(n)
	_, err := w.writer.Write([]byte(":" + str + "\r\n"))
	return err
}
