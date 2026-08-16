package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Roman4k-gg/My-Own-Redis/aof"
	"github.com/Roman4k-gg/My-Own-Redis/handler"
	"github.com/Roman4k-gg/My-Own-Redis/resp"
	"github.com/Roman4k-gg/My-Own-Redis/storage"
)

func main() {
	done := make(chan struct{})

	db := storage.NewStorage()
	db.StartGarbageCollector(done)

	aofFile, err := aof.NewAOF("database.aof")
	if err != nil {
		fmt.Println("Error opening AOF:", err)
		return
	}
	defer aofFile.Close()

	cmdHandler := handler.NewHandler(db, aofFile)

	dummyWriter := resp.NewWriter(io.Discard)
	if err := aofFile.Read(func(value resp.Value) {
		_ = cmdHandler.Handle(value, dummyWriter)
	}); err != nil {
		fmt.Println("Error replaying AOF:", err)
	}

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server is running on port 6379...")

	var wg sync.WaitGroup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		close(done)
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-done:
				wg.Wait()
				fmt.Println("Server stopped.")
				return
			default:
				fmt.Println("Error accepting connection:", err)
				continue
			}
		}

		wg.Add(1)
		go handleConnection(conn, cmdHandler, &wg)
	}
}

func handleConnection(conn net.Conn, h *handler.Handler, wg *sync.WaitGroup) {
	defer conn.Close()
	defer wg.Done()

	fmt.Println("New client connected!")

	r := resp.NewReader(conn)
	w := resp.NewWriter(conn)

	for {
		val, err := r.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected")
				break
			}
			fmt.Println("Error reading from client:", err)
			break
		}
		err = h.Handle(val, w)
		if err != nil {
			fmt.Println("Error handling command:", err)
		}
	}
}
