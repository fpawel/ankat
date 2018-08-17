package dataproducts

import (
	"github.com/fpawel/ankat"
	"github.com/jmoiron/sqlx"
	"sort"
	"time"
)

type PartyInfo struct {
	PartyID          ankat.PartyID         `db:"party_id"`
	CreatedAt        time.Time             `db:"created_at"`
	Vars             map[ankat.Var]string  `db:"-"`
	Products         []ankat.ProductSerial `db:"-"`
	Values           []KeyStr              `db:"-"`
	Coefficients     Coefficients          `db:"-"`
	ProductVarValues ProductVarValues      `db:"-"`
}

type Coefficients map[ankat.Coefficient]map[ankat.ProductSerial]float64

type KeyStr struct {
	Key string `db:"name"`
	Str string `db:"value"`
}

type ProductVarValues map[ankat.Sect]map[ankat.Var]map[ankat.Point]map[ankat.ProductSerial]float64

func GetPartyInfo(x *sqlx.DB, partyID ankat.PartyID) (party PartyInfo) {
	dbMustGet(x, &party, `
SELECT party_id, created_at FROM party WHERE party_id = $1;`, partyID)

	dbMustSelect(x, &party.Products, `
SELECT product_serial
FROM product
WHERE party_id = $1
ORDER BY product_serial ASC;`, partyID)

	dbMustSelect(x, &party.Values, `SELECT name, value FROM party_value2 WHERE party_id = ?;`, partyID)

	var coefficients []struct {
		Coefficient   ankat.Coefficient   `db:"coefficient_id"`
		ProductSerial ankat.ProductSerial `db:"product_serial"`
		Value         float64             `db:"value"`
	}
	dbMustSelect(x, &coefficients, `
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
	dbMustSelect(x, &productVarValues, `
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

func (x ProductVarValues) Sects() (sects []ankat.Sect) {
	for sect := range x {
		sects = append(sects, sect)
	}
	sort.Slice(sects, func(i, j int) bool {
		return sects[i] < sects[j]
	})
	return
}

func (x ProductVarValues) Vars() (vars []ankat.Var) {
	m := map [ankat.Var]struct{}{}
	for sect := range x {
		for v := range x[sect] {
			m[v] = struct{}{}
		}
	}
	for v := range m {
		vars = append(vars, v)
	}

	sort.Slice(vars, func(i, j int) bool {
		return vars[i] < vars[j]
	})
	return
}

func (x ProductVarValues) Points() (points []ankat.Point) {
	m := map [ankat.Point]struct{}{}
	for sect := range x {
		for v := range x[sect] {
			for p := range x[sect][v] {
				m[p] = struct{}{}
			}
		}
	}

	for v := range m {
		points = append(points, v)
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i] < points[j]
	})
	return
}

func (x ProductVarValues) Products() (products []ankat.ProductSerial) {
	m := map [ankat.ProductSerial]struct{}{}
	for sect := range x {
		for v := range x[sect] {
			for p := range x[sect][v] {
				for p := range x[sect][v][p] {
					m[p] = struct{}{}
				}
			}
		}
	}

	for v := range m {
		products = append(products, v)
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i] < products[j]
	})
	return
}

func (x ProductVarValues) SectVarPointValues(sect ankat.Sect, v ankat.Var, p ankat.Point) (map[ankat.ProductSerial]float64,bool) {
	if s, ok := x[sect]; ok {
		if v,ok := s[v]; ok{
			p,ok := v[p]
			return p, ok
		}
	}
	return nil, false

}

func (x Coefficients) Coefficients() (coefficients []ankat.Coefficient) {
	for coefficient := range x {
		coefficients = append(coefficients, coefficient)
	}
	sort.Slice(coefficients, func(i, j int) bool {
		return coefficients[i] < coefficients[j]
	})
	return
}

func (x Coefficients) Products() (products []ankat.ProductSerial) {
	xs := map[ankat.ProductSerial]struct{}{}
	for _, ps := range x {
		for p := range ps {
			xs[p] = struct{}{}
		}
	}

	for p := range xs {
		products = append(products, p)
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i] < products[j]
	})

	return
}
