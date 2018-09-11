package dataproducts

import (
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
	"github.com/jmoiron/sqlx"
)


func MustOpen(fileName string) (db *sqlx.DB) {
	db = dbutils.MustOpen(fileName, "sqlite3", )
	db.MustExec(SQLAnkat)
	return
}

func PartyExists(db *sqlx.DB) (exists bool) {
	dbutils.MustGet(db, &exists, `SELECT exists(SELECT party_id FROM party);`)
	return
}

func GetCurrentParty(db *sqlx.DB) (party Party) {
	party.db = db
	dbutils.MustGet(db, &party, `SELECT * FROM current_party ;`)
	return
}

func GetParty(db *sqlx.DB, partyID ankat.PartyID) (party Party) {
	party.db = db
	dbutils.MustGet(db, &party, `SELECT * FROM party WHERE party_id = $1;`, partyID)
	return
}

func CurrentProducts(db *sqlx.DB) (products []CurrentProduct) {
	dbutils.MustSelect(db, &products, `SELECT * FROM current_party_products_config;`)
	for i := range products{
		products[i].db = db
		products[i].PartyID = GetCurrentParty(db).PartyID
	}
	return
}

func CurrentProductOrderBySerial(db *sqlx.DB, productSerial ankat.ProductSerial ) int{
	var xs []int
	dbutils.MustSelect(db, &xs, `
SELECT ordinal
FROM current_party_products_config
WHERE product_serial = ?;`, productSerial)
	if len(xs) == 0 {
		return -1
	}
	return xs[0]
}

func CheckedProducts(db *sqlx.DB) (products []CurrentProduct) {
	dbutils.MustSelect(db, &products,
		`SELECT * FROM current_party_products_config WHERE checked=1;`)
	for i := range products{
		products[i].db = db
		products[i].PartyID = GetCurrentParty(db).PartyID
	}
	return
}

func GetCurrentProduct(db *sqlx.DB, n int) (p CurrentProduct) {
	dbutils.MustGet(db, &p, `SELECT * FROM current_party_products_config WHERE ordinal = $1;`, n)
	p.PartyID = GetCurrentParty(db).PartyID
	p.db = db
	return
}

func VarName(db *sqlx.DB, v ankat.Var) (s string) {
	dbutils.MustGet(db, &s, `SELECT COALESCE( ( SELECT name FROM read_var WHERE var=? ), '#' || ?);`, v, v)
	return
}

func Vars(db *sqlx.DB) (vars []Var) {
	dbutils.MustSelect(db, &vars, `SELECT * FROM read_var_enumerated`)
	return
}

func CheckedVars(db *sqlx.DB) (vars []Var) {
	dbutils.MustSelect(db, &vars, `SELECT * FROM read_var_enumerated WHERE checked = 1`)
	return
}

func Coefficients(db *sqlx.DB) (coefficients []Coefficient) {
	dbutils.MustSelect(db, &coefficients, `SELECT * FROM coefficient_enumerated`)
	return
}

func CheckedCoefficients(db *sqlx.DB) (coefficients []Coefficient) {
	dbutils.MustSelect(db, &coefficients, `SELECT * FROM coefficient_enumerated WHERE checked =1`)
	return
}
