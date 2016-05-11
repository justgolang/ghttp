package ghttp

import (
	"net"
	"fmt"
)

type Connection struct {
	net.Conn
	listener *Listener
	closed bool
}

func (cnt *Connection) Close() error {
	fmt.Println("close......")
	if !cnt.closed {
		cnt.closed = true
		cnt.listener.wg.Done()
	}

	return cnt.Conn.Close()
}
