package main

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {
	serverReady := make(chan struct{})
	done := make(chan struct{})

	go func() {
		server, err := net.Listen("tcp", ":8080")
		if err != nil {
			t.Errorf("Failed to listen: %s", err)
			return
		}

		serverReady <- struct{}{}
		conn, err := server.Accept()
		if err != nil {
			return
		}

		t.Logf("conn = %#v", conn)

		// this is a HACK.
		buf := make([]byte, 100)
		sofar := 0
		expectedLen := -1
		colonPos := 0
		for {
			n, readErr := conn.Read(buf[sofar:])
			if n > 0 {
				t.Logf("read %d bytes, buf = %s", n, buf)
				sofar += n
			}

			if expectedLen == -1 {
				if colonPos = strings.IndexByte(string(buf), ':'); colonPos > -1 {
					x, err := strconv.ParseInt(string(buf[:colonPos]), 10, 64)
					if err != nil {
						t.Errorf("Failed to parse length in netstring: %s", err)
						break
					}
					expectedLen = int(x)
					t.Logf("Expecting %d bytes", expectedLen)
				}
			}

			if expectedLen > -1 {
				if sofar-(colonPos+1) >= expectedLen {
					payload := buf[colonPos+1 : colonPos+1+expectedLen]
					t.Logf("Buf = %s", payload)
					conn.Write(payload)
					break
				}
			}

			if readErr != nil {
				t.Logf("error while reading: %s", err)
				break
			}
		}
		done <- struct{}{}
	}()

	<-serverReady

	appAddr := "127.0.0.1:8080"
	proxyAddr := "127.0.0.1:8888"

	quit := make(chan struct{})
	p := NewProxy(proxyAddr, appAddr)
	go p.Run(quit)

	p.WaitReady()

	//	<-proxyReady
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		t.Errorf("Failed to connect to server: %s", err)
		return
	}

	t.Logf("Sending string")
	conn.Write([]byte("4:ab"))
	time.Sleep(time.Second)
	conn.Write([]byte("cd"))

	buf := make([]byte, 4)
	conn.Read(buf)
	conn.Close()

	<-done

	t.Logf("resopnse = %s", buf)

}