package main

import (
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/guartutils/modbus"
	"fmt"
	"github.com/fpawel/procmq"
	"github.com/fpawel/guartutils/fetch"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/ankat/data/dataworks"
)

type productDevice struct {
	port     *comport.Port
	product  Product
	pipe     *procmq.ProcessMQ
	workCtrl uiworks.Runner
	data     data
}

type readProductResult struct {
	Product, Var int
	Value        float64
	Error        string
}

func notifyProductConnected(productOrdinal int, pipe  *procmq.ProcessMQ,  err error, format string, a ...interface{} ) {
	if fetch.Canceled(err) {
		return
	}
	var b struct {
		Product int
		Ok bool
		Text string
	}
	b.Product = productOrdinal
	if err == nil  {
		b.Ok = true
		b.Text = fmt.Sprintf(format, a...)
	} else {
		b.Text = err.Error()
	}
	pipe.Send("PRODUCT_CONNECTED",b)
}


func (x productDevice) notifyConnected(err error, format string, a ...interface{} ) {
	notifyProductConnected(x.product.Ordinal, x.pipe, err,  format, a...)
}

func (x productDevice) readCoefficient(coefficient int) (value float64, err error) {
	req := modbus.NewReadCoefficient(1,coefficient)
	var bytes []byte
	bytes, err = x.port.Fetch(req.Bytes())
	if fetch.Canceled(err) {
		return 0, err
	}
	if err == nil {
		value, err = req.ParseBCDValue(bytes)
		if err == nil {
			x.data.SetCoefficientValue(x.product.Serial, coefficient, value)
		}
	}
	x.notifyConnected(err, "K%d=%v", coefficient, value)

	for _,a := range x.data.Coefficients() {
		if a.Coefficient == coefficient {
			x.pipe.Send("READ_COEFFICIENT", readProductResult{
				Var:a.Ordinal,
				Product:x.product.Ordinal,
				Error:fmtErr(err),
				Value:value,
			})
			break
		}
	}
	if err == nil {
		x.writeInfo("K%d=%v", coefficient, value )
	} else {
		x.writeError("считывание K%d: %v", coefficient, err )
	}
	return value, err
}

func (x productDevice) readVar(v int) (value float64, err error) {
	req := modbus.NewReadBCD(1, uint16(v))

	var bytes []byte
	bytes, err = x.port.Fetch(req.Bytes())
	if fetch.Canceled(err) {
		return 0, err
	}
	if err == nil {
		value, err = req.ParseBCDValue(bytes)
	}
	x.notifyConnected(err,"$%d=%v", v, value)
	for _,a := range x.data.Vars() {
		if a.Var == v {
			x.pipe.Send("READ_VAR", readProductResult{
				Var:a.Ordinal,
				Product:x.product.Ordinal,
				Error:fmtErr(err),
				Value:value,
			})
			break
		}
	}

	return value, err
}

func (x productDevice) sendCmd(cmd uint16, value float64) error {

	_, err := x.port.Fetch(modbus.NewWriteCmdBCD(1, 0x16, cmd, value).Bytes())
	if fetch.NoAnswer(err) || modbus.ProtocolError(err) {
		x.workCtrl.WriteLog(x.product.Serial, dataworks.Warning, err.Error())
		_, err = x.port.Fetch(modbus.NewWriteCmdBCD(1, 0x10, cmd, value).Bytes())
	}
	if fetch.Canceled(err) {
		return nil
	}
	return err
}

func (x productDevice) sendCmdLog(cmd uint16, value float64) error {
	err := x.sendCmd(cmd, value)
	if err == nil {
		x.writeInfo("%s: %v", x.data.formatCmd(cmd), value )
	} else {
		x.writeError("%s: %v: %v", x.data.formatCmd(cmd), value, err )
	}
	return err
}


func (x productDevice) writeCoefficient(coefficient int) error {
	v := x.data.CoefficientValue(x.product.Serial, coefficient)
	if !v.Valid {
		x.workCtrl.WriteLog(x.product.Serial, dataworks.Warning, fmt.Sprintf(
			"запись К%d: значение коэффициента не задано", coefficient))
	}

	err := x.sendCmd( uint16( ( 0x80 << 8 ) + coefficient ), v.Float64 )

	if fetch.Canceled(err) {
		return nil
	}
	x.notifyConnected( err, "K%d:=%v", coefficient, v.Float64)
	if err == nil {
		x.writeInfo("K%d:=%v", coefficient, v.Float64 )
	} else {
		x.writeError("запись K%d:=%v: %v", coefficient, v.Float64, err )
	}
	for _,a := range x.data.Coefficients() {
		if a.Coefficient == coefficient {
			x.pipe.Send("READ_COEFFICIENT", readProductResult{
				Var:a.Ordinal,
				Product:x.product.Ordinal,
				Error:fmtErr(err),
				Value:v.Float64,
			})
			break
		}
	}
	return err
}

func (x productDevice) writeInfo(format string, a ...interface{}) {
	x.workCtrl.WriteLog(x.product.Serial, dataworks.Info, fmt.Sprintf(format, a...))
}

func (x productDevice) writeError(format string, a ...interface{} ) {
	x.workCtrl.WriteLog(x.product.Serial, dataworks.Error, fmt.Sprintf(format, a...))
}


