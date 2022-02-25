package main

import "text/template"

const metricsDDL = `
CREATE TABLE metrics (
	{{ range $i, $e := . }}
	{{ if $i }},{{ end }}
	{{ $e }}
	{{ end }}
);
`

var metricsDDLTmpl = template.Must(template.New("metricsddl").Parse(metricsDDL))

const insertMetric = `
INSERT INTO metrics VALUES (
	{{ range $i, $e := . }}
	{{ if $i }},{{ end }}
	?
	{{ end }}
);
`

var insertMetricTmpl = template.Must(template.New("metricsinsert").Parse(insertMetric))
