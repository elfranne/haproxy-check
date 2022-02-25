package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestReadHTTPGoodServer(t *testing.T) {
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, err := w.Write(testDataCSV); err != nil {
			panic(err)
		}
	}))

	url, err := url.Parse(goodServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	data, err := readHTTP(url, Config{})
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data.data, testDataCSV) {
		fmt.Println(string(data.data))
		t.Error("got bad data from test server")
	}
}

func TestReadHTTPBadServer(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "oh noooooo", http.StatusInternalServerError)
	}))

	url, err := url.Parse(badServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := readHTTP(url, Config{}); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestReadHTTPAuthServer(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if username != "username" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if password != "password" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if _, err := w.Write(testDataCSV); err != nil {
			panic(err)
		}
	}))

	url, err := url.Parse(authServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := readHTTP(url, Config{}); err == nil {
		t.Fatal("expected non-nil error")
	}

	if _, err := readHTTP(url, Config{AdminUser: "username", AdminPass: "password"}); err != nil {
		t.Error(err)
	}

}

func TestReadSocket(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmpfile.Close()
	if err := os.RemoveAll(tmpfile.Name()); err != nil {
		t.Fatal(err)
	}
	sock, err := net.Listen("unix", tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer sock.Close()
	go func() {
		conn, err := sock.Accept()
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		buf := make([]byte, 10)
		if _, err := conn.Read(buf); err != nil {
			panic(err)
		}

		if string(buf) != "show stat\n" {
			panic(string(buf))
		}

		if _, err := conn.Write(testDataCSV); err != nil {
			panic(err)
		}
	}()
	url, err := url.Parse(sock.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	data, err := readUnix(url)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data.data, testDataCSV) {
		t.Error("unexpected data")
	}
}
