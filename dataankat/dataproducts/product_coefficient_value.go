package dataproducts

import "github.com/fpawel/ankat"

type ProductCoefficientValue struct {
	ProductSerial ankat.ProductSerial
	Coefficient ankat.Coefficient
	Value float64
}
