package ghttp

import (
	"crypto/tls"
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
	}
)

const (
	ghttpKey = "ghttp"
)

var (
	originalWD, _ = os.Getwd()
)

func ListenAndServe(addr string, handler http.Handler) error {
	return newServer(addr, handler).ListenAndServe()
}

func ListenAndServeTLS(addr string, certFile string, keyFile string, handler http.Handler) error {
	return newServer(addr, handler).ListenAndServeTLS(certFile, keyFile)
}

func newServer(addr string, handler http.Handler) *Server {
	isGracefulRestart := false
	if os.Getenv(ghttpKey) != "" {
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
	addr := srv.httpServer.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := srv.getListener(addr)
	if err != nil {
		return err
	}

	go handleSignals(ln)
	err = srv.httpServer.Serve(ln)

	log.Printf("waiting for connection close...")
	ln.Wait()
	log.Printf("all connection closed, process with pid %d is shutting down...", os.Getpid())

	return err
}

func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	addr := srv.httpServer.Addr
	if addr == "" {
		addr = ":https"
	}

	config := &tls.Config{}
	if srv.httpServer.TLSConfig != nil {
		*config = *srv.httpServer.TLSConfig
	}

	if !strSliceContains(config.NextProtos, "http/1.1") {
		config.NextProtos = append(config.NextProtos, "http/1.1")
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	ln, err := srv.getListener(addr)
	if err != nil {
		return err
	}
	ln.config = config

	go handleSignals(ln)
	err = srv.httpServer.Serve(ln)

	log.Printf("waiting for connection close...")
	ln.Wait()
	log.Printf("all connection closed, process with pid %d is shutting down...", os.Getpid())

	return err
}

func strSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func handleSignals(ln *Listener) {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGUSR2)
	for sig := range signals {
		if sig == syscall.SIGTERM {
			log.Printf("got signal TERM, shutting down...")
			ln.Close()
		}

		if sig == syscall.SIGUSR2 {
			log.Printf("got signal USR2, restarting...")
			if pid, err := startNewProcess(ln); err != nil {
				log.Printf("fail to start new process: %v", err)
			} else {
				log.Printf("start new process with pid: %d", pid)
				ln.Close()
			}
		}
	}
}

func startNewProcess(ln *Listener) (int, error) {
	argv0, err := exec.LookPath(os.Args[0])
	if err != nil {
		log.Println("fail")
		return 0, err
	}

	// Set a flag for the new process start process
	os.Setenv(ghttpKey, "1")

	var env []string
	for _, v := range os.Environ() {
		env = append(env, v)
	}

	file, err := ln.File()
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

	return newListener(ln.(*net.TCPListener)), nil
}
