package aof

import (
	"os"
	"testing"

	"github.com/Roman4k-gg/My-Own-Redis/resp"
)

func newTestAOF(t *testing.T) (*AOF, func()) {
	t.Helper()
	f, err := os.CreateTemp("", "aof-test-*.aof")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	f.Close()
	a, err := NewAOF(name)
	if err != nil {
		t.Fatal(err)
	}
	return a, func() {
		a.Close()
		os.Remove(name)
	}
}

func makeCmd(name string, args ...string) resp.Value {
	arr := make([]resp.Value, 0, 1+len(args))
	arr = append(arr, resp.Value{Str: name})
	for _, a := range args {
		arr = append(arr, resp.Value{Str: a})
	}
	return resp.Value{Typ: "array", Array: arr}
}

func TestAOFWriteAndRead(t *testing.T) {
	a, cleanup := newTestAOF(t)
	defer cleanup()

	cmds := []resp.Value{
		makeCmd("SET", "key1", "val1"),
		makeCmd("SET", "key2", "val2"),
		makeCmd("DEL", "key1"),
	}

	for _, c := range cmds {
		if err := a.Write(c); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	var replayed []resp.Value
	if err := a.Read(func(v resp.Value) {
		replayed = append(replayed, v)
	}); err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(replayed) != len(cmds) {
		t.Fatalf("expected %d commands, got %d", len(cmds), len(replayed))
	}

	for i, v := range replayed {
		if v.Typ != "array" {
			t.Fatalf("cmd %d: expected array, got %q", i, v.Typ)
		}
		if len(v.Array) != len(cmds[i].Array) {
			t.Fatalf("cmd %d: expected %d args, got %d", i, len(cmds[i].Array), len(v.Array))
		}
		for j, arg := range v.Array {
			if arg.Str != cmds[i].Array[j].Str {
				t.Fatalf("cmd %d arg %d: expected %q, got %q", i, j, cmds[i].Array[j].Str, arg.Str)
			}
		}
	}
}

func TestAOFReadEmpty(t *testing.T) {
	a, cleanup := newTestAOF(t)
	defer cleanup()

	var count int
	if err := a.Read(func(v resp.Value) { count++ }); err != nil {
		t.Fatalf("unexpected error reading empty AOF: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 commands from empty file, got %d", count)
	}
}

func TestAOFPersistsAcrossReopen(t *testing.T) {
	f, err := os.CreateTemp("", "aof-reopen-*.aof")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	f.Close()
	defer os.Remove(name)

	a, err := NewAOF(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Write(makeCmd("SET", "x", "42")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	a.Close()

	a2, err := NewAOF(name)
	if err != nil {
		t.Fatal(err)
	}
	defer a2.Close()

	var replayed []resp.Value
	if err := a2.Read(func(v resp.Value) { replayed = append(replayed, v) }); err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(replayed) != 1 {
		t.Fatalf("expected 1 command after reopen, got %d", len(replayed))
	}
	if replayed[0].Array[0].Str != "SET" {
		t.Fatalf("expected SET command, got %q", replayed[0].Array[0].Str)
	}
}
