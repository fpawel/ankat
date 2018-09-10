package dataproducts

import (
	"fmt"
	"github.com/fpawel/ankat"
	"time"
)

type Series struct {
	createdAt time.Time
	records []seriesValue
}

type seriesValue = struct {
	SecondsOffset float64
	Value float64
	ProductSerial ankat.ProductSerial
	Var ankat.Var
}

func NewSeries() (x *Series)  {
	return &Series{
		createdAt:time.Now(),
	}
}

func (x *Series) AddRecord(p ankat.ProductSerial, v ankat.Var, value float64)  {
	x.records = append(x.records, seriesValue{
		ProductSerial:p,
		Var:v,
		SecondsOffset:time.Since(x.createdAt).Seconds(),
		Value:value,
	} )
}

func (x *Series) Save(db DBProducts, name string)  {
	partyID := db.CurrentParty().PartyID
	r := db.DB.MustExec(`
INSERT INTO  series ( created_at, name, party_id) 
VALUES (?, ?, ?);`, x.createdAt, name, partyID)
	seriesID,err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	var args []interface{}
	queryStr := `
INSERT INTO chart_value(series_id, party_id, product_serial, var, seconds_offset, value) VALUES `
	for i,a := range x.records {
		b := []interface{}{
			seriesID, partyID, a.ProductSerial, a.Var,
			a.SecondsOffset, a.Value,
		}
		s := fmt.Sprintf("(%d, %d, %d, %d, %v, %v)", b...)
		args = append(args, b...)
		if i == len(x.records)-1 {
			s += ";"
		} else {
			s += ", "
		}
		queryStr += s
	}
	db.DB.MustExec(queryStr, args...)

}




