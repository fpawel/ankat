package products

import "github.com/fpawel/ankat/internal/ankat"

type ProductCoefficientValue struct {
	ProductSerial ankat.ProductSerial
	Coefficient ankat.Coefficient
	Value float64
}

type Coefficient struct {
	Coefficient ankat.Coefficient `db:"coefficient_id"`
	Checked     bool              `db:"checked"`
	Ordinal     int               `db:"ordinal"`
	Name        string            `db:"name"`
	Description string            `db:"description"`
}

type CurrentProduct struct {
	Product
	Checked       bool                `db:"checked"`
	Comport       string              `db:"comport"`
	Ordinal       int                 `db:"ordinal"`
}
