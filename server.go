package gracego

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type (
	Server struct {
		httpServer        *http.Server
		isGracefulRestart bool
		listener          *Listener
	}
)

const (
	gracefulRestartKey = "gogracefully"
)

var (
	originalWD, _ = os.Getwd()
)

func ListenAndServe(addr string, handler http.Handler) error {
	return newServer(addr, handler).ListenAndServe()
}

func newServer(addr string, handler http.Handler) *Server {
	isGracefulRestart := false
	if os.Getenv(gracefulRestartKey) != "" {
		isGracefulRestart = true
	}
	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		isGracefulRestart: isGracefulRestart,
	}
}

func (srv *Server) ListenAndServe() error {
	ln, err := srv.getListener(srv.httpServer.Addr)
	if err != nil {
		return err
	}

	go handleSignals(ln)
	err = srv.httpServer.Serve(ln)

	//listener close后　wait老连接处理完毕
	log.Println("waiting for close")
	ln.Wait()
	log.Println("old connection has closed")
	return err
}

func handleSignals(ln *Listener) {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGUSR2)
	for sig := range signals {
		if sig == syscall.SIGTERM {

		}

		if sig == syscall.SIGUSR2 {
			if pid, err := startNewProcess(ln); err != nil {
				log.Printf("fail to start new process: %v", err)
			} else {
				//wait
				ln.Close()
				log.Printf("successfully restarted with new pid: %d", pid)
			}
		}
	}
}

func startNewProcess(listener *Listener) (int, error) {
	argv0, err := exec.LookPath(os.Args[0])
	if err != nil {
		log.Println("fail")
		return 0, err
	}

	// Set a flag for the new process start process
	os.Setenv(gracefulRestartKey, "1")

	var env []string
	for _, v := range os.Environ() {
		env = append(env, v)
	}

	file, err := listener.File()
	if err != nil {
		return 0, nil
	}
	pid, err := syscall.ForkExec(argv0, os.Args, &syscall.ProcAttr{
		Dir:   originalWD,
		Env:   env,
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), file.Fd()},
	})
	if err != nil {
		return 0, nil
	}

	return pid, nil
}

func (srv *Server) getListener(addr string) (*Listener, error) {
	if tcpListener, err := srv.getTCPListener(srv.httpServer.Addr); err != nil {
		return newListener(tcpListener), nil
	} else {
		return nil, err
	}
}

func (srv *Server) getTCPListener(addr string) (*net.TCPListener, error) {
	var err error
	var ln net.Listener
	if srv.isGracefulRestart {
		file := os.NewFile(3, "")
		if ln, err = net.FileListener(file); err != nil {
			return nil, fmt.Errorf("fail get Listener on net.FileListener:%v", err)
		}
	} else {
		if ln, err = net.Listen("tcp", addr); err != nil {
			return nil, fmt.Errorf("fail get Listener on net.Listen:%v", err)
		}
	}
	return ln.(*net.TCPListener), nil
}
