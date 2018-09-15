package main

import (
	"fmt"
	"github.com/fpawel/ankat/internal/ankat"
	"github.com/fpawel/ankat/internal/db/products"
	"github.com/fpawel/ankat/internal/db/worklog"
	"math"
)

type productData struct {
	products.CurrentProduct
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

func (x productData) writeLog(level worklog.Level, str string) {
	x.app.uiWorks.WriteLog(x.ProductSerial, level, str)
}

func (x productData) writeLogf(level worklog.Level, format string, a ...interface{}) {
	x.app.uiWorks.WriteLog(x.ProductSerial, level, fmt.Sprintf(format, a...))
}

func (x productData) writeInfo(str string) {
	x.writeLog(worklog.Info, str)
}

func (x productData) writeInfof(format string, a ...interface{}) {
	x.writeLogf(worklog.Info, format, a... )
}

func (x productData) writeError(str string) {
	x.writeLog( worklog.Error, str)
}

func (x productData) writeErrorf(format string, a ...interface{}) {
	x.writeLogf(worklog.Error, format, a... )
}
