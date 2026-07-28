package main

import (
	"fmt"
	"net"
	"github.com/Roman4k-gg/My-Own-Redis/resp"
	"github.com/Roman4k-gg/My-Own-Redis/handler"
	"github.com/Roman4k-gg/My-Own-Redis/storage"
	"io"
)

func main() {
	db := storage.NewStorage()

	cmdHandler := handler.NewHandler(db)
	listener, err := net.Listen("tcp", ":6379")

	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("Server is running on port 6379...")
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn, cmdHandler)
	}
}

func handleConnection(conn net.Conn, h *handler.Handler) {
	defer conn.Close()
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