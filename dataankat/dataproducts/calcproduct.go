package dataproducts

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/numeth"
)

func interpolate(xs [] numeth.Coordinate)([]float64, error) {
	if coefficients, ok := numeth.InterpolationCoefficients(xs); ok {
		return coefficients, nil
	}
	return nil, fmt.Errorf("не удалось выполнить интерполяцию: %v", xs)
}

func (x DBProducts) currentProductValue(p ankat.ProductSerial, k ankat.ProductVar) (float64, error) {
	v,ok := x.CurrentPartyProductValue(p, k)
	if ok {
		return v, nil
	}
	return 0, fmt.Errorf("нет значения в точке %v", k)
}

func (x DBProducts) CalculateT0(productSerial ankat.ProductSerial, chanel ankat.ConcentrationChannel) (coefficients []float64, xs []numeth.Coordinate, err error) {
	ankat.MustValidConcentrationChannel(chanel)

	for i:= ankat.Point(0); i<3; i++ {
		var c numeth.Coordinate

		c.X,err = x.currentProductValue(productSerial, ankat.ProductVar{
			Sect: chanel.T0(),
			Var: chanel.Tpp(),
			Point:i,
		})
		if err != nil {
			return
		}

		c.Y,err  = x.currentProductValue(productSerial,  ankat.ProductVar{
			Sect: chanel.T0(),
			Var: chanel.Var2(),
			Point:i,
		})
		if err != nil {
			return
		}
		c.Y *= -1
		xs = append(xs, c)
	}
	coefficients, err  = interpolate(xs)
	return
}
