package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/data/dataproducts"
	"github.com/fpawel/guartutils/comport"
	"github.com/jmoiron/sqlx"
	"time"
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

type db struct {
	dbConfig, dbProducts *sqlx.DB
}

func (x db) formatCmd(cmd uint16) (s string) {
	var xs []string
	dbMustSelect(x.dbProducts, &xs, `SELECT description FROM command WHERE command_id = $1;`, cmd)
	if len(xs) == 0 {
		s = fmt.Sprintf("команда %d", cmd)
	} else {
		s = fmt.Sprintf("%d: %s", cmd, xs[0])
	}
	return
}

func (x db) ComportSets(section string) (c comport.Config) {
	c.Serial.ReadTimeout = time.Millisecond

	q := `SELECT value FROM config WHERE section_name = ? AND property_name = ?;`
	dbMustGet(x.dbConfig, &c.Serial.Name, q, section, "port")
	dbMustGet(x.dbConfig, &c.Serial.Baud, q, section, "baud")

	dbMustGet(x.dbConfig, &c.Fetch.ReadTimeout, q, section, "timeout")
	c.Fetch.ReadTimeout *= time.Millisecond

	dbMustGet(x.dbConfig, &c.Fetch.ReadByteTimeout, q, section, "byte_timeout")
	c.Fetch.ReadByteTimeout *= time.Millisecond

	dbMustGet(x.dbConfig, &c.Fetch.MaxAttemptsRead, q, section, "repeat_count")

	dbMustGet(x.dbConfig, &c.BounceTimeout, q, section, "bounce_timeout")
	c.BounceTimeout *= time.Millisecond
	return
}

func (x db) ProductValue(partyID ankat.PartyID, serial ankat.ProductSerial, p ankat.ProductVar) (value float64, exits bool) {
	var xs []float64
	dbMustSelect(x.dbProducts, &xs, `
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

func (x db) CurrentPartyID() (currentPartyID ankat.PartyID) {
	dbMustGet(x.dbProducts, &currentPartyID, `SELECT party_id FROM current_party`)
	return
}

func (x db) CurrentPartyProductValue(serial ankat.ProductSerial, p ankat.ProductVar) (float64, bool) {
	return x.ProductValue(x.CurrentPartyID(), serial, p)
}

func (x db) SetProductValue(serial ankat.ProductSerial, p ankat.ProductVar, value float64) {
	x.dbProducts.MustExec(`
INSERT OR REPLACE INTO product_value (party_id, product_serial, section, point, var, value)
VALUES ((SELECT * FROM current_party_id), ?, ?, ?, ?, ?); `, serial, p.Sect, p.Point, p.Var, value)
}

func (x db) CurrentPartyValue(name string) float64 {
	return dataproducts.CurrentPartyValue(x.dbProducts, name)
}

//func (x db) TryCurrentPartyValueStr(name string) (string, bool) {
//	return dataproducts.TryCurrentPartyValueStr(x.dbProducts, name)
//}

func (x db) CurrentPartyValueStr(name string) (value string) {
	return dataproducts.CurrentPartyValueStr(x.dbProducts, name)
}

func (x db) IsTwoConcentrationChannels() bool {
	return x.CurrentPartyValue("sensors_count") == 2
}

func (x db) IsPressureSensor() bool {
	return x.CurrentPartyValue("pressure_sensor") == 1
}

func (x db) IsCO2() bool {
	return x.CurrentPartyValueStr("gas1") == "CO₂"
}

func (x db) CheckedVars() (vars []Var) {
	dbMustSelect(x.dbProducts, &vars, `SELECT * FROM read_var_enumerated WHERE checked = 1`)
	return
}

func (x db) VarName(v ankat.Var) (s string) {
	dbMustGet(x.dbProducts, &s, `SELECT COALESCE( ( SELECT name FROM read_var WHERE var=? ), '#' || ?);`, v, v)
	return
}

func (x db) Vars() (vars []Var) {
	dbMustSelect(x.dbProducts, &vars, `SELECT * FROM read_var_enumerated`)
	return
}

func (x db) Coefficients() (coefficients []Coefficient) {
	dbMustSelect(x.dbProducts, &coefficients, `SELECT * FROM coefficient_enumerated`)
	return
}

func (x db) CheckedCoefficients() (coefficients []Coefficient) {
	dbMustSelect(x.dbProducts, &coefficients, `SELECT * FROM coefficient_enumerated WHERE checked =1`)
	return
}

func (x db) ProductsCount() (n int) {
	dbMustGet(x.dbProducts, &n, `SELECT count(*) FROM current_party_products`)
	return
}

func (x db) Products() (products []Product) {
	dbMustSelect(x.dbProducts, &products, `SELECT * FROM current_party_products_config;`)
	return
}

func (x db) CheckedProducts() (products []Product) {
	dbMustSelect(x.dbProducts, &products,
		`SELECT * FROM current_party_products_config WHERE checked=1;`)
	return
}

func (x db) Product(n int) (p Product) {
	dbMustGet(x.dbProducts, &p, `SELECT * FROM current_party_products_config WHERE ordinal = $1;`, n)
	return
}

func (x db) SetCoefficientValue(productSerial ankat.ProductSerial, coefficient ankat.Coefficient, value float64) {
	dataproducts.SetCoefficientValue(x.dbProducts, productSerial, coefficient, value)
}

func (x db) CoefficientValue(productSerial ankat.ProductSerial, coefficient ankat.Coefficient) (float64, bool) {
	return dataproducts.CoefficientValue(x.dbProducts, productSerial, coefficient)
}

func (x db) ComportProductsBounceTimeout() time.Duration {
	var n time.Duration
	x.dbConfig.Get(&n, `SELECT value FROM config WHERE var = 'comport_products_bounce_timeout';`)
	return n * time.Millisecond
}

func (x db) ConfigDuration(section, property string) time.Duration {
	return time.Duration( x.ConfigValue(section, property ) )
}

func (x db) ConfigValue(section, property string) float64 {
	var xs []float64
	dbMustSelect(x.dbConfig, &xs,
		`SELECT value FROM config WHERE section_name = ? AND config.property_name = ?;`,
		section, property)

	if len(xs) > 0 {
		return xs[0]
	}
	var v float64
	dbMustGet(x.dbConfig, &v,
		`SELECT default_value FROM property WHERE property_name = ?`,
		property)
	return v
}

func (x db) PartyInfo(partyID ankat.PartyID) dataproducts.PartyInfo {
	return dataproducts.GetPartyInfo(x.dbProducts, partyID)
}

func (x db) CurrentPartyInfo() dataproducts.PartyInfo {
	return dataproducts.GetPartyInfo(x.dbProducts, x.CurrentPartyID())
}
