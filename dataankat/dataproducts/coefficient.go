package dataproducts

import (
	"github.com/fpawel/ankat"
)

type Coefficient struct {
	Coefficient ankat.Coefficient `db:"coefficient_id"`
	Checked     bool              `db:"checked"`
	Ordinal     int               `db:"ordinal"`
	Name        string            `db:"name"`
	Description string            `db:"description"`
}
