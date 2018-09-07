package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dataproducts"
	"github.com/fpawel/ankat/dataankat/dataworks"
	"math"
)

type productData struct {
	dataproducts.CurrentProduct
	app *app
}


func (x productData) interpolateSect(sect ankat.Sect)  {

	coefficients, values, err := x.InterpolateSect(sect)

	if err == nil {
		for i := range coefficients {
			coefficients[i] = math.Round(coefficients[i]*1000000.) / 1000000.
		}
		x.writeInfof("расчёт %v: %v: [%s] = [%v]", sect, coefficients, sect.CoefficientsStr(), values)
	} else {
		x.writeErrorf("расчёт %v не удался: %v", sect, err)
	}
}

func (x productData) writeLog(level dataworks.Level, str string) {
	x.app.uiWorks.WriteLog(x.ProductSerial, level, str)
}

func (x productData) writeLogf(level dataworks.Level, format string, a ...interface{}) {
	x.app.uiWorks.WriteLog(x.ProductSerial, level, fmt.Sprintf(format, a...))
}

func (x productData) writeInfo(str string) {
	x.writeLog(dataworks.Info, str)
}

func (x productData) writeInfof(format string, a ...interface{}) {
	x.writeLogf(dataworks.Info, format, a... )
}

func (x productData) writeError(str string) {
	x.writeLog( dataworks.Error, str)
}

func (x productData) writeErrorf(format string, a ...interface{}) {
	x.writeLogf(dataworks.Error, format, a... )
}
