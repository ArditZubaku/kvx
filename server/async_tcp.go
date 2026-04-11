package server

import (
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/ArditZubaku/kvx/config"
	"github.com/ArditZubaku/kvx/core"
)

var connectedClients int = 0

func RunAsyncTCPServer() error {
	log.Println("Starting an asynchronous TCP server on", config.Host, config.Port)

	maxClients := 20_000

	// TODO: Try to implement the kqueue or IOCP ones as well!!!
	// create EPOLL Event Objects to hold events
	events := make([]syscall.EpollEvent, maxClients)

	// create an IPv4 TCP socket in non-blocking mode
	// we get back the file descriptor we want to be monitored by epoll
	serverFd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(serverFd)

	// set the socket to operate in a non-blocking mode
	// TODO: This might not be needed since we already set it via the flag above
	// but let's be safe and set it again just in case, if it's already set it should be a no-op
	if err = syscall.SetNonblock(serverFd, true); err != nil {
		return err
	}

	// bind the IP and the port
	ipv4 := net.ParseIP(config.Host).To4()
	if ipv4 == nil {
		return fmt.Errorf("invalid IPv4 address: %s", config.Host)
	}

	var addr [4]byte
	copy(addr[:], ipv4)

	if err = syscall.Bind(
		serverFd,
		&syscall.SockaddrInet4{
			Port: config.Port,
			Addr: addr,
		},
	); err != nil {
		return err
	}

	// start listening
	if err = syscall.Listen(serverFd, maxClients); err != nil {
		return err
	}

	//----------------------------------------------------------------
	// async IO starts here!!!

	// creating EPOLL instance
	// syscall.EPOLL_CLOEXEC - automatically closes FD on execve()
	epollFd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC) // this is the modern API in comparison to EpollCreate
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(epollFd)

	// notify me when this file descriptor becomes ready for these events (incoming connections)
	serverSocketEvent := syscall.EpollEvent{
		Events: syscall.EPOLLIN, // tell me when serverFd is readable
		Fd:     int32(serverFd),
	}

	// listen to read events on the server itself - registering what we want to be notified uppon
	if err = syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, serverFd, &serverSocketEvent); err != nil {
		return err
	}

	for {
		// see if any fd is ready for IO
		// newEvents is basically the number of events you want to read out of `events`,
		// the new ones that Epoll has put in `events`
		newEvents, err := syscall.EpollWait(epollFd, events, -1)
		if err != nil {
			log.Printf("EpollWait error: %+v\n", err)
			continue
		}

		for i := range newEvents {
			// if the socket server itself ("my server") is ready for IO
			// meaning a new client wants to connect
			if int(events[i].Fd) == serverFd {
				// accept the incoming connection from a client
				// connFd - is the socket between the server and the newly connected client
				connFd, sockAddr, err := syscall.Accept(serverFd)
				if err != nil {
					log.Printf("Accept error: %+v\n", err)
					continue
				}

				log.Printf("Accepted connection from %s", sockAddr)

				// set the client socket non-blocking
				if err := syscall.SetNonblock(connFd, true); err != nil {
					log.Printf("SetNonblock error: %+v\n", err)
					syscall.Close(connFd)
					continue
				}

				// increase the number of concurrent clients count
				connectedClients++

				// add this new TCP connection to be monitored
				clientSocketEvent := syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(connFd),
				}

				// register the client socket event to be notified uppon as well
				if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, connFd, &clientSocketEvent); err != nil {
					log.Fatal(err)
				}
			} else {
				// if some client that is already connected to the server is ready for IO
				// meaning a client wants to send data to the server
				fd := core.FD(events[i].Fd)
				cmd, err := readCommand(fd)
				if err != nil {
					syscall.Close(int(fd))
					connectedClients--
					continue
				}
				respond(cmd, fd)
			}
		}

	}
}
