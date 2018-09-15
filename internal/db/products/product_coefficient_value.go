package products

import "github.com/fpawel/ankat/internal/ankat"

type ProductCoefficientValue struct {
	ProductSerial ankat.ProductSerial
	Coefficient ankat.Coefficient
	Value float64
}
