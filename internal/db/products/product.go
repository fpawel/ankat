package products

import (
	"github.com/fpawel/ankat/internal/ankat"
	"github.com/fpawel/goutils/dbutils"
	"github.com/jmoiron/sqlx"
)

type Product struct {
	PartyID       ankat.PartyID       `db:"party_id"`
	ProductSerial ankat.ProductSerial `db:"product_serial"`
	db            *sqlx.DB
}

func (x Product) Value(p ankat.ProductVar) (value float64, exits bool) {
	var xs []float64
	dbutils.MustSelect(x.db, &xs, `
SELECT value FROM product_value 
WHERE party_id = ? AND product_serial=? AND var = ? AND section = ? AND point = ?;`,
		x.PartyID, x.ProductSerial, p.Var, p.Sect, p.Point)
	if len(xs) == 0 {
		return
	}
	if len(xs) > 1 {
		panic("len must be 1 or 0")
	}
	value = xs[0]
	exits = true
	return
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

func (x Product) SetCoefficientValue(coefficient ankat.Coefficient, value float64) {
	x.db.MustExec(`
INSERT OR REPLACE INTO product_coefficient_value (party_id, product_serial, coefficient_id, value)
VALUES ((SELECT party_id FROM current_party),
        $1, $2, $3); `, x.ProductSerial, coefficient, value)
}



func (x Product) SetSectCoefficients(sect ankat.Sect, values []float64) {
	for i := range values {
		x.SetCoefficientValue(sect.Coefficient0()+ankat.Coefficient(i), values[i])
	}
}


func (x Product) SetValue(p ankat.ProductVar, value float64) {
	x.db.MustExec(`
INSERT OR REPLACE INTO product_value (party_id, product_serial, section, point, var, value)
VALUES ((SELECT party_id FROM current_party), ?, ?, ?, ?, ?); `, x.ProductSerial, p.Sect, p.Point, p.Var, value)
}