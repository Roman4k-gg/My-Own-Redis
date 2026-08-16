package resp

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestWriterWriteString(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	_ = w.WriteString("PONG")
	if got := buf.String(); got != "+PONG\r\n" {
		t.Fatalf("expected '+PONG\\r\\n', got %q", got)
	}
}

func TestWriterWriteError(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	_ = w.WriteError(fmt.Errorf("ERR bad"))
	if got := buf.String(); got != "-ERR bad\r\n" {
		t.Fatalf("expected '-ERR bad\\r\\n', got %q", got)
	}
}

func TestWriterWriteNull(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	_ = w.WriteNull()
	if got := buf.String(); got != "$-1\r\n" {
		t.Fatalf("expected '$-1\\r\\n', got %q", got)
	}
}

func TestWriterWriteBulk(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	_ = w.WriteBulk("hello")
	if got := buf.String(); got != "$5\r\nhello\r\n" {
		t.Fatalf("expected '$5\\r\\nhello\\r\\n', got %q", got)
	}
}

func TestWriterWriteInt(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	_ = w.WriteInt(42)
	if got := buf.String(); got != ":42\r\n" {
		t.Fatalf("expected ':42\\r\\n', got %q", got)
	}
}

func TestWriterWriteArray(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	_ = w.WriteArray(3)
	if got := buf.String(); got != "*3\r\n" {
		t.Fatalf("expected '*3\\r\\n', got %q", got)
	}
}

func TestReaderRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	_ = w.WriteArray(3)
	_ = w.WriteBulk("SET")
	_ = w.WriteBulk("key")
	_ = w.WriteBulk("value")

	r := NewReader(&buf)
	val, err := r.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Typ != "array" {
		t.Fatalf("expected array, got %q", val.Typ)
	}
	if len(val.Array) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(val.Array))
	}
	if val.Array[0].Str != "SET" {
		t.Fatalf("expected SET, got %q", val.Array[0].Str)
	}
	if val.Array[1].Str != "key" {
		t.Fatalf("expected key, got %q", val.Array[1].Str)
	}
	if val.Array[2].Str != "value" {
		t.Fatalf("expected value, got %q", val.Array[2].Str)
	}
}

func TestReaderBulkString(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	_ = w.WriteBulk("hello")

	r := NewReader(&buf)
	val, err := r.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Typ != "bulk" {
		t.Fatalf("expected bulk, got %q", val.Typ)
	}
	if val.Str != "hello" {
		t.Fatalf("expected hello, got %q", val.Str)
	}
}

func TestReaderEmptyArray(t *testing.T) {
	input := "*0\r\n"
	r := NewReader(strings.NewReader(input))
	val, err := r.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Typ != "array" {
		t.Fatalf("expected array, got %q", val.Typ)
	}
	if len(val.Array) != 0 {
		t.Fatalf("expected empty array, got %d elements", len(val.Array))
	}
}

func TestReaderUnknownType(t *testing.T) {
	r := NewReader(strings.NewReader("!invalid\r\n"))
	_, err := r.Read()
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}
