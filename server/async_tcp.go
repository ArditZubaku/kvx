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

	serverFd, err := setupTCPServer(maxClients)
	if err != nil {
		return err
	}
	defer func(fd int) {
		if closeErr := syscall.Close(fd); closeErr != nil {
			log.Fatalf("failed to close server FD %d: %v", fd, closeErr)
		}
	}(serverFd)

	// async IO starts here
	epollFd, err := setupEpoll(serverFd)
	if err != nil {
		return err
	}
	defer func(fd int) {
		if closeErr := syscall.Close(fd); closeErr != nil {
			log.Fatalf("failed to close epoll FD %d: %v", fd, closeErr)
		}
	}(epollFd)

	for {
		// TODO: Think about the case when epoll is blocked and the cron needs to run?
		if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
			core.DeleteExpiredKeys()
			lastCronExecTime = time.Now()
		}

		// newEvents is the count of events Epoll has placed in the slice
		newEvents, err := syscall.EpollWait(epollFd, events, -1)
		if err != nil {
			log.Printf("EpollWait error: %+v\n", err)
			continue
		}

		for i := range newEvents {
			if int(events[i].Fd) == serverFd {
				acceptClient(serverFd, epollFd)
			} else {
				handleClientIO(core.FD(events[i].Fd))
			}
		}
	}
}

// setupTCPServer creates a non-blocking TCP socket bound to config.Host:config.Port and starts listening.
func setupTCPServer(maxClients int) (int, error) {
	// TODO: Try to implement the kqueue or IOCP ones as well.
	serverFd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		return 0, err
	}

	// TODO: This might not be needed since we already set it via the flag above
	// but let's be safe and set it again just in case, if it's already set it should be a no-op
	if err := syscall.SetNonblock(serverFd, true); err != nil {
		_ = syscall.Close(serverFd)
		return 0, err
	}

	ipv4 := net.ParseIP(config.Host).To4()
	if ipv4 == nil {
		_ = syscall.Close(serverFd)
		return 0, fmt.Errorf("invalid IPv4 address: %s", config.Host)
	}

	var addr [4]byte
	copy(addr[:], ipv4)

	if err := syscall.Bind(serverFd, &syscall.SockaddrInet4{Port: config.Port, Addr: addr}); err != nil {
		_ = syscall.Close(serverFd)
		return 0, err
	}

	if err := syscall.Listen(serverFd, maxClients); err != nil {
		_ = syscall.Close(serverFd)
		return 0, err
	}

	return serverFd, nil
}

// setupEpoll creates an epoll instance and registers serverFd for read events.
func setupEpoll(serverFd int) (int, error) {
	// syscall.EPOLL_CLOEXEC - automatically closes FD on execve()
	epollFd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return 0, err
	}

	// notify me when serverFd becomes readable (incoming connections)
	serverSocketEvent := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFd),
	}

	if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, serverFd, &serverSocketEvent); err != nil {
		_ = syscall.Close(epollFd)
		return 0, err
	}

	return epollFd, nil
}

// acceptClient accepts a new connection on serverFd and registers it with epoll.
func acceptClient(serverFd, epollFd int) {
	connFd, sockAddr, err := syscall.Accept(serverFd)
	if err != nil {
		log.Printf("Accept error: %+v\n", err)
		return
	}

	if addr, ok := sockAddr.(*syscall.SockaddrInet4); ok {
		log.Printf("Accepted connection from %s:%d", net.IP(addr.Addr[:]).String(), addr.Port)
	} else {
		log.Printf("Accepted connection from %+v", sockAddr)
	}

	if err := syscall.SetNonblock(connFd, true); err != nil {
		log.Printf("SetNonblock error: %+v\n", err)
		if closeErr := syscall.Close(connFd); closeErr != nil {
			log.Printf("failed to close conn FD %d: %v\n", connFd, closeErr)
		}
		return
	}

	connectedClients++
	log.Println("Current connected clients - ", connectedClients)

	clientSocketEvent := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(connFd),
	}

	if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, connFd, &clientSocketEvent); err != nil {
		log.Printf("syscall error - epollCtl, %+v", err)
	}
}

// handleClientIO reads commands from a connected client fd and writes the response.
func handleClientIO(fd core.FD) {
	cmds, err := readCommands(fd)
	if err != nil {
		if closeErr := syscall.Close(int(fd)); closeErr != nil {
			log.Printf("failed to close conn FD %d: %v\n", fd, closeErr)
		}

		connectedClients--
		log.Println("Current connected clients - ", connectedClients)

		return
	}

	respond(cmds, fd)
}
