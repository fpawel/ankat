package dataproducts

import (
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
	"github.com/jmoiron/sqlx"
)



type Var struct {
	Var     ankat.Var `db:"var"`
	Checked bool      `db:"checked"`
	Ordinal int       `db:"ordinal"`
}

type Coefficient struct {
	Coefficient ankat.Coefficient `db:"coefficient_id"`
	Checked     bool              `db:"checked"`
	Ordinal     int               `db:"ordinal"`
}



func MustOpen(fileName string) (db *sqlx.DB) {
	db = dbutils.MustOpen(fileName, "sqlite3", )
	db.MustExec(SQLAnkat)
	return
}

func PartyExists(db *sqlx.DB) (exists bool) {
	dbutils.MustGet(db, &exists, `SELECT exists(SELECT party_id FROM party);`)
	return
}

func productValue(db *sqlx.DB, partyID ankat.PartyID, serial ankat.ProductSerial, p ankat.ProductVar) (value float64, exits bool) {
	var xs []float64
	dbutils.MustSelect(db, &xs, `
SELECT value FROM product_value 
WHERE party_id = ? AND product_serial=? AND var = ? AND section = ? AND point = ?;`,
		partyID, serial, p.Var, p.Sect, p.Point)
	if len(xs) == 0 {
		return
	}
	if len(xs) > 1 {
		panic("len must be 1 or 0")
	}
	value = xs[0]
	exits = true
	return
}

func CheckedVars(db *sqlx.DB) (vars []Var) {
	dbutils.MustSelect(db, &vars, `SELECT * FROM read_var_enumerated WHERE checked = 1`)
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

func Coefficients(db *sqlx.DB) (coefficients []Coefficient) {
	dbutils.MustSelect(db, &coefficients, `SELECT * FROM coefficient_enumerated`)
	return
}

func CheckedCoefficients(db *sqlx.DB) (coefficients []Coefficient) {
	dbutils.MustSelect(db, &coefficients, `SELECT * FROM coefficient_enumerated WHERE checked =1`)
	return
}

func GetParty(db *sqlx.DB, partyID ankat.PartyID) (party Party) {
	party.db = db
	dbutils.MustGet(db, &party, `SELECT * FROM party WHERE party_id = $1;`, partyID)
	return
}

func GetCurrentParty(db *sqlx.DB) (party CurrentParty) {
	party.db = db
	dbutils.MustGet(db, &party, `SELECT * FROM current_party ;`)
	return
}
