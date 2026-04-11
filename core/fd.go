package core

import "syscall"

// Wrapper type so I can implement the io.ReadWriter and wrap syscall.FD
type FD int

func (fd FD) Write(b []byte) (int, error) {
	return syscall.Write(int(fd), b)
}

func (fd FD) Read(b []byte) (int, error) {
	return syscall.Read(int(fd), b)
}
