package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/data/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/numeth"
	"github.com/fpawel/procmq"
)

type productData struct {
	product  Product
	pipe     *procmq.ProcessMQ
	workCtrl uiworks.Runner
	db       db
}

func (x productData) calculateT0(chanel ankat.ConcentrationChannel) (coefficients []float64, xs []numeth.Coordinate, err error) {
	ankat.MustValidConcentrationChannel(chanel)

	ok := false
	for i:= ankat.Point(0); i<2; i++ {
		var c numeth.Coordinate
		pt := ankat.ProductVar{
			Sect: chanel.T0(),
			Var: chanel.Tpp(),
			Point:i,
		}
		c.X,ok = x.db.CurrentPartyProductValue(x.product.Serial, pt)
		if !ok {
			err = fmt.Errorf("нет значения в точке %v", pt)
			return
		}

		pt = ankat.ProductVar{
			Sect: chanel.T0(),
			Var: chanel.Var2(),
			Point:i,
		}
		c.Y,ok = x.db.CurrentPartyProductValue(x.product.Serial, pt)
		if !ok {
			err = fmt.Errorf("нет значения в точке %v", pt)
			return
		}
		c.Y *= -1
		xs = append(xs, c)
	}
	coefficients, ok = numeth.InterpolationCoefficients(xs)
	if !ok {
		err = fmt.Errorf("не удалось выполнить интерполяцию: %v", xs)
	}
	return
}

func (x productData) calculateLin(chanel ankat.ConcentrationChannel) (coefficients []float64, xs []numeth.Coordinate, err error) {
	points := chanel.LinPoints(x.db.IsCO2())
	xs = make([]numeth.Coordinate, len(points))
	var ok bool
	for i, pt := range points {
		xs[i].X = x.db.CurrentPartyValue(pt.GasCode.Var())
		xs[i].Y, ok = x.db.CurrentPartyProductValue(x.product.Serial, pt.ProductVar)
		if !ok {
			err = fmt.Errorf("нет значения в точке: %s",
				pt.GasCode.Description())
			return
		}
	}
	coefficients, ok = numeth.InterpolationCoefficients(xs)
	if !ok {
		err = fmt.Errorf("не удалось выполнить интерполяцию: %v", xs)
		return
	}
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
