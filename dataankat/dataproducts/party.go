package dataproducts

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
	"github.com/jmoiron/sqlx"
	"time"
)

type Party struct {
	PartyID           ankat.PartyID `db:"party_id"`
	CreatedAt         time.Time     `db:"created_at"`
	ProductTypeNumber int           `db:"product_type_number"`
	SensorsCount      int           `db:"sensors_count"`
	PressureSensor    bool          `db:"pressure_sensor"`

	CGas1 float64 `db:"cgas1"`
	CGas2 float64 `db:"cgas2"`
	CGas3 float64 `db:"cgas3"`
	CGas4 float64 `db:"cgas4"`
	CGas5 float64 `db:"cgas5"`
	CGas6 float64 `db:"cgas6"`

	TemperatureMinus float64 `db:"temperature_minus"`
	TemperaturePlus  float64 `db:"temperature_plus"`

	Gas1   ankat.GasComponent       `db:"gas1"`
	Gas2   ankat.GasComponent       `db:"gas2"`
	Units1 ankat.ConcentrationUnits `db:"units1"`
	Units2 ankat.ConcentrationUnits `db:"units2"`
	Scale1 ankat.Scale              `db:"scale1"`
	Scale2 ankat.Scale              `db:"scale2"`

	db *sqlx.DB
}



func (x Party) What() string {
	s := fmt.Sprintf("%d, %s%d", x.ProductTypeNumber, x.Gas1, x.Scale1)
	if x.IsTwoConcentrationChannels() {
		s += fmt.Sprintf(", %s%d", x.Gas2, x.Scale2)
	}
	return s
}

func (x Party) IsTwoConcentrationChannels() bool {
	return x.SensorsCount == 2
}


func (x Party) IsCO2() bool {
	return x.Gas1 == ankat.GasCO2
}

func (x Party) ProductSerials() (productSerials []ankat.ProductSerial) {
	dbutils.MustSelect(x.db, &productSerials, `
SELECT product_serial
FROM product
WHERE party_id = $1
ORDER BY product_serial ASC;`, x.PartyID)
	return
}

func (x Party) Coefficients() (result ankat.PartyCoefficients) {
	var coefficients []struct {
		Coefficient   ankat.Coefficient   `db:"coefficient_id"`
		ProductSerial ankat.ProductSerial `db:"product_serial"`
		Value         float64             `db:"value"`
	}
	dbutils.MustSelect(x.db, &coefficients, `
SELECT coefficient_id, product_serial, value FROM product_coefficient_value WHERE party_id = ?;
`, x.PartyID)

	for _, k := range coefficients {
		if len(result) == 0 {
			result = make(ankat.PartyCoefficients)
		}
		if _, f := result[k.Coefficient]; !f {
			result[k.Coefficient] = make(map[ankat.ProductSerial]float64)
		}
		result[k.Coefficient][k.ProductSerial] = k.Value
	}
	return
}

func (x Party) ProductVarValues() (result ankat.ProductVarValues) {
	var productVarValues []struct {
		Sect   ankat.Sect          `db:"section"`
		Var    ankat.Var           `db:"var"`
		Point  ankat.Point         `db:"point"`
		Serial ankat.ProductSerial `db:"product_serial"`
		Value  float64             `db:"value"`
	}
	dbutils.MustSelect(x.db, &productVarValues, `
SELECT section, var, point, product_serial, value 
FROM product_value
WHERE party_id = ?;
`, x.PartyID)

	for _, k := range productVarValues {
		if len(result) == 0 {
			result = make(ankat.ProductVarValues)
		}
		if _, f := result[k.Sect]; !f {
			result[k.Sect] = make(map[ankat.Var]map[ankat.Point]map[ankat.ProductSerial]float64)
		}
		if _, f := result[k.Sect][k.Var]; !f {
			result[k.Sect][k.Var] = make(map[ankat.Point]map[ankat.ProductSerial]float64)
		}
		if _, f := result[k.Sect][k.Var][k.Point]; !f {
			result[k.Sect][k.Var][k.Point] = make(map[ankat.ProductSerial]float64)
		}
		if _, f := result[k.Sect][k.Var][k.Point]; !f {
			result[k.Sect][k.Var][k.Point] = make(map[ankat.ProductSerial]float64)
		}
		result[k.Sect][k.Var][k.Point][k.Serial] = k.Value
	}

	return
}


func (x Party) VerificationGasConcentration(gas ankat.GasCode) float64 {
	switch gas {
	case ankat.GasNitrogen:
		return x.CGas1
	case ankat.GasChan1Middle1:
		return x.CGas2
	case ankat.GasChan1Middle2:
		return x.CGas3
	case ankat.GasChan1End:
		return x.CGas4
	case ankat.GasChan2Middle:
		return x.CGas5
	case ankat.GasChan2End:
		return x.CGas6
	default:
		panic(fmt.Sprintf("unknown gas: %d", gas))
	}
}