package main

import (
	"fmt"
	"net"
	"os"
)

const (
	protocol = "tcp"
	host     = "localhost"
	port     = "8080"
)

const endpoint = host + ":" + port

func main() {

	listener, err := net.Listen(protocol, endpoint)

	if err != nil {
		fmt.Println("Error listening on server: ", err.Error())
		os.Exit(1)
	}

	defer listener.Close()

	fmt.Println("Listening on server @", endpoint)

	for {
		conn, err := listener.Accept()

		connRemoteAddr := conn.RemoteAddr().String()

		fmt.Println("Remote connection:", connRemoteAddr)

		if err != nil {
			fmt.Println("Error listening on connection: ", err.Error())
			continue
		}

		buf := make([]byte, 1024)

		readLength, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error listening while reading: ", err.Error())
			continue
		}

		rcvMessage := string(buf[:readLength])

		fmt.Println("Received message:"+"["+connRemoteAddr+"]", rcvMessage)

		conn.Write([]byte("Message received..."))

		conn.Close()

	}

}
