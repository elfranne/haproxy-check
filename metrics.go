package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func init() {
	for _, v := range prometheusMetrics {
		prometheus.MustRegister(v)
	}
	// unregister the default collectors. We could also create a new registry,
	// but this was actually easier to figure out how to do.
	prometheus.Unregister(collectors.NewGoCollector())
	prometheus.Unregister(collectors.NewBuildInfoCollector())
	prometheus.Unregister(collectors.NewExpvarCollector(nil))
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
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

func lookupHelp(key string) string {
	if k, ok := nameLookup[key]; ok {
		key = k
	}
	return helpLookup[key]
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

var helpLookup = map[string]string{
	"pxname":         "proxy name",
	"svname":         "service name",
	"qcur":           "current queued requests",
	"qmax":           "max queued requests",
	"scur":           "session current",
	"smax":           "session max",
	"slim":           "session limit",
	"stot":           "session total",
	"bin":            "bytes in",
	"bout":           "bytes out",
	"dreq":           "request denied security",
	"dresp":          "response denied security",
	"ereq":           "request errors",
	"econ":           "connection errors",
	"eresp":          "response errors",
	"wretr":          "warning retries",
	"wredis":         "warning redispatched",
	"status":         "status",
	"weight":         "weight",
	"act":            "servers active",
	"bck":            "servers backup",
	"chkfail":        "healthcheck failed",
	"chkdown":        "healthcheck transitions",
	"lastchg":        "healthcheck seconds since change",
	"downtime":       "healthcheck downtime",
	"qlimit":         "server queue limit",
	"pid":            "process id",
	"iid":            "proxy id",
	"sid":            "server id",
	"throttle":       "server throttle percent",
	"lbtot":          "server selected",
	"tracked":        "tracked server id",
	"type":           "type",
	"rate":           "session rate",
	"rate_lim":       "session rate limit",
	"rate_max":       "session rate max",
	"check_status":   "check status",
	"check_code":     "check code",
	"check_duration": "healthcheck duration",
	"hrsp_1xx":       "response status 1xx",
	"hrsp_2xx":       "response status 2xx",
	"hrsp_3xx":       "response status 3xx",
	"hrsp_4xx":       "response status 4xx",
	"hrsp_5xx":       "response status 5xx",
	"hrsp_other":     "response status other",
	"hanafail":       "failed healthcheck details",
	"req_rate":       "requests per second",
	"req_rate_max":   "requests per second max",
	"req_tot":        "total requests",
	"cli_abrt":       "client transfer aborts",
	"srv_abrt":       "server transfer aborts",
	"comp_in":        "compressor in",
	"comp_out":       "compressor out",
	"comp_byp":       "compressor bytes",
	"comp_rsp":       "compressor responses",
	"lastsess":       "session last assigned seconds",
	"last_chk":       "healthcheck contents",
	"last_agt":       "agent check contents",
	"qtime":          "queue time",
	"ctime":          "connect time",
	"rtime":          "response time",
	"ttime":          "average time",
	"agent_status":   "agent status",
	"agent_code":     "agent code",
	"agent_duration": "agent duration",
	"check_desc":     "check description",
	"agent_desc":     "agent description",
	"check_rise":     "check rise",
	"check_fall":     "check fall",
	"check_health":   "check health",
	"agent_rise":     "agent rise",
	"agent_fall":     "agent fall",
	"agent_health":   "agent health",
	"addr":           "address",
	"cookie":         "cookie",
	"mode":           "mode",
	"algo":           "algorithm",
	"conn_rate":      "connection rate",
	"conn_rate_max":  "connection rate max",
	"conn_tot":       "connection tot",
	"intercepted":    "requests intercepted",
	"dcon":           "connection requests denied",
	"dses":           "session requests denied",
}

var prometheusMetrics = map[string]*prometheus.GaugeVec{
	"active_servers": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_active_servers",
		Help: lookupHelp("active_servers"),
	}, tags),
	"backup_servers": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_backup_servers",
		Help: lookupHelp("backup_servers"),
	}, tags),
	"bin": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_bin",
		Help: lookupHelp("bin"),
	}, tags),
	"bout": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_bout",
		Help: lookupHelp("bout"),
	}, tags),
	"chkfail": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_chkfail",
		Help: lookupHelp("chkfail"),
	}, tags),
	"ctime": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_ctime",
		Help: lookupHelp("ctime"),
	}, tags),
	"dreq": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_dreq",
		Help: lookupHelp("dreq"),
	}, tags),
	"dresp": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_dresp",
		Help: lookupHelp("dresp"),
	}, tags),
	"econ": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_econ",
		Help: lookupHelp("econ"),
	}, tags),
	"ereq": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_ereq",
		Help: lookupHelp("ereq"),
	}, tags),
	"eresp": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_eresp",
		Help: lookupHelp("eresp"),
	}, tags),
	"http_response_1xx": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_1xx",
		Help: lookupHelp("http_response_1xx"),
	}, tags),
	"http_response_2xx": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_2xx",
		Help: lookupHelp("http_response_2xx"),
	}, tags),
	"http_response_3xx": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_3xx",
		Help: lookupHelp("http_response_3xx"),
	}, tags),
	"http_response_4xx": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_4xx",
		Help: lookupHelp("http_response_4xx"),
	}, tags),
	"http_response_5xx": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_5xx",
		Help: lookupHelp("http_response_5xx"),
	}, tags),
	"http_response_other": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_http_response_other",
		Help: lookupHelp("http_response_other"),
	}, tags),
	"qcur": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_qcur",
		Help: lookupHelp("qcur"),
	}, tags),
	"qmax": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_qmax",
		Help: lookupHelp("qmax"),
	}, tags),
	"qtime": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_qtime",
		Help: lookupHelp("qtime"),
	}, tags),
	"rate": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_rate",
		Help: lookupHelp("rate"),
	}, tags),
	"rtime": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_rtime",
		Help: lookupHelp("rtime"),
	}, tags),
	"scur": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_scur",
		Help: lookupHelp("scur"),
	}, tags),
	"slim": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_slim",
		Help: lookupHelp("slim"),
	}, tags),
	"smax": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_smax",
		Help: lookupHelp("smax"),
	}, tags),
	"ttime": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_ttime",
		Help: lookupHelp("ttime"),
	}, tags),
	"weight": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_weight",
		Help: lookupHelp("weight"),
	}, tags),
	"wredis": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_wredis",
		Help: lookupHelp("wredis"),
	}, tags),
	"wretr": prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "haproxy_wretr",
		Help: lookupHelp("wretr"),
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

// do not modify this without also modifying tags
var itags = []interface{}{
	"pxname",
	"addr",
	"type",
	//"proxy_system",
	//"component",
	"svname",
}

// do not modify this without also modifying itags
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

// ScanArgs returns the arguments which are passed positionally to rows.Scan.
// If you want to add support for a new tag, this list must be added to.
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

// SetPrometheus writes the contents of the row to the prometheus gatherer.
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
	var hapType string
	if !r.Type.Valid {
		hapType = ""
	} else if r.Type.Int64 < int64(len(instanceTypes)) {
		hapType = instanceTypes[r.Type.Int64]
	}
	gauge.WithLabelValues(r.Proxy, r.Host.String, hapType, r.Service).Set(r.Metric.Float64)
}

// outputMetrics writes all the scraped CSV metrics to prometheus, and then
// scrapes prometheus to produce the final output.
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
