package dataproducts

import (
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
	"github.com/jmoiron/sqlx"
)

type Product struct {
	PartyID       ankat.PartyID       `db:"party_id"`
	ProductSerial ankat.ProductSerial `db:"product_serial"`
	db            *sqlx.DB
}

func (x Product) Value(p ankat.ProductVar) (value float64, exits bool) {
	return productValue(x.db, x.PartyID, x.ProductSerial, p)
}

func (x Product) CoefficientValue(coefficient ankat.Coefficient) (float64, bool) {
	var xs []float64
	dbutils.MustSelect(x.db, &xs, `
SELECT value FROM current_party_coefficient_value 
WHERE product_serial=$1 AND coefficient_id = $2;`, x.ProductSerial, coefficient)
	if len(xs) > 0 {
		return xs[0], true
	}
	return 0, false
}