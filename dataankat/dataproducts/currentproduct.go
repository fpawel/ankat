package dataproducts

import (
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
	"github.com/jmoiron/sqlx"
)

type CurrentProduct struct {
	Checked       bool                `db:"checked"`
	Comport       string              `db:"comport"`
	ProductSerial ankat.ProductSerial `db:"product_serial"`
	Ordinal       int                 `db:"ordinal"`

	db *sqlx.DB
}

func (x CurrentProduct) SetCoefficientValue(coefficient ankat.Coefficient, value float64) {
	x.db.MustExec(`
INSERT OR REPLACE INTO product_coefficient_value (party_id, product_serial, coefficient_id, value)
VALUES ((SELECT party_id FROM current_party),
        $1, $2, $3); `, x.ProductSerial, coefficient, value)
}

func (x CurrentProduct) CoefficientValue(coefficient ankat.Coefficient) (float64, bool) {
	var xs []float64
	dbutils.MustSelect(x.db, &xs, `
SELECT value FROM current_party_coefficient_value 
WHERE product_serial=$1 AND coefficient_id = $2;`, x.ProductSerial, coefficient)
	if len(xs) > 0 {
		return xs[0], true
	}
	return 0, false
}

func (x CurrentProduct) SetSectCoefficients(sect ankat.Sect, values []float64) {
	for i := range values {
		x.SetCoefficientValue(sect.Coefficient0()+ankat.Coefficient(i), values[i])
	}
}

func (x CurrentProduct) Value(p ankat.ProductVar) (float64, bool) {
	return DBProducts{x.db}.CurrentParty().ProductValue(x.ProductSerial, p)
}

func (x CurrentProduct) SetValue(p ankat.ProductVar, value float64) {
	x.db.MustExec(`
INSERT OR REPLACE INTO product_value (party_id, product_serial, section, point, var, value)
VALUES ((SELECT party_id FROM current_party), ?, ?, ?, ?, ?); `, x.ProductSerial, p.Sect, p.Point, p.Var, value)
}
