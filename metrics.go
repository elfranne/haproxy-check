package main

import "errors"

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

var tags = []string{
	"proxy_cluster",
	"host",
	"type",
	"proxy_system",
	"component",
	"proxy",
	"sv",
}

func outputMetrics(data *statsData) error {
	return errors.New("not implemented")
}
