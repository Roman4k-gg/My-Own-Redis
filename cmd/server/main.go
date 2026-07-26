package main

import (
	"fmt"
	"net"
)

func main(){
	listener,err :=net.Listen("tcp",":6379")
	if err != nil{
		panic(err)
	}
	defer listener.Close()
	fmt.Println("Server is running on port :6379")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn){

	defer conn.Close()
    fmt.Println("New client connected!")

}