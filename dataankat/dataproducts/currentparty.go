package dataproducts

import (
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
)

type CurrentParty struct {
	Party
}

func (x CurrentParty) CurrentProducts() (products []CurrentProduct) {
	dbutils.MustSelect(x.db, &products, `SELECT * FROM current_party_products_config;`)
	for i := range products{
		products[i].db = x.db
		products[i].PartyID = x.PartyID
	}
	return
}

func (x CurrentParty) CheckedProducts() (products []CurrentProduct) {
	dbutils.MustSelect(x.db, &products,
		`SELECT * FROM current_party_products_config WHERE checked=1;`)
	for i := range products{
		products[i].db = x.db
		products[i].PartyID = x.PartyID
	}
	return
}

func (x CurrentParty) CurrentProduct(n int) (p CurrentProduct) {
	dbutils.MustGet(x.db, &p, `SELECT * FROM current_party_products_config WHERE ordinal = $1;`, n)
	p.PartyID = x.PartyID
	p.db = x.db
	return
}



func (x CurrentParty) SetMainErrorConcentrationValue(ankatChan ankat.AnkatChan, scalePos ankat.ScalePosition,
	temperaturePos ankat.TemperaturePosition, serial ankat.ProductSerial, value float64) {
}
