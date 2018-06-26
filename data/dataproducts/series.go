package dataproducts

import "time"

type Series struct{
	PartyID     PartyID `db:"party_id"`
	CreatedAt   time.Time   `db:"created_at"`
	SeriesID int64      `db:"series_id"`
	Name string `db:"name"`
}

func (x DB) GetLastSeries() (series Series){
	dbMustGet(x.DB, series, `SELECT * FROM last_series`)
	return
}
