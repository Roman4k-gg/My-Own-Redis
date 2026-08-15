package aof

import (
	"github.com/Roman4k-gg/My-Own-Redis/resp"
	"io"
	"os"
	"sync"
)

type AOF struct {
	file *os.File
	mu   sync.Mutex
}

func NewAOF(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &AOF{file: f}, nil
}

func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Close()
}

func (a *AOF) Write(val resp.Value) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	w := resp.NewWriter(a.file)

	if err := w.WriteArray(len(val.Array)); err != nil {
		return err
	}
	for _, arg := range val.Array {
		if err := w.WriteBulk(arg.Str); err != nil {
			return err
		}
	}
	return nil
}

func (a *AOF) Read(callback func(value resp.Value)) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, err := a.file.Seek(0, 0); err != nil {
		return err
	}
	reader := resp.NewReader(a.file)
	for {
		val, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		callback(val)
	}
	return nil
}
