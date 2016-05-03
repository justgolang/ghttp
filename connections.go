package gracego

import (
	"net"
	"fmt"
)

type Connection struct {
	net.Conn
	listener *Listener
	closed bool
}

func (this *Connection) Close() error {
	fmt.Println("close......")
	if !this.closed {
		this.closed = true
		this.listener.wg.Done()
	}

	return this.Conn.Close()
}
