package dataproducts

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/numeth"
	"github.com/pkg/errors"
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
	return 0, fmt.Errorf("нет значения в точке %s[%d]%s",k.Sect, k.Point, x.VarName(k.Var))
}

func (x DBProducts) SetSectCoefficients(productSerial ankat.ProductSerial, sect ankat.Sect, values []float64) {
	for i := range values{
		x.SetCoefficientValue(productSerial, sect.Coefficient0() + ankat.Coefficient(i), values[i])
	}
}


func (x DBProducts) InterpolateLin(productSerial ankat.ProductSerial, chanel ankat.AnkatChan) (coefficients []float64, xs []numeth.Coordinate, err error) {
	points := chanel.LinPoints(x.IsCO2())
	xs = make([]numeth.Coordinate, len(points))
	for i, pt := range points {
		xs[i].Y = x.CurrentPartyVerificationGasConcentration(pt.GasCode)
		xs[i].X, err  = x.currentProductValue(productSerial, pt.ProductVar)
		if err != nil {
			return
		}
	}
	coefficients, err  = interpolate(xs)
	if err == nil {
		x.SetSectCoefficients(productSerial, chanel.Lin(), coefficients)
	}
	return
}

func (x DBProducts) InterpolateT0(productSerial ankat.ProductSerial, chanel ankat.AnkatChan) (coefficients []float64, xs []numeth.Coordinate, err error) {
	ankat.MustValidChan(chanel)

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
	if err == nil {
		x.SetSectCoefficients(productSerial, chanel.T0(), coefficients)
	}
	return
}

func (x DBProducts) InterpolateTK(productSerial ankat.ProductSerial, chanel ankat.AnkatChan) (coefficients []float64, xs []numeth.Coordinate, err error) {
	ankat.MustValidChan(chanel)

	for i:= ankat.Point(0); i<3; i++ {
		var tpp, var2, var0 float64

		tpp,err = x.currentProductValue(productSerial, ankat.ProductVar{
			Sect: chanel.TK(),
			Var: chanel.Tpp(),
			Point:i,
		})
		if err != nil {
			return
		}

		var2,err = x.currentProductValue(productSerial, ankat.ProductVar{
			Sect: chanel.TK(),
			Var: chanel.Var2(),
			Point:i,
		})
		if err != nil {
			return
		}

		var0,err = x.currentProductValue(productSerial, ankat.ProductVar{
			Sect: chanel.T0(),
			Var: chanel.Var2(),
			Point:i,
		})
		if err != nil {
			return
		}

		if var2 == var0 {
			err = errors.Errorf("не удалось выполнить расчёт термокомпенсации конца шкалы в точке №%d: " +
				"сигнал в конце шкалы равен сигналу в начале шкалы %v: ", i+1, var0, )
			return
		}

		xs = append(xs, numeth.Coordinate{
			X:tpp,
			Y:var2-var0,
		})
	}

	v1 := xs[1].Y
	for i:= ankat.Point(0); i<3; i++ {
		xs[i].Y = v1 / xs[i].Y
	}

	coefficients, err  = interpolate(xs)
	if err == nil {
		x.SetSectCoefficients(productSerial, chanel.TK(), coefficients)
	}
	return
}

func (x DBProducts) InterpolatePT(productSerial ankat.ProductSerial) (coefficients []float64, xs []numeth.Coordinate, err error) {
	for i:= ankat.Point(0); i<3; i++ {
		var c numeth.Coordinate

		c.X,err = x.currentProductValue(productSerial, ankat.ProductVar{
			Sect: ankat.PT,
			Var: ankat.TppCh1,
			Point:i,
		})
		if err != nil {
			return
		}

		c.Y,err  = x.currentProductValue(productSerial,  ankat.ProductVar{
			Sect: ankat.PT,
			Var: ankat.VdatP,
			Point:i,
		})
		if err != nil {
			return
		}
		xs = append(xs, c)
	}
	coefficients, err  = interpolate(xs)
	if err == nil {
		x.SetSectCoefficients(productSerial, ankat.PT, coefficients)
	}
	return
}

func (x DBProducts) InterpolateT01(productSerial ankat.ProductSerial) ([]float64, []numeth.Coordinate, error) {
	return x.InterpolateT0(productSerial, ankat.Chan1)
}

func (x DBProducts) InterpolateT02(productSerial ankat.ProductSerial) ([]float64, []numeth.Coordinate, error) {
	return x.InterpolateT0(productSerial, ankat.Chan2)
}

func (x DBProducts) InterpolateTK1(productSerial ankat.ProductSerial) ([]float64, []numeth.Coordinate, error) {
	return x.InterpolateTK(productSerial, ankat.Chan1)
}

func (x DBProducts) InterpolateTK2(productSerial ankat.ProductSerial) ([]float64, []numeth.Coordinate, error) {
	return x.InterpolateTK(productSerial, ankat.Chan2)
}

func (x DBProducts) InterpolateLin1(productSerial ankat.ProductSerial) ([]float64, []numeth.Coordinate, error) {
	return x.InterpolateLin(productSerial, ankat.Chan1)
}

func (x DBProducts) InterpolateLin2(productSerial ankat.ProductSerial) ([]float64, []numeth.Coordinate, error) {
	return x.InterpolateLin(productSerial, ankat.Chan2)
}