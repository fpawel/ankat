package ankat

import (
	"sort"
)

type ProductVarValues map[Sect]map[Var]map[Point]map[ProductSerial]float64

func (x ProductVarValues) Sects() (sects []Sect) {
	for sect := range x {
		sects = append(sects, sect)
	}
	sort.Slice(sects, func(i, j int) bool {
		return sects[i] < sects[j]
	})
	return
}

func (x ProductVarValues) Vars() (vars []Var) {
	m := map [Var]struct{}{}
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

func (x ProductVarValues) Points() (points []Point) {
	m := map [Point]struct{}{}
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

func (x ProductVarValues) Products() (products []ProductSerial) {
	m := map [ProductSerial]struct{}{}
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

func (x ProductVarValues) SectVarPointValues(sect Sect, v Var, p Point) (map[ProductSerial]float64,bool) {
	if s, ok := x[sect]; ok {
		if v,ok := s[v]; ok{
			p,ok := v[p]
			return p, ok
		}
	}
	return nil, false

}
