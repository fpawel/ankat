package dataproducts

import (
	"github.com/fpawel/ankat"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func SetCoefficientValue(x *sqlx.DB, productSerial ankat.ProductSerial, coefficient ankat.Coefficient, value float64) {
	x.MustExec(`
INSERT OR REPLACE INTO product_coefficient_value (party_id, product_serial, coefficient_id, value)
VALUES ((SELECT * FROM current_party_id),
        $1, $2, $3); `, productSerial, coefficient, value)
}

func CoefficientValue(x *sqlx.DB, productSerial ankat.ProductSerial, coefficient ankat.Coefficient) (value float64, exits bool) {
	var xs []float64
	dbMustSelect(x, &xs, `
SELECT value FROM current_party_coefficient_value 
WHERE product_serial=$1 AND coefficient_id = $2;`, productSerial, coefficient)
	if len(xs) > 0 {
		value = xs[0]
		exits = true
	}
	return
}

func CurrentPartyValue(x *sqlx.DB, name string) (value float64) {
	dbMustGet(x, &value, `
SELECT value FROM party_value 
WHERE var=$1 AND party_id IN ( SELECT * FROM current_party_id);`, name)
	return
}

//func TryCurrentPartyValueStr(x *sqlx.DB, name string) (value string, exists bool) {
//	var s []string
//	dbMustSelect(x, &s, `SELECT value FROM party_value WHERE var=$1 AND party_id IN ( SELECT * FROM current_party_id);`, name)
//	if len(s) == 1 {
//		value = s[0]
//		exists = true
//	}
//	return
//}
//
//func TryCurrentPartyValue(x *sqlx.DB, name string) (value float64, exists bool) {
//	var v []float64
//	dbMustSelect(x, &v, `SELECT value FROM party_value WHERE var=$1 AND party_id IN ( SELECT * FROM current_party_id);`, name)
//	if len(v) == 1 {
//		value = v[0]
//		exists = true
//	}
//	return
//}

func CurrentPartyValueStr(x *sqlx.DB, name string) (value string) {
	dbMustGet(x, &value, `
SELECT value FROM party_value 
WHERE var=$1 AND party_id IN ( SELECT * FROM current_party_id);`, name)
	return
}

func dbMustGet(db *sqlx.DB, dest interface{}, query string, args ...interface{}) {
	if err := db.Get(dest, query, args...); err != nil {
		panic(err)
	}
}

func dbMustSelect(db *sqlx.DB, dest interface{}, query string, args ...interface{}) {
	if err := db.Select(dest, query, args...); err != nil {
		panic(err)
	}
}
