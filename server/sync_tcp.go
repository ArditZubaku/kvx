package server

import (
	"errors"
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
			cmds, readErr := readCommands(conn)
			if readErr != nil {
				if err := conn.Close(); err != nil {
					log.Printf("failed to close connection: %v\n", err)
				}
				connClients--
				log.Println("Client disconnected", conn.RemoteAddr(), "concurrent clients running:", connClients)
				if errors.Is(readErr, io.EOF) {
					break
				}
				log.Println("Error:", readErr)
				break
			}

			respond(cmds, conn)
		}
	}
}
