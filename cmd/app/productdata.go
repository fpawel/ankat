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

func (x productData) calculateLin1Coefficients() (coefficients []float64, xs []numeth.Coordinate, err error) {
	points := ankat.Lin1Points(x.db.IsCO2())
	xs = make([]numeth.Coordinate, len(points))
	var ok bool
	for i, pt := range points {
		xs[i].X = x.db.CurrentPartyValue(pt.GasCode.Var())
		xs[i].Y, ok = x.db.CurrentPartyProductValue(x.product.Serial, pt.ProductVar)
		if !ok {
			err = fmt.Errorf("нет значения в точке: %s",
				ankat.GasCodeDescription(pt.GasCode))
			return
		}
	}
	coefficients, ok = numeth.InterpolationCoefficients(xs)
	if !ok {
		err = fmt.Errorf("не удалось выполнить интерполяцию: %v", xs)
	}
	return
}

func (x productData) calculateLin2Coefficients() (coefficients []float64, xs []numeth.Coordinate, err error) {
	points := ankat.Lin2Points()
	xs = make([]numeth.Coordinate, len(points))
	var ok bool
	for i, pt := range points {
		xs[i].X = x.db.CurrentPartyValue(pt.GasCode.Var())
		xs[i].Y, ok = x.db.CurrentPartyProductValue(x.product.Serial, pt.ProductVar)
		if !ok {
			err = fmt.Errorf("нет значения в точке %s", ankat.GasCodeDescription(pt.GasCode))
			return
		}
	}
	coefficients, ok = numeth.InterpolationCoefficients(xs)
	if !ok {
		err = fmt.Errorf("не удалось выполнить интерполяцию: %v", xs)
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
