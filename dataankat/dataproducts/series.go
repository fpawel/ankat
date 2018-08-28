package dataproducts

import (
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
	"time"
)

type Series struct {
	PartyID   ankat.PartyID `db:"party_id"`
	CreatedAt time.Time     `db:"created_at"`
	SeriesID  int64         `db:"series_id"`
	Name      string        `db:"name"`
}

func (x DBProducts) CreateNewSeries() {
	x.DB.MustExec(`
INSERT INTO  series (work_id) 
VALUES ( (SELECT work_id FROM work  ORDER BY created_at DESC LIMIT 1));`)
}

func (x DBProducts) GetLastSeries() (series Series) {
	dbutils.MustGet(x.DB, series, `SELECT * FROM last_series`)
	return
}

func (x DBProducts) AddChartValue( serial ankat.ProductSerial, keyVar ankat.Var, value float64) {
	x.DB.MustExec(`
INSERT INTO chart_value (series_id, product_serial, var, x, y)
VALUES
  ( (SELECT series_id FROM last_series), ?, ?,
    (SELECT (julianday(STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))  -
             julianday( (SELECT created_at FROM last_series) ))),
    ? );;
`, serial, keyVar, value)
}

func (x DBProducts) DeleteLastEmptySeries() {
	x.DB.MustExec(`
WITH a AS (SELECT series_id FROM last_series)
DELETE FROM series WHERE
  series_id IN (SELECT * FROM a) AND
  NOT exists(
    SELECT * FROM chart_value WHERE chart_value.series_id IN (SELECT * FROM a))
`)
}

func (x DBProducts) DeleteAllEmptySeries() {
	x.DB.MustExec(`
DELETE FROM series WHERE NOT exists(
    SELECT * FROM chart_value WHERE chart_value.series_id = series.series_id)
`)
}
