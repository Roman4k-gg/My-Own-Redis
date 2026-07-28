package handler

import (
	"fmt"
	"strings"
	"github.com/Roman4k-gg/My-Own-Redis/resp"
	"github.com/Roman4k-gg/My-Own-Redis/storage"
)

type Handler struct {
	store *storage.Storage
}

func NewHandler(store *storage.Storage) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Handle(val resp.Value, w *resp.Writer) error {
	if val.Typ != "array" || len(val.Array) == 0 {
		return w.WriteError(fmt.Errorf("invalid request format"))
	}

	command := strings.ToUpper(val.Array[0].Str)
	args := val.Array[1:]

	switch command {
	case "PING":
		if len(args) == 0 {
			return w.WriteString("PONG")
		}
		return w.WriteBulk(args[0].Str)
	case "ECHO":
		if len(args) != 1 {
			return w.WriteError(fmt.Errorf("ERR wrong number of arguments for 'echo' command"))
		}
		return w.WriteBulk(args[0].Str)

	case "SET":
		if len(args) != 2 {
			return w.WriteError(fmt.Errorf("ERR wrong number of arguments for 'set' command"))
		}
		h.store.Set(args[0].Str, args[1].Str)
		return w.WriteString("OK")

	case "GET":
		if len(args) != 1 {
			return w.WriteError(fmt.Errorf("ERR wrong number of arguments for 'get' command"))
		}
		value, ok := h.store.Get(args[0].Str)
		if !ok {
			return w.WriteNull()
		}
		return w.WriteBulk(value)
	case "DEL":
		if len(args) == 0 {
			return w.WriteError(fmt.Errorf("ERR wrong number of arguments for 'del' command"))
		}
		keys := make([]string, len(args))
		for i, arg := range args {
			keys[i] = arg.Str
		}
		deletedCount := h.store.Delete(keys)
		return w.WriteInt(deletedCount)

	case "EXISTS":
		if len(args) == 0 {
			return w.WriteError(fmt.Errorf("ERR wrong number of arguments for 'exists' command"))
		}
		keys := make([]string, len(args))
		for i, arg := range args {
			keys[i] = arg.Str
		}
		existCount := h.store.Exists(keys)
		return w.WriteInt(existCount)


	default:
		return w.WriteError(fmt.Errorf("ERR unknown command '%s'", command))
	}
}
