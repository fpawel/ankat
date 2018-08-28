package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat"
	"github.com/fpawel/ankat/dataankat/dataproducts"
	"github.com/fpawel/ankat/dataankat/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/numeth"
	"github.com/fpawel/procmq"
	"github.com/pkg/errors"
)

type productData struct {
	product  dataproducts.Product
	pipe     *procmq.ProcessMQ
	workCtrl uiworks.Runner
	db       dataankat.DBAnkat
}

func (x productData) value(k ankat.ProductVar) (float64, error) {
	v,ok := x.db.CurrentPartyProductValue(x.product.Serial, k)
	if ok {
		return v, nil
	}
	return 0, fmt.Errorf("нет значения в точке %v", k)
}

func interpolate(xs [] numeth.Coordinate)([]float64, error) {
	if coefficients, ok := numeth.InterpolationCoefficients(xs); ok {
		return coefficients, nil
	}
	return nil, fmt.Errorf("не удалось выполнить интерполяцию: %v", xs)
}

func (x productData) calculateT0(chanel ankat.ConcentrationChannel) (coefficients []float64, xs []numeth.Coordinate, err error) {
	return x.db.CalculateT0(x.product.Serial, chanel)
}

func (x productData) calculateTK(chanel ankat.ConcentrationChannel) (coefficients []float64, xs []numeth.Coordinate, err error) {
	ankat.MustValidConcentrationChannel(chanel)

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

		var0,err = x.value(ankat.ProductVar{
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

func (x productData) calculateLin(chanel ankat.ConcentrationChannel) (coefficients []float64, xs []numeth.Coordinate, err error) {
	points := chanel.LinPoints(x.db.IsCO2())
	xs = make([]numeth.Coordinate, len(points))
	for i, pt := range points {
		xs[i].Y = x.db.CurrentPartyValue(pt.GasCode.Var())
		xs[i].X, err  = x.value(pt.ProductVar)
		if err != nil {
			return
		}
	}
	coefficients, err  = interpolate(xs)
	return
}

func (x productData) writeInfo(str string) {
	x.workCtrl.WriteLog(x.product.Serial, dataworks.Info, str)
}

func (x productData) writeInfof(format string, a ...interface{}) {
	x.writeInfo(fmt.Sprintf(format, a...))
}

func (x productData) writeError(str string) {
	x.workCtrl.WriteLog(x.product.Serial, dataworks.Error, str)
}

func (x productData) writeErrorf(format string, a ...interface{}) {
	x.writeError(fmt.Sprintf(format, a...))
}
