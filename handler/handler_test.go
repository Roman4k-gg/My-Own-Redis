package handler

import (
	"bytes"
	"os"
	"testing"

	"github.com/Roman4k-gg/My-Own-Redis/aof"
	"github.com/Roman4k-gg/My-Own-Redis/resp"
	"github.com/Roman4k-gg/My-Own-Redis/storage"
)

func newTestHandler(t *testing.T) (*Handler, func()) {
	t.Helper()
	store := storage.NewStorage()
	f, err := os.CreateTemp("", "handler-test-*.aof")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	f.Close()
	aofFile, err := aof.NewAOF(name)
	if err != nil {
		t.Fatal(err)
	}
	return NewHandler(store, aofFile), func() {
		aofFile.Close()
		os.Remove(name)
	}
}

func cmd(name string, args ...string) resp.Value {
	arr := make([]resp.Value, 0, 1+len(args))
	arr = append(arr, resp.Value{Str: name})
	for _, a := range args {
		arr = append(arr, resp.Value{Str: a})
	}
	return resp.Value{Typ: "array", Array: arr}
}

func run(h *Handler, v resp.Value) string {
	var buf bytes.Buffer
	w := resp.NewWriter(&buf)
	_ = h.Handle(v, w)
	return buf.String()
}

func TestPingNoArgs(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("PING"))
	if got != "+PONG\r\n" {
		t.Fatalf("expected +PONG, got %q", got)
	}
}

func TestPingWithMessage(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("PING", "hello"))
	if got != "$5\r\nhello\r\n" {
		t.Fatalf("expected bulk hello, got %q", got)
	}
}

func TestEcho(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("ECHO", "world"))
	if got != "$5\r\nworld\r\n" {
		t.Fatalf("expected bulk world, got %q", got)
	}
}

func TestEchoWrongArgs(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("ECHO"))
	if got[0] != '-' {
		t.Fatalf("expected error, got %q", got)
	}
}

func TestSetAndGet(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	if got := run(h, cmd("SET", "k", "v")); got != "+OK\r\n" {
		t.Fatalf("expected +OK, got %q", got)
	}
	if got := run(h, cmd("GET", "k")); got != "$1\r\nv\r\n" {
		t.Fatalf("expected bulk v, got %q", got)
	}
}

func TestGetMissingKey(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("GET", "missing"))
	if got != "$-1\r\n" {
		t.Fatalf("expected null bulk, got %q", got)
	}
}

func TestSetWithTTL(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	if got := run(h, cmd("SET", "k", "v", "EX", "100")); got != "+OK\r\n" {
		t.Fatalf("expected +OK, got %q", got)
	}
}

func TestSetInvalidOption(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("SET", "k", "v", "PX", "100"))
	if got[0] != '-' {
		t.Fatalf("expected error, got %q", got)
	}
}

func TestSetInvalidTTL(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("SET", "k", "v", "EX", "notanumber"))
	if got[0] != '-' {
		t.Fatalf("expected error, got %q", got)
	}
}

func TestDel(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	run(h, cmd("SET", "a", "1"))
	run(h, cmd("SET", "b", "2"))

	got := run(h, cmd("DEL", "a", "b", "missing"))
	if got != ":2\r\n" {
		t.Fatalf("expected :2, got %q", got)
	}
}

func TestDelNoArgs(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("DEL"))
	if got[0] != '-' {
		t.Fatalf("expected error, got %q", got)
	}
}

func TestExists(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	run(h, cmd("SET", "x", "1"))

	got := run(h, cmd("EXISTS", "x", "missing"))
	if got != ":1\r\n" {
		t.Fatalf("expected :1, got %q", got)
	}
}

func TestExistsNoArgs(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("EXISTS"))
	if got[0] != '-' {
		t.Fatalf("expected error, got %q", got)
	}
}

func TestIncr(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("INCR", "counter"))
	if got != ":1\r\n" {
		t.Fatalf("expected :1, got %q", got)
	}
	got = run(h, cmd("INCR", "counter"))
	if got != ":2\r\n" {
		t.Fatalf("expected :2, got %q", got)
	}
}

func TestIncrOnNonInteger(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	run(h, cmd("SET", "k", "not_a_number"))
	got := run(h, cmd("INCR", "k"))
	if got[0] != '-' {
		t.Fatalf("expected error, got %q", got)
	}
}

func TestIncrWrongArgs(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("INCR"))
	if got[0] != '-' {
		t.Fatalf("expected error, got %q", got)
	}
}

func TestUnknownCommand(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("FLUSHALL"))
	if got[0] != '-' {
		t.Fatalf("expected error for unknown command, got %q", got)
	}
}

func TestInvalidRequestFormat(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	var buf bytes.Buffer
	w := resp.NewWriter(&buf)
	_ = h.Handle(resp.Value{Typ: "bulk", Str: "hello"}, w)
	if buf.String()[0] != '-' {
		t.Fatalf("expected error for non-array type, got %q", buf.String())
	}
}

func TestCaseInsensitiveCommands(t *testing.T) {
	h, cleanup := newTestHandler(t)
	defer cleanup()

	got := run(h, cmd("ping"))
	if got != "+PONG\r\n" {
		t.Fatalf("expected +PONG for lowercase ping, got %q", got)
	}
}
