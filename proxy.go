package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type CloseReader interface {
	io.Reader
	CloseRead() error
}

type CloseWriter interface {
	io.Writer
	CloseWrite() error
}

func proxyConn(dst CloseWriter, src CloseReader) {
	defer func() {
		defer recover()
		dst.CloseWrite()
	}()
	defer func() {
		defer recover()
		src.CloseRead()
	}()
	io.Copy(dst, src)
}

type proxy struct {
	bindAddr string
	fwdAddr  string
	ready    chan struct{}
	mutex    *sync.RWMutex
	alive    bool
}

func NewProxy(bindAddr, fwdAddr string) *proxy {
	return &proxy{bindAddr, fwdAddr, make(chan struct{}), &sync.RWMutex{}, true}
}

func (p *proxy) WaitReady() {
	<-p.ready
}

func (p *proxy) Run(quit <-chan struct{}) error {
	server, err := net.Listen("tcp", p.bindAddr)
	if err != nil {
		p.ready <- struct{}{}
		return fmt.Errorf("Failed to listen at %s: %s", p.bindAddr, err)
	}

	p.ready <- struct{}{}

	defer server.Close()

	// We listen in a separate goroutine to allow termination
	acceptCh := make(chan net.Conn)
	go func() {
		for p.running() {
			conn, err := server.Accept()
			if err != nil {
				continue
			}
			acceptCh<-conn
		}
	}()

	for p.running() {
		select {
		case conn := <-acceptCh:
			if conn == nil {
				continue
			}
			p.processConn(conn)
		case <-quit:
			p.stop()
		}
	}

	return nil
}

func (p *proxy) running() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.alive
}

func (p *proxy) stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.alive = false
}

func (p *proxy) processConn(conn net.Conn) {
	endpoint, err := net.Dial("tcp", p.fwdAddr)
	if err != nil {
		return
	}
	go proxyConn(endpoint.(CloseWriter), conn.(CloseReader))
	go proxyConn(conn.(CloseWriter), endpoint.(CloseReader))
}
