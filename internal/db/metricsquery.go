package db

const (
	checkMetricDB = `
select * from metrics;`

	createTable = `
CREATE TABLE metrics (
	         id varchar(20) primary key,
	         m_type varchar NOT NULL,
	         delta INT,
	         value FLOAT);`

	queryGetCounterMetricValue = `
SELECT delta FROM metrics WHERE $1 = id
`
	queryGetMetric = `
SELECT id, m_type, delta, value FROM metrics WHERE $1 = id
`
	queryGetGaugeMetricValue = `
SELECT value FROM metrics WHERE $1 = id
`

	queryUpdateMetrics = `
INSERT INTO metrics(id,
	m_type,
	delta,
	value)
values ($1, $2, $3, $4)
on conflict(id) do 
update set 
	id=excluded.id,
	m_type=excluded.m_type,
	delta=excluded.delta,
	value=excluded.value
`
)