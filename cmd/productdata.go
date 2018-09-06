package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat"
	"github.com/fpawel/ankat/dataankat/dataproducts"
	"github.com/fpawel/ankat/dataankat/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/procmq"
	"math"
)

type productData struct {
	product  dataproducts.Product
	pipe     *procmq.ProcessMQ
	workCtrl uiworks.Runner
	db       dataankat.DBAnkat
}

func (x productData) productDB() dataproducts.DBCurrentProduct {
	return dataproducts.DBCurrentProduct{
		DB: x.db.DBProducts.DB,
		ProductSerial: x.product.Serial,
	}
}

func (x productData) interpolateSect(sect ankat.Sect)  {
	coefficients, values, err := x.productDB().InterpolateSect(sect)

	if err == nil {
		for i := range coefficients {
			coefficients[i] = math.Round(coefficients[i]*1000000.) / 1000000.
		}
		x.writeInfof("расчёт %v: %v: [%s] = [%v]", sect, coefficients, sect.CoefficientsStr(), values)
	} else {
		x.writeErrorf("расчёт %v не удался: %v", sect, err)
	}
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
