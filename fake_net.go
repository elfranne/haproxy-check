package main

import (
	"context"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// The purpose of this code is to bind a prometheus HTTP handler to a net.Listener
// that doesn't actually open any sockets or files. That lets us re-use the prom
// HTTP logic to scrape the metrics without needing any system permissions to
// open ports or files.

var (
	netListener = NewFakeListener()
)

func init() {
	go http.Serve(netListener, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
}

type FakeListener struct {
	client net.Conn
	server net.Conn
}

func (f *FakeListener) Client() *http.Client {
	transport := http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return f.client, nil
		},
	}
	return &http.Client{
		Transport: &transport,
	}
}

func NewFakeListener() *FakeListener {
	client, server := net.Pipe()
	return &FakeListener{
		client: client,
		server: server,
	}
}

func (f *FakeListener) Accept() (net.Conn, error) {
	return f.server, nil
}

func (f *FakeListener) Close() error {
	_ = f.client.Close()
	_ = f.server.Close()
	return nil
}

func (f *FakeListener) Addr() net.Addr {
	return fakeAddr{}
}

type fakeAddr struct {
}

func (fakeAddr) Network() string {
	return "tcp"
}

func (fakeAddr) String() string {
	return "fake"
}
