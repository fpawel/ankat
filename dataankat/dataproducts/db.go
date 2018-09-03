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

func (x DBProducts) CurrentPartyID() (currentPartyID ankat.PartyID) {
	dbutils.MustGet(x.DB, &currentPartyID, `SELECT party_id FROM current_party`)
	return
}

func (x DBProducts) CurrentPartyProductValue(serial ankat.ProductSerial, p ankat.ProductVar) (float64, bool) {
	return x.ProductValue(x.CurrentPartyID(), serial, p)
}

func (x DBProducts) SetCurrentPartyProductValue(serial ankat.ProductSerial, p ankat.ProductVar, value float64) {
	x.DB.MustExec(`
INSERT OR REPLACE INTO product_value (party_id, product_serial, section, point, var, value)
VALUES ((SELECT party_id FROM current_party), ?, ?, ?, ?, ?); `, serial, p.Sect, p.Point, p.Var, value)
}


func (x DBProducts) CurrentPartyValue(name string) ( value float64) {
	dbutils.MustGet(x.DB, &value, "SELECT " + name +" FROM current_party;")
	return
}

func (x DBProducts) CurrentPartyValueStr(name string) (value string) {
	dbutils.MustGet(x.DB, &value, "SELECT " + name +" FROM current_party;" )
	return
}

func (x DBProducts) CurrentPartyScaleCode(c ankat.AnkatChan) float64{
	scale := x.CurrentPartyValue( fmt.Sprintf("scale%d", c))
	switch scale {
	case 2:
		return 2
	case 5:
		return 6
	case 10:
		return 7
	case 100:
		return 21
	}
	panic(fmt.Sprintf("unknown scale: %v: %v", c, scale ))
}

func (x DBProducts) CurrentPartyUnitsCode(c ankat.AnkatChan) float64{
	units := x.CurrentPartyValueStr( fmt.Sprintf("units%d", c))
	switch units {
	case "объемная доля, %":
		return 3
	case "%, НКПР":
		return 4
	}
	panic(fmt.Sprintf("unknown units: %v, %v", c, units ))
}

func (x DBProducts) CurrentPartyGasTypeCode(c ankat.AnkatChan) float64{
	gas := x.CurrentPartyValueStr( fmt.Sprintf("gas%d", c))
	scale := x.CurrentPartyValue(fmt.Sprintf("scale%d", c))
	switch gas {
	case "CH₄":
		return 16
	case "C₃H₈":
		return 14
	case "∑CH":
		return 15
	case "CO₂":
		switch scale {
		case 2:
			return 11
		case 5:
			return 12
		case 10:
			return 13
		}
	}
	panic(fmt.Sprintf("unknown gas and scale: %v: %v, %v", c, gas, scale ))
}

func (x DBProducts) AnkatChannels() (xs []ankat.AnkatChan){
	xs = append(xs, ankat.Chan1)
	if x.IsTwoConcentrationChannels() {
		xs = append(xs, ankat.Chan2)
	}
	return
}

func (x DBProducts) IsTwoConcentrationChannels() bool {
	return x.CurrentPartyValue("sensors_count") == 2
}

func (x DBProducts) IsPressureSensor() bool {
	return x.CurrentPartyValue("pressure_sensor") == 1
}

func (x DBProducts) IsCO2() bool {
	return x.CurrentPartyValueStr("gas1") == "CO₂"
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

func (x DBProducts) ProductsCount() (n int) {
	dbutils.MustGet(x.DB, &n, `SELECT count(*) FROM current_party_products`)
	return
}

func (x DBProducts) Products() (products []Product) {
	dbutils.MustSelect(x.DB, &products, `SELECT * FROM current_party_products_config;`)
	return
}

func (x DBProducts) CheckedProducts() (products []Product) {
	dbutils.MustSelect(x.DB, &products,
		`SELECT * FROM current_party_products_config WHERE checked=1;`)
	return
}

func (x DBProducts) Product(n int) (p Product) {
	dbutils.MustGet(x.DB, &p, `SELECT * FROM current_party_products_config WHERE ordinal = $1;`, n)
	return
}

func (x DBProducts) SetCoefficientValue(productSerial ankat.ProductSerial, coefficient ankat.Coefficient, value float64) {
	x.DB.MustExec(`
INSERT OR REPLACE INTO product_coefficient_value (party_id, product_serial, coefficient_id, value)
VALUES ((SELECT party_id FROM current_party),
        $1, $2, $3); `, productSerial, coefficient, value)
}

func (x DBProducts) CoefficientValue(productSerial ankat.ProductSerial, coefficient ankat.Coefficient) (float64, bool) {
	var xs []float64
	dbutils.MustSelect(x.DB, &xs, `
SELECT value FROM current_party_coefficient_value 
WHERE product_serial=$1 AND coefficient_id = $2;`, productSerial, coefficient)
	if len(xs) > 0 {
		return xs[0], true
	}
	return 0, false
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

func (x DBProducts) CurrentPartyInfo() PartyInfo {
	return x.PartyInfo(x.CurrentPartyID())
}

func (x DBProducts) CurrentPartyVerificationGasConcentration(gas ankat.GasCode) float64 {
	return x.CurrentPartyValue(fmt.Sprintf("cgas%d", gas))
}