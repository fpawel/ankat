package dataproducts

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
)

type DBCurrentParty struct {
	DB DBProducts
}

func (x DBCurrentParty) ID() (id ankat.PartyID) {
	dbutils.MustGet(x.DB.DB, &id, `SELECT party_id FROM current_party`)
	return
}

func (x DBCurrentParty) ProductValue(serial ankat.ProductSerial, p ankat.ProductVar) (value float64, exits bool) {
	return x.DB.ProductValue(x.ID(), serial, p)
}

func (x DBCurrentParty) Value(name string) (value float64) {
	dbutils.MustGet(x.DB.DB, &value, "SELECT "+name+" FROM current_party;")
	return
}

func (x DBCurrentParty) ValueStr(name string) (value string) {
	dbutils.MustGet(x.DB.DB, &value, "SELECT "+name+" FROM current_party;")
	return
}

func (x DBCurrentParty) ScaleCode(c ankat.AnkatChan) float64 {
	scale := x.Value(fmt.Sprintf("scale%d", c))
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
	panic(fmt.Sprintf("unknown scale: %v: %v", c, scale))
}

func (x DBCurrentParty) UnitsCode(c ankat.AnkatChan) float64 {
	units := x.ValueStr(fmt.Sprintf("units%d", c))
	switch units {
	case "объемная доля, %":
		return 3
	case "%, НКПР":
		return 4
	}
	panic(fmt.Sprintf("unknown units: %v, %v", c, units))
}

func (x DBCurrentParty) GasTypeCode(c ankat.AnkatChan) float64 {
	gas := x.ValueStr(fmt.Sprintf("gas%d", c))
	scale := x.Value(fmt.Sprintf("scale%d", c))
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
	panic(fmt.Sprintf("unknown gas and scale: %v: %v, %v", c, gas, scale))
}

func (x DBCurrentParty) AnkatChannels() (xs []ankat.AnkatChan) {
	xs = append(xs, ankat.Chan1)
	if x.IsTwoConcentrationChannels() {
		xs = append(xs, ankat.Chan2)
	}
	return
}

func (x DBCurrentParty) IsTwoConcentrationChannels() bool {
	return x.Value("sensors_count") == 2
}

func (x DBCurrentParty) IsPressureSensor() bool {
	return x.Value("pressure_sensor") == 1
}

func (x DBCurrentParty) IsCO2() bool {
	return x.ValueStr("gas1") == string(ankat.GasCO2)
}

func (x DBCurrentParty) ProductsCount() (n int) {
	dbutils.MustGet(x.DB.DB, &n, `SELECT count(*) FROM current_party_products`)
	return
}

func (x DBCurrentParty) Products() (products []Product) {
	dbutils.MustSelect(x.DB.DB, &products, `SELECT * FROM current_party_products_config;`)
	return
}

func (x DBCurrentParty) CheckedProducts() (products []Product) {
	dbutils.MustSelect(x.DB.DB, &products,
		`SELECT * FROM current_party_products_config WHERE checked=1;`)
	return
}

func (x DBCurrentParty) Product(n int) (p Product) {
	dbutils.MustGet(x.DB.DB, &p, `SELECT * FROM current_party_products_config WHERE ordinal = $1;`, n)
	return
}

func (x DBCurrentParty) Info() PartyInfo {
	return x.DB.PartyInfo(x.ID())
}

func (x DBCurrentParty) VerificationGasConcentration(gas ankat.GasCode) float64 {
	return x.Value(fmt.Sprintf("cgas%d", gas))
}

func (x DBCurrentParty) SetMainErrorConcentrationValue(ankatChan ankat.AnkatChan, scalePos ankat.ScalePosition,
	temperaturePos ankat.TemperaturePosition, serial ankat.ProductSerial, value float64) {

}
