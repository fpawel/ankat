package dataproducts

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type PartyID int64

type Party struct {
	PartyID     PartyID   `db:"party_id"`
	CreatedAt   time.Time `db:"created_at"`
	ProductType string    `db:"product_type"`
	Products    []int     `db:"-"`
}

type KeyValue struct {
	Key, Value interface{}
}

func SetCoefficientValue(x *sqlx.DB, productSerial, coefficient int, value float64) {
	x.MustExec(`
INSERT OR REPLACE INTO product_coefficient_value (party_id, product_serial, coefficient_id, value)
VALUES ((SELECT * FROM current_party_id),
        $1, $2, $3); `, productSerial, coefficient, value)
}

func CoefficientValue(x *sqlx.DB, productSerial, coefficient int) (value sql.NullFloat64) {
	dbMustGet(x, &value, `
SELECT value FROM current_party_coefficient_value 
WHERE product_serial=$1 AND coefficient_id = $2;`, productSerial, coefficient)
	return
}

func CurrentPartyValue(x *sqlx.DB, name string) (value float64) {
	dbMustGet(x, &value, `
SELECT value FROM party_value 
WHERE var=$1 AND party_id IN ( SELECT * FROM current_party_id);`, name)
	return
}

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
