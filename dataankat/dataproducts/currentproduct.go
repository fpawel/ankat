package dataproducts

import (
	"github.com/fpawel/ankat"
)

type CurrentProduct struct {
	Product
	Checked       bool                `db:"checked"`
	Comport       string              `db:"comport"`
	Ordinal       int                 `db:"ordinal"`
}

func (x CurrentProduct) SetCoefficientValue(coefficient ankat.Coefficient, value float64) {
	x.db.MustExec(`
INSERT OR REPLACE INTO product_coefficient_value (party_id, product_serial, coefficient_id, value)
VALUES ((SELECT party_id FROM current_party),
        $1, $2, $3); `, x.ProductSerial, coefficient, value)
}



func (x CurrentProduct) SetSectCoefficients(sect ankat.Sect, values []float64) {
	for i := range values {
		x.SetCoefficientValue(sect.Coefficient0()+ankat.Coefficient(i), values[i])
	}
}


func (x CurrentProduct) SetValue(p ankat.ProductVar, value float64) {
	x.db.MustExec(`
INSERT OR REPLACE INTO product_value (party_id, product_serial, section, point, var, value)
VALUES ((SELECT party_id FROM current_party), ?, ?, ?, ?, ?); `, x.ProductSerial, p.Sect, p.Point, p.Var, value)
}
