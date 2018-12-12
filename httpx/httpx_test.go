package httpx

// Tests happen in the httpx package because we use whitebox testing to exercise error
// conditions.

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/kabukky/httpscerts"
	"github.com/m-lab/go/rtx"
)

func okay(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("\nOK!"))
}

func TestListenAndServeAsync(t *testing.T) {
	for i := 0; i < 1000; i++ {
		mux := http.NewServeMux()
		mux.HandleFunc("/", okay)
		server := &http.Server{
			Addr:    ":9090",
			Handler: mux,
		}
		rtx.Must(ListenAndServeAsync(server), "Could not start server")
		response, err := http.Get("http://localhost:9090/")
		if err != nil {
			t.Fatalf("HTTP server returned %v", err)
		}
		content := make([]byte, 20)
		n, err := response.Body.Read(content)
		if err != io.EOF {
			t.Errorf("Could not read response: %v", err)
		}
		if n != 4 {
			t.Errorf("Too many bytes: %d", n)
		}
		server.Shutdown(context.Background())
	}
}

func TestListenAndServeAsyncFailsWhenListenFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", okay)
	server := &http.Server{
		Addr:    ":9091",
		Handler: mux,
	}
	rtx.Must(ListenAndServeAsync(server), "Could not start server")
	defer server.Shutdown(context.Background())
	// One server works.  The next one should fail.
	server2 := &http.Server{
		Addr:    ":9091",
		Handler: mux,
	}
	err := ListenAndServeAsync(server2)
	if err == nil {
		t.Error("This should have failed")
	}
}

type listenerWithErrors struct {
	*net.TCPListener
}

func (l *listenerWithErrors) Accept() (net.Conn, error) {
	return nil, errors.New("This will not ever work")
}

var fakeFatalfCount = 0

func fakeFatalf(s string, args ...interface{}) {
	log.Printf("log.Fatal called in debug mode: "+s, args...)
	fakeFatalfCount++
}

// Whitebox test.
func TestListenAndServeAsyncWithPermanentNetworkFailure(t *testing.T) {
	// Make sure that we call log.Fatal when the server exits with anything other
	// than ErrServerClosed.
	fakeFatalfCount = 0
	logFatalf = fakeFatalf
	defer func() {
		logFatalf = log.Fatalf
		fakeFatalfCount = 0
	}()
	mux := http.NewServeMux()
	mux.HandleFunc("/", okay)
	server := &http.Server{
		Addr:    ":9092",
		Handler: mux,
	}
	if fakeFatalfCount != 0 {
		t.Errorf("fakeFatalCount should be 0 not %d", fakeFatalfCount)
	}
	serve(server, &listenerWithErrors{})
	if fakeFatalfCount != 1 {
		t.Errorf("fakeFatalCount should be 1 not %d", fakeFatalfCount)
	}
	server.Shutdown(context.Background())
}

func makeTestCertsWithCleanup(sname string) (certFile, keyfile string, roots *x509.CertPool, cleanup func()) {
	dir, err := ioutil.TempDir("", "tlsfilesfortesting")
	rtx.Must(err, "Could not create TLS file dir")
	key := dir + "/key.pem"
	cert := dir + "/cert.pem"
	rtx.Must(httpscerts.Generate(cert, key, sname), "Could not generate certs")
	certBytes, err := ioutil.ReadFile(cert)
	rtx.Must(err, "Could not read new cert")
	rootCAs, err := x509.SystemCertPool()
	rtx.Must(err, "Could not get existing CA pool")
	if !rootCAs.AppendCertsFromPEM(certBytes) {
		log.Fatal("Could not add new cert to root of trust")
	}
	return cert, key, rootCAs, func() {
		os.RemoveAll(dir)
	}
}

func TestListenAndServeTLSAsync(t *testing.T) {
	cert, key, roots, cleanup := makeTestCertsWithCleanup("localhost")
	defer cleanup()
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: roots},
	}}
	// 100 instead of 1000 because TLS server setup and teardown is 10x slower than non-TLS.
	for i := 0; i < 100; i++ {
		mux := http.NewServeMux()
		mux.HandleFunc("/", okay)
		server := &http.Server{
			Addr:    ":9093",
			Handler: mux,
		}
		rtx.Must(ListenAndServeTLSAsync(server, cert, key), "Could not start server")
		response, err := client.Get("https://localhost:9093/")
		if err != nil {
			t.Fatalf("HTTP server returned %v", err)
		}
		content := make([]byte, 20)
		n, err := response.Body.Read(content)
		if err != io.EOF {
			t.Errorf("Could not read response: %v", err)
		}
		if n != 4 {
			t.Errorf("Too many bytes: %d", n)
		}
		server.Shutdown(context.Background())
	}
}

func TestListenAndServeTLSAsyncFailsWhenListenFails(t *testing.T) {
	cert, key, roots, cleanup := makeTestCertsWithCleanup("localhost")
	defer cleanup()
	mux := http.NewServeMux()
	mux.HandleFunc("/", okay)
	server := &http.Server{
		Addr:    ":9094",
		Handler: mux,
	}
	rtx.Must(ListenAndServeTLSAsync(server, cert, key), "Could not start server")
	defer server.Shutdown(context.Background())
	// One race condition remains: if the server is not fully set up by the time
	// Shutdown is called, then the early shutdown might cause ServeTLS to return a
	// non-ErrServerClosed error. By running a GET here, we make sure that the
	// server is up and running, which ensures that its setup has completed.
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: roots},
	}}
	_, err := client.Get("https://localhost:9094/")
	rtx.Must(err, "Could not connect to local TLS server")
	// One server works.  The next one should fail.
	server2 := &http.Server{
		Addr:    ":9094",
		Handler: mux,
	}
	err = ListenAndServeTLSAsync(server2, cert, key)
	if err == nil {
		t.Error("This should have failed")
	}
}

// Whitebox test.
func TestListenAndServeTLSAsyncWithPermanentNetworkFailure(t *testing.T) {
	cert, key, _, cleanup := makeTestCertsWithCleanup("localhost")
	defer cleanup()
	// Make sure that we call log.Fatal when the server exits with anything other
	// than ErrServerClosed.
	fakeFatalfCount = 0
	logFatalf = fakeFatalf
	defer func() {
		logFatalf = log.Fatalf
		fakeFatalfCount = 0
	}()
	mux := http.NewServeMux()
	mux.HandleFunc("/", okay)
	server := &http.Server{
		Addr:    ":9095",
		Handler: mux,
	}
	if fakeFatalfCount != 0 {
		t.Errorf("fakeFatalCount should be 0 not %d", fakeFatalfCount)
	}
	serveTLS(server, &listenerWithErrors{}, cert, key)
	if fakeFatalfCount != 1 {
		t.Errorf("fakeFatalCount should be 1 not %d", fakeFatalfCount)
	}
}
