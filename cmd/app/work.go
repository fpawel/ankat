package main

import (
	"github.com/fpawel/ankat/data/dataproducts"
	"github.com/fpawel/ankat/data/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/guartutils/comport"
	"github.com/pkg/errors"
)

func (x app) runWork(w uiworks.Work) {
	f := w.Action
	w.Action = func() error {
		defer x.closeOpenedComports(x.sendMessage)
		return f()
	}
	x.workCtrl.Perform(w)
}

func (x *app) closeOpenedComports(logger logger) {
	for k, a := range x.comports {
		if a.comport.Opened() {
			if err := a.comport.Close(); err != nil {
				logger(0, dataworks.Error, err.Error())
			}
		}
		delete(x.comports, k)
	}
}
func (x *app) comportProduct(p Product, logger logger) (*comport.Port, error) {
	a, existed := x.comports[p.Comport]
	if !existed || a.err != nil {
		portConfig := x.data.ComportSets("products")
		portConfig.Serial.Name = p.Comport
		a.comport = comport.NewPort(portConfig)
		a.err = a.comport.Open(x.workCtrl)
		x.comports[p.Comport] = a
		if a.err != nil {
			logger(p.Serial, dataworks.Error, a.err.Error())
			notifyProductConnected(p.Ordinal, x.delphiApp, a.err, p.Comport)
		}
	}
	return a.comport, a.err
}

func (x app) sendCmd(cmd uint16, value float64) error {
	x.workCtrl.WriteLogf(0, dataworks.Info, "Отправка команды %s: %v",
		x.data.formatCmd(cmd), value)
	return x.doEachProductDevice(x.workCtrl.WriteLog, func(p productDevice) error {
		_ = p.sendCmdLog(cmd, value)
		return nil
	})
}

func (x app) runReadVarsWork() {

	x.runWork(uiworks.S("Опрос", func() error {
		dataproducts.CreateNewSeries(x.data.dbProducts, "Опрос")
		for !x.workCtrl.Interrupted() {
			if err := x.doEachProductDevice(x.sendMessage, func(p productDevice) error {
				if len(x.data.CheckedVars()) == 0 {
					return errors.New("не выбраны регистры опроса")
				}
				for _, v := range x.data.CheckedVars() {
					if x.workCtrl.Interrupted() {
						return nil
					}
					value, err := p.readVar(v.Var)
					if err == nil {
						dataproducts.AddChartValue(x.data.dbProducts, p.product.Serial, v.Var, value)
					}
				}
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}))
}

func (x app) runReadCoefficientsWork() {

	x.runWork(uiworks.S("Считывание коэффициентов", func() error {
		return x.doEachProductDevice(x.sendMessage, func(p productDevice) error {
			xs := x.data.CheckedCoefficients()
			if len(xs) == 0 {
				return errors.New("не выбраны коэффициенты")
			}
			for _, v := range xs {
				if x.workCtrl.Interrupted() {
					return nil
				}
				_, _ = p.readCoefficient(v.Coefficient)
			}
			return nil
		})
	}))
}

func (x app) runWriteCoefficientsWork() {

	x.runWork(uiworks.S("Запись коэффициентов", func() error {
		return x.doEachProductDevice(x.sendMessage, func(p productDevice) error {
			xs := x.data.CheckedCoefficients()
			if len(xs) == 0 {
				return errors.New("не выбраны коэффициенты")
			}
			for _, v := range xs {
				if x.workCtrl.Interrupted() {
					return nil
				}
				_ = p.writeCoefficient(v.Coefficient)
			}
			return nil
		})
	}))
}

func (x app) doEachProductDevice(logger logger, w func(p productDevice) error) error {
	if len(x.data.CheckedProducts()) == 0 {
		return errors.New("не выбраны приборы")
	}

	for _, p := range x.data.CheckedProducts() {
		if x.workCtrl.Interrupted() {
			return nil
		}
		x.delphiApp.Send("READ_PRODUCT", struct {
			Product int
		}{p.Ordinal})
		if port, err := x.comportProduct(p, logger); err == nil {

			if err = w(productDevice{
				product:  p,
				pipe:     x.delphiApp,
				workCtrl: x.workCtrl,
				port:     port,
				data:     x.data,
			}); err != nil {
				return err
			}
		}
		x.delphiApp.Send("READ_PRODUCT", struct {
			Product int
		}{-1})
	}
	return nil
}
