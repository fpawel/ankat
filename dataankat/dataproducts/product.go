package dataproducts

import (
	"github.com/fpawel/ankat"
	"github.com/jmoiron/sqlx"
)

type Product struct {
	PartyID       ankat.PartyID       `db:"party_id"`
	ProductSerial ankat.ProductSerial `db:"product_serial"`
	db            *sqlx.DB
}
