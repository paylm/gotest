package main

import (
	"net"
	"time"
)

type timeoutConn struct {
	net.Conn
	timeout time.Duration
}

func (t *timeoutConn) Write(b []byte) (n int, err error) {
	d := time.Now().Add(t.timeout)
	t.SetDeadline(d)
	n, err = t.Conn.Write(b)
	return
}

func (t *timeoutConn) Read(b []byte) (n int, err error) {
	d := time.Now().Add(t.timeout)
	t.SetDeadline(d)
	n, err = t.Conn.Read(b)
	return
}
