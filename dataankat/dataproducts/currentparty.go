package dataproducts

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dbutils"
)

type CurrentParty struct {
	Party
}


func (x CurrentParty) AnkatChannels() (xs []ankat.AnkatChan) {
	xs = append(xs, ankat.Chan1)
	if x.IsTwoConcentrationChannels() {
		xs = append(xs, ankat.Chan2)
	}
	return
}




func (x CurrentParty) CurrentProducts() (products []CurrentProduct) {
	dbutils.MustSelect(x.db, &products, `SELECT * FROM current_party_products_config;`)
	for i := range products{
		products[i].db = x.db
	}
	return
}

func (x CurrentParty) CheckedProducts() (products []CurrentProduct) {
	dbutils.MustSelect(x.db, &products,
		`SELECT * FROM current_party_products_config WHERE checked=1;`)
	for i := range products{
		products[i].db = x.db
	}
	return
}

func (x CurrentParty) CurrentProduct(n int) (p CurrentProduct) {
	dbutils.MustGet(x.db, &p, `SELECT * FROM current_party_products_config WHERE ordinal = $1;`, n)
	return
}

func (x CurrentParty) VerificationGasConcentration(gas ankat.GasCode) float64 {
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

func (x CurrentParty) SetMainErrorConcentrationValue(ankatChan ankat.AnkatChan, scalePos ankat.ScalePosition,
	temperaturePos ankat.TemperaturePosition, serial ankat.ProductSerial, value float64) {

}
