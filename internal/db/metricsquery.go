package db

const (
	checkMetricDB = `
select * from metrics;`

	createTable = `
CREATE TABLE metrics (
	         id varchar primary key,
	         m_type varchar NOT NULL,
	         delta bigint,
	         value double precision);`

	queryGetCounterMetricValue = `
SELECT delta FROM metrics WHERE id = $1 
`

	queryGetMetricNames = `
SELECT id FROM metrics;
`

	queryGetMetric = `
SELECT id, m_type, delta, value FROM metrics WHERE $1 = id
`
	queryGetGaugeMetricValue = `
SELECT value FROM metrics WHERE id = $1
`

	queryUpdateMetrics = `
INSERT INTO metrics(id,
	m_type,
	delta,
	value)
values ($1, $2, $3, $4)
on conflict(id) do 
update set 
	m_type=excluded.m_type,
	delta=metrics.delta+excluded.delta,
	value=excluded.value
`
)
