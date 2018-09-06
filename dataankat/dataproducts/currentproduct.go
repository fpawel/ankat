package dataproducts

import (
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
	"github.com/jmoiron/sqlx"
)

type DBCurrentProduct struct {
	DB *sqlx.DB
	ProductSerial ankat.ProductSerial
}

func (x DBCurrentProduct) SetCoefficientValue(coefficient ankat.Coefficient, value float64) {
	x.DB.MustExec(`
INSERT OR REPLACE INTO product_coefficient_value (party_id, product_serial, coefficient_id, value)
VALUES ((SELECT party_id FROM current_party),
        $1, $2, $3); `, x.ProductSerial, coefficient, value)
}

func (x DBCurrentProduct) CoefficientValue( coefficient ankat.Coefficient) (float64, bool) {
	var xs []float64
	dbutils.MustSelect(x.DB, &xs, `
SELECT value FROM current_party_coefficient_value 
WHERE product_serial=$1 AND coefficient_id = $2;`, x.ProductSerial, coefficient)
	if len(xs) > 0 {
		return xs[0], true
	}
	return 0, false
}

func (x DBCurrentProduct) SetSectCoefficients(sect ankat.Sect, values []float64) {
	for i := range values{
		x.SetCoefficientValue(sect.Coefficient0() + ankat.Coefficient(i), values[i])
	}
}

func (x DBCurrentProduct) Value(p ankat.ProductVar) (float64, bool) {
	return DBCurrentParty{DBProducts{x.DB}}.ProductValue(x.ProductSerial, p)
}

func (x DBCurrentProduct) SetValue( p ankat.ProductVar, value float64) {
	x.DB.MustExec(`
INSERT OR REPLACE INTO product_value (party_id, product_serial, section, point, var, value)
VALUES ((SELECT party_id FROM current_party), ?, ?, ?, ?, ?); `, x.ProductSerial, p.Sect, p.Point, p.Var, value)
}