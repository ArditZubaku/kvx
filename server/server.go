package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/ArditZubaku/kvx/config"
	"github.com/ArditZubaku/kvx/core"
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
			// over the socket, continouesly read the command and print it out
			cmd, err := readCommand(conn)
			if err != nil {
				conn.Close()
				connClients--
				log.Println("Client disconnected", conn.RemoteAddr(), "concurrenct clients running:", connClients)
				if err == io.EOF {
					break
				}
				log.Println("Error:", err)
			}

			log.Println("Command:", cmd)
			if err = respond(cmd, conn); err != nil {
				log.Println("Error on write:", err)
			}
		}
	}
}

func readCommand(conn net.Conn) (*core.RedisCmd, error) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil
}

func respond(cmd string, conn net.Conn) error {
	_, err := conn.Write([]byte(cmd))

	return err
}
