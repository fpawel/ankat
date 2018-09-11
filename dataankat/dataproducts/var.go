package dataproducts

import (
	"github.com/fpawel/ankat"
)

type Var struct {
	Var     ankat.Var `db:"var"`
	Checked bool      `db:"checked"`
	Ordinal int       `db:"ordinal"`
}


