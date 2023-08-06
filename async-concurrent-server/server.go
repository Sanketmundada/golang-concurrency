package main

import (
	"fmt"
	"log"
	"net"
	"syscall"
)

const (
	maxClients = 20000
	host       = "127.0.0.1"
	port       = 8080
)

func main() {
	events := make([]syscall.EpollEvent, maxClients)

	// Create socket
	socketfd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(socketfd)

	// Bind socket
	ip := net.ParseIP(host)
	err = syscall.Bind(socketfd, &syscall.SockaddrInet4{
		Port: port,
		Addr: [4]byte{ip[0], ip[1], ip[2], ip[3]},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Listen on socket
	if err = syscall.Listen(socketfd, maxClients); err != nil {
		log.Fatal(err)
	}

	// Async IO
	// Create an Epoll instance
	epollfd, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(epollfd)

	socketServerEvent := syscall.EpollEvent{
		Fd:     int32(socketfd),
		Events: syscall.EPOLLIN,
	}

	if err = syscall.EpollCtl(epollfd, syscall.EPOLL_CTL_ADD, socketfd, &socketServerEvent); err != nil {
		log.Fatal(err)
	}

	log.Println("Listing on port: ", port)

	for {
		nEvents, err := syscall.EpollWait(epollfd, events[:], -1)
		if err != nil {
			log.Println(err)
			continue
		}

		for i := 0; i < nEvents; i++ {
			currentEvent := events[i]

			if currentEvent.Fd == int32(socketfd) {
				// Accept incoming connection
				fd, sockAddr, err := syscall.Accept(socketfd)
				if err != nil {
					log.Println(err)
					continue
				}

				if err = syscall.SetNonblock(fd, true); err != nil {
					log.Println(err)
					continue
				}

				fmt.Println("New connection: ", sockAddr)

				// Add the new client connection to epoll instance
				clientReadEvent := syscall.EpollEvent{
					Fd:     int32(fd),
					Events: syscall.EPOLLIN, // Adding this to listen for read event on socket
				}

				if err = syscall.EpollCtl(epollfd, syscall.EPOLL_CTL_ADD, fd, &clientReadEvent); err != nil {
					log.Fatal(err)
				}
			} else {
				var msg []byte = make([]byte, 1024)
				nRBytes, err := syscall.Read(int(currentEvent.Fd), msg)
				log.Println("Bytes read: ", nRBytes)

				// Check if client disconnected or some other error
				if nRBytes == 0 || err != nil {
					log.Println("Disconnecting client while reading")
					syscall.Close(int(currentEvent.Fd))
					continue
				}

				log.Println("Message: ", string(msg[:nRBytes]))

				if _, err := syscall.Write(int(currentEvent.Fd), msg); err != nil {
					log.Println("Disconnecting client while writing")
					syscall.Close(int(currentEvent.Fd))
					continue
				}

			}
		}
	}
}
