package main

import (
	"bufio"
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"

	_ "modernc.org/sqlite"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	URLs               []string
	AdminUser          string
	AdminPass          string
	TLSCA              string
	TLSCert            string
	TLSKey             string
	InsecureSkipVerify bool
}

var (
	config = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "haproxy-check",
			Short:    "Check health and status of an HAProxy instance",
			Keyspace: "sensu.io/plugins/haproxy-check/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Path:      "urls",
			Env:       "HAPROXY_URLS",
			Argument:  "urls",
			Shorthand: "u",
			Default:   []string{"unix:///run/haproxy/admin.sock"},
			Usage:     "URLs to query for HAProxy stats",
			Value:     &config.URLs,
		},
		&sensu.PluginConfigOption{
			Path:      "admin-user",
			Env:       "HAPROXY_ADMIN_USER",
			Argument:  "admin-user",
			Shorthand: "a",
			Default:   "",
			Usage:     "admin username to be supplied for basic auth, optional",
			Value:     &config.AdminUser,
		},
		&sensu.PluginConfigOption{
			Path:      "admin-pass",
			Env:       "HAPROXY_ADMIN_PASS",
			Argument:  "admin-pass",
			Shorthand: "p",
			Default:   "",
			Usage:     "admin password to be supplied for basic auth, optional",
			Value:     &config.AdminPass,
		},
		&sensu.PluginConfigOption{
			Path:     "tls-ca",
			Env:      "HAPROXY_TLS_CA",
			Argument: "tls-ca",
			Usage:    "TLS CA cert path, optional",
			Value:    &config.TLSCA,
		},
		&sensu.PluginConfigOption{
			Path:     "tls-cert",
			Env:      "HAPROXY_TLS_CERT",
			Argument: "tls-cert",
			Usage:    "TLS cert path, optional",
			Value:    &config.TLSCert,
		},
		&sensu.PluginConfigOption{
			Path:     "tls-key",
			Env:      "HAPROXY_TLS_KEY",
			Argument: "tls-key",
			Usage:    "TLS private key path, optional",
			Value:    &config.TLSKey,
		},
		&sensu.PluginConfigOption{
			Path:     "insecure-skip-verify",
			Env:      "HAPROXY_INSECURE_SKIP_VERIFY",
			Argument: "insecure-skip-verify",
			Usage:    "disable TLS hostname verification (DANGEROUS!)",
			Value:    &config.InsecureSkipVerify,
		},
	}
)

func main() {
	useStdin := false
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}
	//Check the Mode bitmask for Named Pipe to indicate stdin is connected
	if fi.Mode()&os.ModeNamedPipe != 0 {
		log.Println("using stdin")
		useStdin = true
	}

	check := sensu.NewGoCheck(&config.PluginConfig, options, checkArgs, executeCheck, useStdin)
	check.Execute()
}

func checkArgs(event *corev2.Event) (int, error) {
	if len(config.URLs) == 0 {
		return sensu.CheckStateWarning, fmt.Errorf("--url or HAPROXY_URL environment variable is required")
	}
	for _, cfgURL := range config.URLs {
		u, err := url.Parse(cfgURL)
		if err != nil {
			return sensu.CheckStateWarning, fmt.Errorf("invalid URL: %s", err)
		}
		switch u.Scheme {
		case "", "file", "unix", "http", "https":
		default:
			return sensu.CheckStateWarning, fmt.Errorf("unsupported protocol scheme: %s", u.Scheme)
		}
	}
	if err := checkTLS(config); err != nil {
		return sensu.CheckStateWarning, fmt.Errorf("invalid TLS configuration: %s", err)
	}
	return sensu.CheckStateOK, nil
}

func all(values ...string) bool {
	var all = true
	for _, value := range values {
		if value == "" {
			all = false
		}
	}
	return all
}

func any(values ...string) bool {
	var any = false
	for _, value := range values {
		if value != "" {
			any = true
		}
	}
	return any
}

func checkTLS(config Config) error {
	if !all(config.TLSKey, config.TLSCert) && any(config.TLSCA, config.TLSKey, config.TLSCert) {
		return errors.New("partial TLS configuration is not accepted")
	}
	return nil
}

func executeCheck(event *corev2.Event) (int, error) {
	for _, cfgURL := range config.URLs {
		url, err := url.Parse(cfgURL)
		if err != nil {
			// shouldn't happen as inputs are validated elsewhere
			return sensu.CheckStateWarning, err
		}
		if url.Scheme == "" || url.Scheme == "unix" || url.Scheme == "file" {
			data, err := readUnix(url)
			if err != nil {
				return sensu.CheckStateWarning, err
			}
			outputMetrics(data)
		} else if url.Scheme == "http" || url.Scheme == "https" {
			data, err := readHTTP(url, config)
			if err != nil {
				return sensu.CheckStateWarning, err
			}
			outputMetrics(data)
		} else {
			return sensu.CheckStateWarning, fmt.Errorf("unsupported protocol scheme: %s", err)
		}
	}
	return sensu.CheckStateWarning, errors.New("FAIL")
}

type statsData struct {
	data []byte
}

var columnNameRE = regexp.MustCompile(`^[A-Za-z0-9_\-].*$`)

func (s *statsData) ColumnNames() ([]string, error) {
	index := bytes.IndexByte(s.data, '\n')
	if index < 0 {
		return nil, errors.New("invalid stats data")
	}
	header := s.data[:index]
	header = bytes.TrimPrefix(header, []byte("# "))
	sep := []byte{','}
	columns := bytes.Split(bytes.TrimSuffix(header, sep), sep)
	result := make([]string, 0, len(columns))
	for _, column := range columns {
		if !columnNameRE.Match(column) {
			return nil, fmt.Errorf("illegal column name: %q", string(column))
		}
		if bytes.Equal(column, []byte("-")) {
			// who names a column this!?
			column = []byte("dash")
		}
		result = append(result, string(column))
	}
	return result, nil
}

func cast(data string) interface{} {
	if len(data) == 0 {
		return nil
	}
	ival, err := strconv.Atoi(data)
	if err == nil {
		return ival
	}
	fval, err := strconv.ParseFloat(data, 64)
	if err == nil {
		return fval
	}
	return data
}

func (s *statsData) Rows() ([][]interface{}, error) {
	index := bytes.IndexByte(s.data, '\n')
	if index < 0 {
		return nil, errors.New("invalid stats data")
	}
	if index == len(s.data)-1 {
		return nil, errors.New("no stats data found")
	}
	data := s.data[index+1:]
	result := [][]interface{}{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Bytes()
		sep := []byte{','}
		split := bytes.Split(line, sep)
		row := make([]interface{}, 0, len(split))
		for _, elem := range split {
			row = append(row, cast(string(elem)))
		}
		result = append(result, row)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func outputMetrics(data *statsData) error {
	return errors.New("not implemented")
}

func loadCACerts(path string) (*x509.CertPool, error) {
	caCerts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading CA file: %s", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCerts) {
		return nil, fmt.Errorf("No certificates could be parsed out of %s", path)
	}

	return caCertPool, nil
}
