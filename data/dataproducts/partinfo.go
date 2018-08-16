package dataproducts

import (
	"github.com/fpawel/ankat"
	"time"
	"github.com/jmoiron/sqlx"
	"sort"
)

type PartyInfo struct {
	PartyID      ankat.PartyID         `db:"party_id"`
	CreatedAt    time.Time             `db:"created_at"`
	Products     []ankat.ProductSerial `db:"-"`
	Values       []KeyValue            `db:"-"`
	Coefficients Coefficients          `db:"-"`
}

type Coefficients map[ankat.Coefficient]map[ankat.ProductSerial]float64

type KeyValue struct {
	Key   string `db:"name"`
	Value string `db:"value"`
}

func GetPartyInfo(x *sqlx.DB, partyID ankat.PartyID) (party PartyInfo) {
	dbMustGet(x, &party, `
SELECT party_id, created_at FROM party WHERE party_id = $1;`, partyID)

	dbMustSelect(x, &party.Products, `
SELECT product_serial
FROM product
WHERE party_id = $1
ORDER BY product_serial ASC;`, partyID)

	dbMustSelect(x, &party.Values, `SELECT name, value FROM party_value2 WHERE party_id = ?;`, partyID)

	var coefficients []struct {
		Coefficient   ankat.Coefficient   `db:"coefficient_id"`
		ProductSerial ankat.ProductSerial `db:"product_serial"`
		Value         float64             `db:"value"`
	}
	dbMustSelect(x, &coefficients, `
SELECT coefficient_id, product_serial, value FROM product_coefficient_value WHERE party_id = ?;
`, partyID)

	for _, k := range coefficients {
		if len(party.Coefficients) == 0 {
			party.Coefficients = make(Coefficients)
		}
		if _,f := party.Coefficients[k.Coefficient]; !f {
			party.Coefficients[k.Coefficient] = make(map[ankat.ProductSerial]float64)
		}
		party.Coefficients[k.Coefficient][k.ProductSerial] = k.Value
	}

	return
}

func (x Coefficients) Coefficients() ( coefficients []ankat.Coefficient) {
	for coefficient := range x {
		coefficients = append(coefficients, coefficient)
	}
	return
}

func (x Coefficients) Products() ( products []ankat.ProductSerial) {
	for _,ps := range x {
		for p := range ps {
			products = append(products, p)
		}
	}
	sort.Slice(products, func(i, j int) bool {
		return products[i] < products[j]
	})

	return
}