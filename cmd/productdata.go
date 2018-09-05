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
)

type productData struct {
	product  dataproducts.Product
	pipe     *procmq.ProcessMQ
	workCtrl uiworks.Runner
	db       dataankat.DBAnkat
}

func (x productData) value(k ankat.ProductVar) (float64, error) {
	v,ok := x.db.CurrentParty().ProductValue(x.product.Serial, k)
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

func (x productData) calculateT0(chanel ankat.AnkatChan) (coefficients []float64, xs []numeth.Coordinate, err error) {
	return x.db.InterpolateT0(x.product.Serial, chanel)
}

func (x productData) calculateTK(chanel ankat.AnkatChan) (coefficients []float64, xs []numeth.Coordinate, err error) {
	return x.db.InterpolateTK(x.product.Serial, chanel)
}

func (x productData) calculateLin(chanel ankat.AnkatChan) (coefficients []float64, xs []numeth.Coordinate, err error) {
	return x.db.InterpolateLin(x.product.Serial, chanel)
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
