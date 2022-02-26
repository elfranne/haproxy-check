package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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
