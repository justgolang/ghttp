package gracego

import (
	"net"
)

type Connection struct {
	net.Conn
	listener *Listener

	closed bool
}

func (this *Connection) Close() error {

	if !this.closed {
		this.closed = true
		this.listener.wg.Done()
	}

	return this.Conn.Close()
}
