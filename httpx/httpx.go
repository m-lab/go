// Package httpx provides generic functions which extend the capabilities of
// the http package.
//
// The code here eliminates an annoying race condition in net/http that prevents
// you from knowing when it is safe to connect to the server socket. For the
// functions in this package, the listening socket is fully estabished when the
// function returns, and it is safe to run an HTTP GET immediately.
package httpx

import (
	"log"
	"net"
	"net/http"
	"time"
)

var logFatalf = log.Fatalf

// The code here is adapted from https://golang.org/src/net/http/server.go?s=85391:85432#L2742

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func serve(server *http.Server, listener net.Listener) {
	err := server.Serve(listener)
	if err != http.ErrServerClosed {
		logFatalf("Error, server %v closed with unexpected error %v", server, err)
	}
}

// ListenAndServeAsync starts an http server. The server will run until
// Shutdown() or Close() is called, but this function will return once the
// listening socket is established.  This means that when this function
// returns, the server is immediately available for an http GET to be run
// against it.
//
// Returns a non-nil error if the listening socket can't be established. Logs a
// fatal error if the server dies for a reason besides ErrServerClosed.
func ListenAndServeAsync(server *http.Server) error {
	// Start listening synchronously.
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	// Serve asynchronously.
	go serve(server, tcpKeepAliveListener{listener.(*net.TCPListener)})
	return nil
}

func serveTLS(server *http.Server, listener net.Listener, certFile, keyFile string) {
	err := server.ServeTLS(listener, certFile, keyFile)
	if err != http.ErrServerClosed {
		logFatalf("Error, server %v closed with unexpected error %v", server, err)
	}
}

// ListenAndServeTLSAsync starts an https server. The server will run until
// Shutdown() or Close() is called, but this function will return once the
// listening socket is established.  This means that when this function
// returns, the server is immediately available for an https GET to be run
// against it.
//
// Returns a non-nil error if the listening socket can't be established. Logs a
// fatal error if the server dies for a reason besides ErrServerClosed.
func ListenAndServeTLSAsync(server *http.Server, certFile, keyFile string) error {
	// Start listening synchronously.
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	// Serve asynchronously.
	go serveTLS(server, tcpKeepAliveListener{listener.(*net.TCPListener)}, certFile, keyFile)
	return nil
}
