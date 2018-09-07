package ankat

import (
	"sort"
)

type PartyCoefficients map[Coefficient]map[ProductSerial]float64

func (x PartyCoefficients) Coefficients() (coefficients []Coefficient) {
	for coefficient := range x {
		coefficients = append(coefficients, coefficient)
	}
	sort.Slice(coefficients, func(i, j int) bool {
		return coefficients[i] < coefficients[j]
	})
	return
}

func (x PartyCoefficients) Products() (products []ProductSerial) {
	xs := map[ProductSerial]struct{}{}
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
