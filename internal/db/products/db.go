package products

import (
	"github.com/fpawel/ankat/internal/ankat"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	DB *sqlx.DB
}

func (x DB) PartyExists() bool {
	return PartyExists(x.DB)
}

func (x DB) Party(partyID ankat.PartyID) Party {
	return GetParty(x.DB,partyID)
}

func (x DB) CurrentParty() Party {
	return GetCurrentParty(x.DB)
}

func (x DB) CurrentProducts() []CurrentProduct{
	return CurrentProducts(x.DB)
}

func (x DB) CheckedProducts() []CurrentProduct{
	return CheckedProducts(x.DB)
}

func (x DB) CurrentProductAt(n int)  CurrentProduct{
	return GetCurrentProduct(x.DB,n)
}

func (x DB) CurrentProductOrderBySerial(productSerial ankat.ProductSerial ) int{
	return CurrentProductOrderBySerial(x.DB, productSerial)
}

func (x DB) Vars() []Var {
	return Vars(x.DB)
}

func (x DB) CheckedVars() []Var {
	return CheckedVars(x.DB)
}

func (x DB) Coefficients() []Coefficient {
	return Coefficients(x.DB)
}



func (x DB) Var(varID ankat.Var) Var {
	return GetVar(x.DB, varID)
}

func (x DB) Coefficient(coefficient ankat.Coefficient) Coefficient {
	return GetCoefficient(x.DB, coefficient)
}

func (x DB) CheckedCoefficients() []Coefficient {
	return CheckedCoefficients(x.DB)
}

func (x DB) CheckedOrAllCoefficients() []Coefficient {
	xs := x.CheckedCoefficients()
	if len(xs) == 0 {
		xs = x.Coefficients()
	}
	return xs
}



