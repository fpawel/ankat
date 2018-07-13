package dataproducts

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type Series struct {
	PartyID   PartyID   `db:"party_id"`
	CreatedAt time.Time `db:"created_at"`
	SeriesID  int64     `db:"series_id"`
	Name      string    `db:"name"`
}

func CreateNewSeries(x *sqlx.DB) {
	x.MustExec(`
INSERT INTO  series (work_id) 
VALUES ( (SELECT work_id FROM work  ORDER BY created_at DESC LIMIT 1));`)
}

func GetLastSeries() (x *sqlx.DB, series Series) {
	dbMustGet(x, series, `SELECT * FROM last_series`)
	return
}

func AddChartValue(x *sqlx.DB, serial, keyVar int, value float64) {
	x.MustExec(`
INSERT INTO chart_value (series_id, product_serial, read_var_id, x, y)
VALUES
  ( (SELECT series_id FROM last_series), ?, ?,
    (SELECT (julianday(STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))  -
             julianday( (SELECT created_at FROM last_series) ))),
    ? );;
`, serial, keyVar, value)
}

func DeleteLastEmptySeries(x *sqlx.DB) {
	x.MustExec(`
WITH a AS (SELECT series_id FROM last_series)
DELETE FROM series WHERE
  series_id IN (SELECT * FROM a) AND
  NOT exists(
    SELECT * FROM chart_value WHERE chart_value.series_id IN (SELECT * FROM a))
`)
}

func DeleteAllEmptySeries(x *sqlx.DB) {
	x.MustExec(`
DELETE FROM series WHERE NOT exists(
    SELECT * FROM chart_value WHERE chart_value.series_id = series.series_id)
`)
}
