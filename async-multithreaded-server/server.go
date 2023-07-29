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

type Connection struct {
	conn net.Conn
	msgs chan []byte
}

func NewConnection(connection net.Conn) *Connection {
	return &Connection{
		conn: connection,
		msgs: make(chan []byte, 10),
	}
}

func (c *Connection) readLoop() {

	defer func() {
		close(c.msgs)
		c.conn.Close()
	}()
	buf := make([]byte, 1024)

	for {
		readLength, err := c.conn.Read(buf)
		if err != nil {
			fmt.Println("Error listening while reading: ", err.Error())
			break
		}
		receivedBuff := buf[:readLength]
		c.msgs <- receivedBuff
		rcvMessage := string(receivedBuff)

		fmt.Println("Received message:"+"["+c.conn.RemoteAddr().String()+"]", rcvMessage)
	}

}

func (c *Connection) writeLoop() {

	for echoMsg := range c.msgs {
		c.conn.Write(echoMsg)
	}

}

func main() {

	listener, err := net.Listen(protocol, endpoint)

	if err != nil {
		fmt.Println("Error listening on server: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("Listening on server @", endpoint)

	defer listener.Close()

	for {
		conn, err := listener.Accept()

		connRemoteAddr := conn.RemoteAddr().String()
		fmt.Println("Remote connection:", connRemoteAddr)

		if err != nil {
			fmt.Println("Error listening on connection: ", err.Error())
			continue
		}

		connection := NewConnection(conn)

		go connection.readLoop()
		go connection.writeLoop()
	}

}
