package main

import "fmt"

var metrics = []string{
	"active_servers",
	"backup_servers",
	"bin",
	"bout",
	"chkfail",
	"ctime",
	"dreq",
	"dresp",
	"econ",
	"ereq",
	"eresp",
	"http_response_1xx",
	"http_response_2xx",
	"http_response_3xx",
	"http_response_4xx",
	"http_response_5xx",
	"http_response_other",
	"qcur",
	"qmax",
	"qtime",
	"rate",
	"rtime",
	"scur",
	"slim",
	"smax",
	"ttime",
	"weight",
	"wredis",
	"wretr",
}

var nameLookup = map[string]string{
	"proxy":               "pxname",
	"sv":                  "svname",
	"active_servers":      "act",
	"backup_servers":      "bck",
	"cli_abort":           "cli_abrt",
	"srv_abort":           "srv_abrt",
	"http_response.1xx":   "hrsp_1xx",
	"http_response.2xx":   "hrsp_2xx",
	"http_response.3xx":   "hrsp_3xx",
	"http_response.4xx":   "hrsp_4xx",
	"http_response.5xx":   "hrsp_5xx",
	"http_response.other": "hrsp_other",
}

var instanceTypes = []string{
	"frontend",
	"backend",
	"server",
	"listener",
}

func lookupName(metric string) string {
	val, ok := nameLookup[metric]
	if ok {
		return val
	}
	return metric
}

var tags = []interface{}{
	"proxy_cluster",
	"host",
	"type",
	"proxy_system",
	"component",
	"proxy",
	"sv",
}

type Query struct {
	Metric       string
	ProxyCluster string
	Type         string
	Host         string
	ProxySystem  string
	Component    string
	Proxy        string
	Service      string
}

func (q Query) SQL() string {
	return fmt.Sprintf(`SELECT %s, %s, %s, %s, %s, %s, %s, %s
	FROM metrics;
	`, append(tags, q.Metric)...)
}

func outputMetrics(data *statsData) error {
	db, err := createDB(data)
	if err != nil {
		return err
	}

}
