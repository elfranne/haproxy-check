package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/docker/go-units"
)

func readHTTP(url *url.URL, config Config) (*statsData, error) {
	urlString := url.String()
	if !strings.HasSuffix(urlString, ";csv") && strings.HasSuffix(urlString, "/stats") {
		// where is the content-type support, haproxy??
		urlString = path.Join(urlString, ";csv")
	}
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return nil, err
	}
	if config.AdminUser != "" || config.AdminPass != "" {
		req.SetBasicAuth(config.AdminUser, config.AdminPass)
	}

	var client http.Client

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	var tlsCfg *tls.Config

	if config.TLSCA != "" {
		tlsCfg = new(tls.Config)
		caCertPool, err := loadCACerts(config.TLSCA)
		if err != nil {
			return nil, err
		}
		// client trust store should ONLY consist of specified CAs
		tlsCfg.ClientCAs = caCertPool
	}

	if config.TLSCert != "" {
		if tlsCfg == nil {
			tlsCfg = new(tls.Config)
		}
		cert, err := tls.LoadX509KeyPair(config.TLSCert, config.TLSKey)
		if err != nil {
			return nil, fmt.Errorf("error loading x509 keypair: %s", err)
		}

		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	if config.InsecureSkipVerify {
		if tlsCfg == nil {
			tlsCfg = new(tls.Config)
		}
		tlsCfg.InsecureSkipVerify = config.InsecureSkipVerify
	}

	transport.TLSClientConfig = tlsCfg

	client.Transport = transport

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server responded with status %d", resp.StatusCode)
	}

	reader := io.LimitReader(resp.Body, units.MB)

	var buf bytes.Buffer

	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, err
	}

	return &statsData{data: buf.Bytes()}, nil
}

func readUnix(url *url.URL) (*statsData, error) {
	conn, err := net.Dial("unix", url.Path)
	if err != nil {
		return nil, fmt.Errorf("error dialing %s: %s", url.String(), err)
	}
	defer conn.Close()
	if _, err := conn.Write([]byte("show stat\n")); err != nil {
		return nil, fmt.Errorf("error querying %s: %s", url.String(), err)
	}
	reader := io.LimitReader(conn, units.MB)
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, fmt.Errorf("error reading %s: %s", url.String(), err)
	}
	return &statsData{data: buf.Bytes()}, nil
}
