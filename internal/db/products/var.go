package products

import (
	"github.com/fpawel/ankat/internal/ankat"
)

type Var struct {
	Var         ankat.Var `db:"var"`
	Checked     bool      `db:"checked"`
	Ordinal     int       `db:"ordinal"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}
