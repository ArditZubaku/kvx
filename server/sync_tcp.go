package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/ArditZubaku/kvx/config"
)

func RunSyncTCPServer() {
	log.Println("Starting a synchronous TCP server on", config.Host, config.Port)

	connClients := 0

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.Host, strconv.Itoa(config.Port)))
	if err != nil {
		panic(err)
	}

	for {
		// blocking call - waiting for the new clients to connect
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		connClients++
		log.Println("Client connected with address:", conn.RemoteAddr(), "concurrent clients running:", connClients)

		for {
			// over the socket, continuously read the command and print it out
			cmd, err := readCommand(conn)
			if err != nil {
				err := conn.Close()
				if err != nil {
					log.Printf("failed to close connection: %v\n", err)
					return
				}
				connClients--
				log.Println("Client disconnected", conn.RemoteAddr(), "concurrent clients running:", connClients)
				if err == io.EOF {
					break
				}
				log.Println("Error:", err)
			}

			respond(cmd, conn)
		}
	}
}
