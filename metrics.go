package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	for _, v := range prometheusMetrics {
		prometheus.MustRegister(v)
	}
	// unregister go collector
	prometheus.Unregister(prometheus.NewGoCollector())
	prometheus.Unregister(prometheus.NewBuildInfoCollector())
	prometheus.Unregister(prometheus.NewExpvarCollector(nil))
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}

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
	"http_response_1xx":   "hrsp_1xx",
	"http_response_2xx":   "hrsp_2xx",
	"http_response_3xx":   "hrsp_3xx",
	"http_response_4xx":   "hrsp_4xx",
	"http_response_5xx":   "hrsp_5xx",
	"http_response_other": "hrsp_other",
}

var reverseNameLookup = map[string]string{
	"pxname":     "proxy",
	"svname":     "sv",
	"act":        "active_servers",
	"bck":        "backup_servers",
	"cli_abrt":   "cli_abort",
	"srv_abrt":   "srv_abort",
	"hrsp_1xx":   "http_response_1xx",
	"hrsp_2xx":   "http_response_2xx",
	"hrsp_3xx":   "http_response_3xx",
	"hrsp_4xx":   "http_response_4xx",
	"hrsp_5xx":   "http_response_5xx",
	"hrsp_other": "http_response_other",
}

var prometheusMetrics = map[string]*prometheus.GaugeVec{
	"active_servers": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_active_servers",
		Help: "?",
	}, tags),
	"backup_servers": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_backup_servers",
		Help: "?",
	}, tags),
	"bin": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_bin",
		Help: "?",
	}, tags),
	"bout": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_bout",
		Help: "?",
	}, tags),
	"chkfail": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_chkfail",
		Help: "?",
	}, tags),
	"ctime": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_ctime",
		Help: "?",
	}, tags),
	"dreq": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_dreq",
		Help: "?",
	}, tags),
	"dresp": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_dresp",
		Help: "?",
	}, tags),
	"econ": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_econ",
		Help: "?",
	}, tags),
	"ereq": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_ereq",
		Help: "?",
	}, tags),
	"eresp": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_eresp",
		Help: "?",
	}, tags),
	"http_response_1xx": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_1xx",
		Help: "?",
	}, tags),
	"http_response_2xx": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_2xx",
		Help: "?",
	}, tags),
	"http_response_3xx": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_3xx",
		Help: "?",
	}, tags),
	"http_response_4xx": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_4xx",
		Help: "?",
	}, tags),
	"http_response_5xx": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_5xx",
		Help: "?",
	}, tags),
	"http_response_other": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_other",
		Help: "?",
	}, tags),
	"qcur": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_qcur",
		Help: "?",
	}, tags),
	"qmax": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_qmax",
		Help: "?",
	}, tags),
	"qtime": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_qtime",
		Help: "?",
	}, tags),
	"rate": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_rate",
		Help: "?",
	}, tags),
	"rtime": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_rtime",
		Help: "?",
	}, tags),
	"scur": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_scur",
		Help: "?",
	}, tags),
	"slim": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_slim",
		Help: "?",
	}, tags),
	"smax": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_smax",
		Help: "?",
	}, tags),
	"ttime": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_ttime",
		Help: "?",
	}, tags),
	"weight": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_weight",
		Help: "?",
	}, tags),
	"wredis": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_wredis",
		Help: "?",
	}, tags),
	"wretr": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_wretr",
		Help: "?",
	}, tags),
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

var itags = []interface{}{
	"pxname",
	"addr",
	"type",
	//"proxy_system",
	//"component",
	"svname",
}

var tags = []string{
	"proxy",
	"host",
	"type",
	//"proxy_system",
	//"component",
	"sv",
}

type Row struct {
	ProxyCluster sql.NullString
	Host         sql.NullString
	Type         sql.NullInt64
	ProxySystem  sql.NullString
	Component    sql.NullString
	Proxy        string
	Service      string
	MetricName   string
	Metric       sql.NullFloat64
}

func (r *Row) ScanArgs() []interface{} {
	return []interface{}{
		&r.Proxy,
		&r.Host,
		&r.Type,
		//&r.ProxySystem,
		//&r.Component,
		&r.Service,
		&r.Metric,
		&r.MetricName,
	}
}

var types = map[int64]string{
	0: "frontend",
	1: "backend",
	2: "server",
}

func (r Row) SetPrometheus() {
	if !r.Metric.Valid {
		return
	}
	name, ok := reverseNameLookup[r.MetricName]
	if !ok {
		name = r.MetricName
	}
	gauge, ok := prometheusMetrics[name]
	if !ok {
		panic(fmt.Sprintf("can't find metric name: %s", r.MetricName))
	}
	hapType := types[r.Type.Int64]
	if !r.Type.Valid {
		hapType = ""
	}
	gauge.WithLabelValues(r.Proxy, r.Host.String, hapType, r.Service).Set(r.Metric.Float64)
}

func outputMetrics(data *statsData) error {
	db, err := createDB(data)
	if err != nil {
		return err
	}
	defer db.Close()
	for _, metric := range metrics {
		metric = lookupName(metric)
		fmtstr := "%s,'%s'"
		for range itags {
			fmtstr = "%s," + fmtstr
		}
		cols := fmt.Sprintf(fmtstr, append(itags, []interface{}{metric, metric}...)...)
		query := fmt.Sprintf("SELECT %s FROM metrics;", cols)
		err := doQuery(db, query)
		if err != nil {
			return err
		}
	}
	client := netListener.Client()
	resp, err := client.Get("http://fake/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func doQuery(db *sql.DB, query string) error {
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var row Row
		if err := rows.Scan(row.ScanArgs()...); err != nil {
			return err
		}
		row.SetPrometheus()
	}
	return nil
}
