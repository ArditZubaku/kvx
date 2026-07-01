package core

import (
	"bytes"
	"fmt"
	"io"
	"syscall"
)

type Client struct {
	io.ReadWriter
	Fd       int
	cmdQueue []RedisCmd
	isTxn    bool
}

func NewClient(fd int) *Client {
	return &Client{
		Fd:       fd,
		cmdQueue: make([]RedisCmd, 0), // TODO: Prealloc
	}
}

func (c Client) Write(b []byte) (int, error) {
	return syscall.Write(c.Fd, b)
}

func (c Client) Read(b []byte) (int, error) {
	return syscall.Read(c.Fd, b)
}

func (c *Client) TxnBegin() {
	c.isTxn = true
}

func (c *Client) TxnExec() []byte {
	// TODO: Prealloc
	var out []byte
	buf := bytes.NewBuffer(out)

	fmt.Fprintf(buf, "*%d\r\n", len(c.cmdQueue))

	for _, cmd := range c.cmdQueue {
		buf.Write(executeCommand(cmd, c))
	}

	c.cmdQueue = make([]RedisCmd, 0)
	c.isTxn = false

	return buf.Bytes()
}

func (c *Client) TxnQueue(cmd RedisCmd) {
	c.cmdQueue = append(c.cmdQueue, cmd)
}

func (c *Client) TxnDiscard() {
	c.cmdQueue = make([]RedisCmd, 0)
	c.isTxn = false
}
