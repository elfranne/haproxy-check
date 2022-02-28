package main

import "text/template"

// In order to make the metrics queryable, they are imported into an in-memory
// sqlite3 database. Since we don't know the schema of the metrics ahead of time,
// and since the schema varies for different HAProxy versions, some dynamic
// code is needed to produce the sqlite3 schema. These templates are used with
// the scraped CSV to create the schema and insert the metrics as SQL rows.
//
// Some SQL injection protection has been implemented. For instance the columns
// are sanitized so that they cannot contain anything other than alphanumeric
// characters. For the inserts, placeholders are used so that there is no
// opportunity for SQLi.

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
