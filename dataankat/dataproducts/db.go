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

type DBProducts struct {
	DB *sqlx.DB
}

func MustOpen(fileName string) (db DBProducts) {
	db = DBProducts{dbutils.MustOpen(fileName, "sqlite3", )}
	db.DB.MustExec(SQLAnkat)
	return
}

func (x DBProducts) PartyExists() (exists bool) {
	dbutils.MustGet(x.DB, &exists, `SELECT exists(SELECT party_id FROM party);`)
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

func (x DBProducts) CheckedVars() (vars []Var) {
	dbutils.MustSelect(x.DB, &vars, `SELECT * FROM read_var_enumerated WHERE checked = 1`)
	return
}

func (x DBProducts) VarName(v ankat.Var) (s string) {
	dbutils.MustGet(x.DB, &s, `SELECT COALESCE( ( SELECT name FROM read_var WHERE var=? ), '#' || ?);`, v, v)
	return
}

func (x DBProducts) Vars() (vars []Var) {
	dbutils.MustSelect(x.DB, &vars, `SELECT * FROM read_var_enumerated`)
	return
}

func (x DBProducts) Coefficients() (coefficients []Coefficient) {
	dbutils.MustSelect(x.DB, &coefficients, `SELECT * FROM coefficient_enumerated`)
	return
}

func (x DBProducts) CheckedCoefficients() (coefficients []Coefficient) {
	dbutils.MustSelect(x.DB, &coefficients, `SELECT * FROM coefficient_enumerated WHERE checked =1`)
	return
}

func (x DBProducts) Party(partyID ankat.PartyID) (party Party) {
	party.db = x.DB
	dbutils.MustGet(x.DB, &party, `SELECT * FROM party WHERE party_id = $1;`, partyID)
	return
}

func (x DBProducts) CurrentParty() (party CurrentParty) {
	party.db = x.DB
	dbutils.MustGet(x.DB, &party, `SELECT * FROM current_party ;`)
	return
}
