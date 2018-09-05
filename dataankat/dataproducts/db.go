package dataproducts

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
	"github.com/jmoiron/sqlx"
)



type Product struct {
	Checked bool                `db:"checked"`
	Comport string              `db:"comport"`
	Serial  ankat.ProductSerial `db:"product_serial"`
	Ordinal int                 `db:"ordinal"`
}

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

type ScalePosition string
type TemperaturePosition string

const (
	ConcentrationScaleBegin ScalePosition = "B"
	ConcentrationScaleMidle ScalePosition = "M"
	ConcentrationScaleEnd ScalePosition = "E"

	HighTemperature TemperaturePosition = "H"
	LowTemperature TemperaturePosition = "L"
	NormalTemperature TemperaturePosition = "N"
)

func MustOpen(fileName string) DBProducts {
	return DBProducts{dbutils.MustOpen(fileName, "sqlite3", SQLAnkat) }
}

func (x DBProducts) PartyExists() (exists bool){
	dbutils.MustGet(x.DB, &exists, `SELECT exists(SELECT party_id FROM party);`)
	return
}

func (x DBProducts) FormatCmd(cmd uint16) (s string) {
	var xs []string
	dbutils.MustSelect(x.DB, &xs, `SELECT description FROM command WHERE command_id = $1;`, cmd)
	if len(xs) == 0 {
		s = fmt.Sprintf("команда %d", cmd)
	} else {
		s = fmt.Sprintf("%d: %s", cmd, xs[0])
	}
	return
}

func (x DBProducts) ProductValue(partyID ankat.PartyID, serial ankat.ProductSerial, p ankat.ProductVar) (value float64, exits bool) {
	var xs []float64
	dbutils.MustSelect(x.DB, &xs, `
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


func (x DBProducts) PartyInfo(partyID ankat.PartyID) (party PartyInfo) {
	dbutils.MustGet(x.DB, &party, `
SELECT party_id, created_at FROM party WHERE party_id = $1;`, partyID)

	dbutils.MustSelect(x.DB, &party.Products, `
SELECT product_serial
FROM product
WHERE party_id = $1
ORDER BY product_serial ASC;`, partyID)

	dbutils.MustSelect(x.DB, &party.Values, `SELECT name, value FROM party_value2 WHERE party_id = ?;`, partyID)

	var coefficients []struct {
		Coefficient   ankat.Coefficient   `db:"coefficient_id"`
		ProductSerial ankat.ProductSerial `db:"product_serial"`
		Value         float64             `db:"value"`
	}
	dbutils.MustSelect(x.DB, &coefficients, `
SELECT coefficient_id, product_serial, value FROM product_coefficient_value WHERE party_id = ?;
`, partyID)

	for _, k := range coefficients {
		if len(party.Coefficients) == 0 {
			party.Coefficients = make(Coefficients)
		}
		if _, f := party.Coefficients[k.Coefficient]; !f {
			party.Coefficients[k.Coefficient] = make(map[ankat.ProductSerial]float64)
		}
		party.Coefficients[k.Coefficient][k.ProductSerial] = k.Value
	}

	var productVarValues []struct {
		Sect   ankat.Sect          `db:"section"`
		Var    ankat.Var           `db:"var"`
		Point ankat.Point			`db:"point"`
		Serial ankat.ProductSerial `db:"product_serial"`
		Value  float64             `db:"value"`
	}
	dbutils.MustSelect(x.DB, &productVarValues, `
SELECT section, var, point, product_serial, value 
FROM product_value
WHERE party_id = ?;
`, partyID)

	for _, k := range productVarValues {
		if len(party.ProductVarValues) == 0 {
			party.ProductVarValues = make(ProductVarValues)
		}
		if _, f := party.ProductVarValues[k.Sect]; !f {
			party.ProductVarValues[k.Sect] = make(map[ankat.Var]map[ankat.Point]map[ankat.ProductSerial]float64)
		}
		if _, f := party.ProductVarValues[k.Sect][k.Var]; !f {
			party.ProductVarValues[k.Sect][k.Var] = make(map[ankat.Point]map[ankat.ProductSerial]float64)
		}
		if _, f := party.ProductVarValues[k.Sect][k.Var][k.Point]; !f {
			party.ProductVarValues[k.Sect][k.Var][k.Point] = make(map[ankat.ProductSerial]float64)
		}
		if _, f := party.ProductVarValues[k.Sect][k.Var][k.Point]; !f {
			party.ProductVarValues[k.Sect][k.Var][k.Point] = make(map[ankat.ProductSerial]float64)
		}
		party.ProductVarValues[k.Sect][k.Var][k.Point][k.Serial] = k.Value
	}

	return
}

func (x DBProducts) CurrentParty() DBCurrentParty  {
	return DBCurrentParty{x}
}

