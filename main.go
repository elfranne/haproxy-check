package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	URL                string
	Backends           []string
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
			Path:      "url",
			Env:       "HAPROXY_URL",
			Argument:  "url",
			Shorthand: "u",
			Default:   "unix:///run/haproxy/admin.sock",
			Usage:     "URL to the HAProxy administration service",
			Value:     &config.URL,
		},
		&sensu.PluginConfigOption{
			Path:      "backends",
			Env:       "HAPROXY_BACKENDS",
			Argument:  "backends",
			Shorthand: "b",
			Default:   []string{},
			Usage:     "list of backends to fetch stats from (fetch all by default)",
			Value:     &config.Backends,
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
		fmt.Printf("Error check stdin: %v\n", err)
		panic(err)
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
	if len(config.URL) == 0 {
		return sensu.CheckStateWarning, fmt.Errorf("--url or HAPROXY_URL environment variable is required")
	}
	if _, err := url.Parse(config.URL); err != nil {
		return sensu.CheckStateWarning, fmt.Errorf("invalid URL: %s", err)
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
	return sensu.CheckStateWarning, errors.New("FAIL")
}
