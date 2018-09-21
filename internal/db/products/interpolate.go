package products

import (
	"fmt"
	"github.com/fpawel/ankat/internal/ankat"
	"github.com/fpawel/goutils/numeth"
	"github.com/pkg/errors"
)

func interpolate(xs [] numeth.Coordinate)([]float64, error) {
	if coefficients, ok := numeth.InterpolationCoefficients(xs); ok {
		return coefficients, nil
	}
	return nil, fmt.Errorf("не удалось выполнить интерполяцию: %v", xs)
}

func (x CurrentProduct) value(k ankat.ProductVar) (float64, error) {
	v,ok := x.Value(k)
	if ok {
		return v, nil
	}
	return 0, fmt.Errorf("нет значения в точке %s[%d]%s",k.Sect, k.Point, GetVar(x.db, k.Var).Name)
}



func (x CurrentProduct) interpolateLin(chanel ankat.AnkatChan) (coefficients []float64, xs []numeth.Coordinate, err error) {
	party := GetCurrentParty(x.db )
	points := chanel.LinPoints(party.IsCO2())
	xs = make([]numeth.Coordinate, len(points))
	for i, pt := range points {
		xs[i].Y = party.VerificationGasConcentration(pt.GasCode)
		xs[i].X, err  = x.value(pt.ProductVar)
		if err != nil {
			return
		}
	}
	coefficients, err  = interpolate(xs)
	return
}

func (x CurrentProduct) interpolateT0(chanel ankat.AnkatChan) (coefficients []float64, xs []numeth.Coordinate, err error) {
	ankat.MustValidChan(chanel)

	for i:= ankat.Point(0); i<3; i++ {
		var c numeth.Coordinate

		c.X,err = x.value(ankat.ProductVar{
			Sect: chanel.T0(),
			Var: chanel.Tpp(),
			Point:i,
		})
		if err != nil {
			return
		}

		c.Y,err  = x.value(ankat.ProductVar{
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

func (x CurrentProduct) interpolateTK(chanel ankat.AnkatChan) (coefficients []float64, xs []numeth.Coordinate, err error) {
	ankat.MustValidChan(chanel)

	for i:= ankat.Point(0); i<3; i++ {
		var tpp, var2, var0 float64

		tpp,err = x.value(ankat.ProductVar{
			Sect: chanel.TK(),
			Var: chanel.Tpp(),
			Point:i,
		})
		if err != nil {
			return
		}

		var2,err = x.value(ankat.ProductVar{
			Sect: chanel.TK(),
			Var: chanel.Var2(),
			Point:i,
		})
		if err != nil {
			return
		}

		var0,err = x.value( ankat.ProductVar{
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
	return
}

func (x CurrentProduct) interpolatePT() (coefficients []float64, xs []numeth.Coordinate, err error) {
	for i:= ankat.Point(0); i<3; i++ {
		var c numeth.Coordinate

		c.X,err = x.value( ankat.ProductVar{
			Sect: ankat.PT,
			Var: ankat.TppCh1,
			Point:i,
		})
		if err != nil {
			return
		}

		c.Y,err  = x.value(  ankat.ProductVar{
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

	return
}

func (x CurrentProduct) InterpolateSect(sect ankat.Sect) ([]float64, []numeth.Coordinate, error) {
	coefficients, xs, err := x.doInterpolateSect(sect)
	if err == nil {
		x.SetSectCoefficients( sect, coefficients)
	}
	return coefficients, xs, err
}

func (x CurrentProduct) doInterpolateSect(sect ankat.Sect) ([]float64, []numeth.Coordinate, error) {
	switch sect {
	case ankat.Lin1:
		return x.interpolateLin(ankat.Chan1)
	case ankat.Lin2:
		return x.interpolateLin(ankat.Chan2)
	case ankat.T01:
		return x.interpolateT0(ankat.Chan1)
	case ankat.TK1:
		return x.interpolateTK(ankat.Chan1)
	case ankat.T02:
		return x.interpolateT0(ankat.Chan2)
	case ankat.TK2:
		return x.interpolateTK(ankat.Chan2)
	case ankat.PT:
		return x.interpolatePT()
	default:
		panic("unknown interpolate sect: " + sect)
	}
}

