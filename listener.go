package gracego

import (
	"net"
	"sync"
	"time"
)

type Listener struct {
	*net.TCPListener
	wg *sync.WaitGroup
}

func newListener(ln *net.TCPListener) *Listener {
	return &Listener{
		TCPListener: ln,
		wg:          &sync.WaitGroup{},
	}
}

func (ln *Listener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()

	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)

	ln.wg.Add(1)
	conn := &Connection{
		Conn:     tc,
		listener: ln,
	}

	return conn, nil
}

func (ln *Listener) Wait() {
	ln.wg.Wait()
}
