package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"testing"
)

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
