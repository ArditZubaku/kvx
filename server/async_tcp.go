package server

import (
	"fmt"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/ArditZubaku/kvx/config"
	"github.com/ArditZubaku/kvx/core"
)

var connectedClients = 0

// TODO: Redis does it 10x a second, maybe make this configurable.
var (
	cronFrequency    = 1 * time.Second
	lastCronExecTime = time.Now()
)

func RunAsyncTCPServer() error {
	log.Println("Starting an asynchronous TCP server on", config.Host, config.Port)

	maxClients := 20_000

	// TODO: Try to implement the kqueue or IOCP ones as well.
	// create EPOLL Event Objects to hold events
	events := make([]syscall.EpollEvent, maxClients)

	// create an IPv4 TCP socket in non-blocking mode
	// we get back the file descriptor we want to be monitored by epoll
	serverFd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	defer func(fd int) {
		closeErr := syscall.Close(fd)
		if closeErr != nil {
			log.Fatalf("failed to close server FD %d: %v", fd, err)
		}
	}(serverFd)

	// set the socket to operate in a non-blocking mode
	// TODO: This might not be needed since we already set it via the flag above
	// but let's be safe and set it again just in case, if it's already set it should be a no-op
	if nbErr := syscall.SetNonblock(serverFd, true); nbErr != nil {
		return nbErr
	}

	// bind the IP and the port
	ipv4 := net.ParseIP(config.Host).To4()
	if ipv4 == nil {
		return fmt.Errorf("invalid IPv4 address: %s", config.Host)
	}

	var addr [4]byte
	copy(addr[:], ipv4)

	if bindErr := syscall.Bind(
		serverFd,
		&syscall.SockaddrInet4{
			Port: config.Port,
			Addr: addr,
		},
	); bindErr != nil {
		return bindErr
	}

	// start listening
	if lErr := syscall.Listen(serverFd, maxClients); lErr != nil {
		return lErr
	}

	//----------------------------------------------------------------
	// async IO starts here!!!

	// creating EPOLL instance
	// syscall.EPOLL_CLOEXEC - automatically closes FD on execve()
	epollFd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC) // this is the modern API in comparison to EpollCreate
	if err != nil {
		log.Printf("syscall error - epollCreate1, %+v", err)
	}
	defer func(fd int) {
		closeErr := syscall.Close(fd)
		if closeErr != nil {
			log.Fatalf("failed to close epoll FD %d: %v", fd, err)
		}
	}(epollFd)

	// notify me when this file descriptor becomes ready for these events (incoming connections)
	serverSocketEvent := syscall.EpollEvent{
		Events: syscall.EPOLLIN, // tell me when serverFd is readable
		Fd:     int32(serverFd),
	}

	// listen to read events on the server itself - registering what we want to be notified uppon
	if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, serverFd, &serverSocketEvent); err != nil {
		return err
	}

	for {
		// TODO: Think about the case when epoll is blocked and the cron needs to run?
		if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
			core.DeleteExpiredKeys()

			lastCronExecTime = time.Now()
		}

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

				if addr, ok := sockAddr.(*syscall.SockaddrInet4); ok {
					log.Printf("Accepted connection from %s:%d", net.IP(addr.Addr[:]).String(), addr.Port)
				} else {
					log.Printf("Accepted connection from %+v", sockAddr)
				}

				// set the client socket non-blocking
				if err := syscall.SetNonblock(connFd, true); err != nil {
					log.Printf("SetNonblock error: %+v\n", err)

					if closeErr := syscall.Close(connFd); closeErr != nil {
						log.Printf("failed to close conn FD %d: %v\n", connFd, closeErr)
					}

					continue
				}

				// increase the number of concurrent clients count
				connectedClients++
				log.Println("Current connected clients - ", connectedClients)

				// add this new TCP connection to be monitored
				clientSocketEvent := syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(connFd),
				}

				// register the client socket event to be notified uppon as well
				if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, connFd, &clientSocketEvent); err != nil {
					log.Printf("syscall error - epollCtl, %+v", err)
				}
			} else {
				// if some client that is already connected to the server is ready for IO
				// meaning a client wants to send data to the server
				fd := core.FD(events[i].Fd)

				cmds, err := readCommands(fd)
				if err != nil {
					if closeErr := syscall.Close(int(fd)); closeErr != nil {
						log.Printf("failed to close conn FD %d: %v\n", fd, closeErr)
					}

					connectedClients--
					log.Println("Current connected clients - ", connectedClients)

					continue
				}

				respond(cmds, fd)
			}
		}
	}
}
