package main

import (
	"fmt"
	"github.com/fpawel/ankat/data/dataproducts"
	"github.com/fpawel/guartutils/comport"
	"github.com/jmoiron/sqlx"
	"time"
)

type Product struct {
	Checked bool   `db:"checked"`
	Comport string `db:"comport"`
	Serial  int    `db:"product_serial"`
	Ordinal int    `db:"ordinal"`
}

type Var struct {
	Var     int  `db:"var"`
	Checked bool `db:"checked"`
	Ordinal int  `db:"ordinal"`
}

type Coefficient struct {
	Coefficient int  `db:"coefficient_id"`
	Checked     bool `db:"checked"`
	Ordinal     int  `db:"ordinal"`
}

type data struct {
	dbConfig, dbProducts *sqlx.DB
}

func (x data) formatCmd(cmd uint16) (s string) {
	var xs []string
	dbMustSelect(x.dbProducts, &xs, `SELECT description FROM command WHERE command_id = $1;`, cmd)
	if len(xs) == 0 {
		s = fmt.Sprintf("команда %d", cmd)
	} else {
		s = fmt.Sprintf("%d: %s", cmd, xs[0])
	}
	return
}

func (x data) ComportSets(id string) (c comport.Config) {
	c.Serial.ReadTimeout = time.Millisecond

	s := "comport_" + id
	q := `SELECT value FROM config WHERE var = $1;`
	dbMustGet(x.dbConfig, &c.Serial.Name, q, s)
	dbMustGet(x.dbConfig, &c.Serial.Baud, q, s+"_baud")

	dbMustGet(x.dbConfig, &c.Fetch.ReadTimeout, q, s+"_timeout")
	c.Fetch.ReadTimeout *= time.Millisecond

	dbMustGet(x.dbConfig, &c.Fetch.ReadByteTimeout, q, s+"_byte_timeout")
	c.Fetch.ReadByteTimeout *= time.Millisecond

	dbMustGet(x.dbConfig, &c.Fetch.MaxAttemptsRead, q, s+"_repeat_count")

	dbMustGet(x.dbConfig, &c.BounceTimeout, q, s+"_bounce_timeout")
	c.BounceTimeout *= time.Millisecond

	dbMustGet(x.dbConfig, &c.BounceLimit, q, s+"_bounce_limit")

	return
}

func (x data) EnsurePartyExists() {
	dataproducts.EnsurePartyExists(x.dbProducts)
}

func (x data) CurrentPartyValue(name string) float64 {
	return dataproducts.CurrentPartyValue(x.dbProducts, name)
}

func (x data) CurrentPartyValueStr(name string) (value string) {
	return dataproducts.CurrentPartyValueStr(x.dbProducts, name)
}

func (x data) IsTwoConcentrationChannels() bool {
	return x.CurrentPartyValue("sensors_count") == 2
}

func (x data) CheckedVars() (vars []Var) {
	dbMustSelect(x.dbProducts, &vars, `SELECT * FROM read_var_enumerated WHERE checked = 1`)
	return
}

func (x data) Vars() (vars []Var) {
	dbMustSelect(x.dbProducts, &vars, `SELECT * FROM read_var_enumerated`)
	return
}

func (x data) Coefficients() (coefficients []Coefficient) {
	dbMustSelect(x.dbProducts, &coefficients, `SELECT * FROM coefficient_enumerated`)
	return
}

func (x data) CheckedCoefficients() (coefficients []Coefficient) {
	dbMustSelect(x.dbProducts, &coefficients, `SELECT * FROM coefficient_enumerated WHERE checked =1`)
	return
}

func (x data) ProductsCount() (n int) {
	dbMustGet(x.dbProducts, &n, `SELECT count(*) FROM current_party_products`)
	return
}

func (x data) Products() (products []Product) {
	dbMustSelect(x.dbProducts, &products, `SELECT * FROM current_party_products_config;`)
	return
}

func (x data) CheckedProducts() (products []Product) {
	dbMustSelect(x.dbProducts, &products,
		`SELECT * FROM current_party_products_config WHERE checked=1;`)
	return
}

func (x data) Product(n int) (p Product) {
	dbMustGet(x.dbProducts, &p, `SELECT * FROM current_party_products_config WHERE ordinal = $1;`, n)
	return
}

func (x data) SetCoefficientValue(productSerial, coefficient int, value float64) {
	dataproducts.SetCoefficientValue(x.dbProducts, productSerial, coefficient, value)
}

func (x data) CoefficientValue(productSerial, coefficient int) (float64, bool) {
	return dataproducts.CoefficientValue(x.dbProducts, productSerial, coefficient)
}

func (x data) ComportProductsBounceTimeout() time.Duration {
	var n time.Duration
	x.dbConfig.Get(&n, `SELECT value FROM config WHERE var = 'comport_products_bounce_timeout';`)
	return n * time.Millisecond
}

func (x data) ConfigDuration(name string) time.Duration {
	var n time.Duration
	x.dbConfig.Get(&n, `SELECT value FROM config WHERE var = ?;`, name)
	return n
}

func (x data) ConfigValue(name string) float64 {
	var n float64
	x.dbConfig.Get(&n, `SELECT value FROM config WHERE var = ?;`, name)
	return n
}
