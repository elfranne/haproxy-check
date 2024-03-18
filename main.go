package main

import (
	"bytes"
	"crypto/x509"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"

	_ "modernc.org/sqlite"

	corev2 "github.com/sensu/core/v2"
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

	options = []sensu.ConfigOption{
		&sensu.SlicePluginConfigOption[string]{
			Path:      "urls",
			Env:       "HAPROXY_URLS",
			Argument:  "urls",
			Shorthand: "u",
			Default:   []string{"unix:///run/haproxy/admin.sock"},
			Usage:     "URLs to query for HAProxy stats",
			Value:     &config.URLs,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "admin-user",
			Env:       "HAPROXY_ADMIN_USER",
			Argument:  "admin-user",
			Shorthand: "a",
			Default:   "",
			Usage:     "admin username to be supplied for basic auth, optional",
			Value:     &config.AdminUser,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "admin-pass",
			Env:       "HAPROXY_ADMIN_PASS",
			Argument:  "admin-pass",
			Shorthand: "p",
			Default:   "",
			Usage:     "admin password to be supplied for basic auth, optional",
			Value:     &config.AdminPass,
		},
		&sensu.PluginConfigOption[string]{
			Path:     "tls-ca",
			Env:      "HAPROXY_TLS_CA",
			Argument: "tls-ca",
			Usage:    "TLS CA cert path, optional",
			Value:    &config.TLSCA,
		},
		&sensu.PluginConfigOption[string]{
			Path:     "tls-cert",
			Env:      "HAPROXY_TLS_CERT",
			Argument: "tls-cert",
			Usage:    "TLS cert path, optional",
			Value:    &config.TLSCert,
		},
		&sensu.PluginConfigOption[string]{
			Path:     "tls-key",
			Env:      "HAPROXY_TLS_KEY",
			Argument: "tls-key",
			Usage:    "TLS private key path, optional",
			Value:    &config.TLSKey,
		},
		&sensu.PluginConfigOption[bool]{
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
			if err := outputMetrics(data); err != nil {
				return sensu.CheckStateWarning, err
			}
		} else if url.Scheme == "http" || url.Scheme == "https" {
			data, err := readHTTP(url, config)
			if err != nil {
				return sensu.CheckStateWarning, err
			}
			if err := outputMetrics(data); err != nil {
				return sensu.CheckStateWarning, err
			}
		} else {
			return sensu.CheckStateWarning, fmt.Errorf("unsupported protocol scheme: %s", err)
		}
	}
	return sensu.CheckStateOK, nil
}

type statsData struct {
	data []byte
}

var columnNameRE = regexp.MustCompile(`^[A-Za-z0-9_\-].*$`)

func (s *statsData) ColumnNames() ([]string, error) {
	reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(s.data, []byte("#"))))
	reader.TrimLeadingSpace = true
	columns, err := reader.Read()
	if err != nil {
		return nil, err
	}
	for i := range columns {
		if columns[i] == "-" {
			// who names a column this!?
			columns[i] = "dash"
		}
		if columns[i] == "" && i == len(columns)-1 {
			columns = columns[:len(columns)-1]
			break
		}
		if !columnNameRE.MatchString(columns[i]) {
			return nil, fmt.Errorf("illegal column name: %q", columns[i])
		}
	}
	return columns, nil
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
	columns, err := s.ColumnNames()
	if err != nil {
		return nil, err
	}

	lenCols := len(columns)

	reader := csv.NewReader(bytes.NewReader(s.data))
	reader.Comment = '#'
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	result := make([][]interface{}, 0, len(rows))
	for _, row := range rows {
		irow := make([]interface{}, 0, len(row))
		for i, elem := range row {
			if i >= lenCols {
				break
			}
			irow = append(irow, cast(string(elem)))
		}
		result = append(result, irow)
	}

	return result, nil
}

func loadCACerts(path string) (*x509.CertPool, error) {
	caCerts, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading CA file: %s", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCerts) {
		return nil, fmt.Errorf("no certificates could be parsed out of %s", path)
	}

	return caCertPool, nil
}
