package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/data/dataproducts"
	"github.com/fpawel/ankat/data/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/guartutils/modbus"
	"github.com/pkg/errors"
	"time"
)

func (x app) runWork(ordinal int, w uiworks.Work) {
	x.uiWorks.Perform(ordinal, w, func() {
		x.closeOpenedComports(x.sendMessage)
	})
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
		a.err = a.comport.Open(x.uiWorks)
		x.comports[p.Comport] = a
		if a.err != nil {
			logger(p.Serial, dataworks.Error, a.err.Error())
			notifyProductConnected(p.Ordinal, x.delphiApp, a.err, p.Comport)
		}
	}
	return a.comport, a.err
}

func (x *app) comport(name string) (*comport.Port, error) {
	portConfig := x.data.ComportSets(name)
	a, existed := x.comports[portConfig.Serial.Name]
	if !existed || a.err != nil {
		a.comport = comport.NewPort(portConfig)
		a.err = a.comport.Open(x.uiWorks)
		x.comports[portConfig.Serial.Name] = a
		if a.err != nil {
			x.uiWorks.WriteLog(0, dataworks.Error, a.err.Error())
		}
	}
	return a.comport, a.err
}

func (x app) sendCmd(cmd uint16, value float64) error {
	x.uiWorks.WriteLogf(0, dataworks.Info, "Отправка команды %s: %v",
		x.data.formatCmd(cmd), value)
	return x.doEachProductDevice(x.uiWorks.WriteLog, func(p productDevice) error {
		_ = p.sendCmdLog(cmd, value)
		return nil
	})
}

func (x app) runMainWork() {

}

func (x app) runReadVarsWork() {

	x.runWork(0, uiworks.S("Опрос", func() error {
		dataproducts.CreateNewSeries(x.data.dbProducts, "Опрос")
		defer dataproducts.DeleteLastEmptySeries(x.data.dbProducts)
		for !x.uiWorks.Interrupted() {
			if err := x.doEachProductDevice(x.sendMessage, func(p productDevice) error {
				if len(x.data.CheckedVars()) == 0 {
					return errors.New("не выбраны регистры опроса")
				}
				for _, v := range x.data.CheckedVars() {
					if x.uiWorks.Interrupted() {
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

	x.runWork(0, uiworks.S("Считывание коэффициентов", func() error {
		return x.doEachProductDevice(x.sendMessage, func(p productDevice) error {
			xs := x.data.CheckedCoefficients()
			if len(xs) == 0 {
				return errors.New("не выбраны коэффициенты")
			}
			for _, v := range xs {
				if x.uiWorks.Interrupted() {
					return nil
				}
				_, _ = p.readCoefficient(v.Coefficient)
			}
			return nil
		})
	}))
}

func (x app) runWriteCoefficientsWork() {

	x.runWork(0, uiworks.S("Запись коэффициентов", func() error {
		return x.doEachProductDevice(x.sendMessage, func(p productDevice) error {
			xs := x.data.CheckedCoefficients()
			if len(xs) == 0 {
				return errors.New("не выбраны коэффициенты")
			}
			for _, v := range xs {
				if x.uiWorks.Interrupted() {
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
		if x.uiWorks.Interrupted() {
			return nil
		}
		if err := x.doProductDevice(p, logger, w); err != nil {
			return err
		}
	}
	return nil
}

func (x *app) doProductDevice(p Product, logger logger, w func(p productDevice) error) error {
	x.delphiApp.Send("READ_PRODUCT", struct {
		Product int
	}{p.Ordinal})
	port, err := x.comportProduct(p, logger)
	if err != nil {
		return err
	}
	err = w(productDevice{
		product:  p,
		pipe:     x.delphiApp,
		workCtrl: x.uiWorks,
		port:     port,
		data:     x.data,
	})
	x.delphiApp.Send("READ_PRODUCT", struct {
		Product int
	}{-1})
	return err
}

func (x *app) doDelay(what string, duration time.Duration) error {
	dataproducts.CreateNewSeries(x.data.dbProducts, what)
	vars := ankat.MainVars1()
	if x.data.IsTwoConcentrationChannels() {
		vars = append(vars, ankat.MainVars2()...)
	}
	iV, iP := 0, 0
	return x.uiWorks.Delay(what, duration, func() error {
		products := x.data.CheckedProducts()
		if len(products) == 0 {
			return errors.New("не отмечено ни одного прибора")
		}
		if iP >= len(products) {
			iP, iV = 0, 0
		}
		x.doProductDevice(products[iP], x.sendMessage, func(p productDevice) error {
			value, err := p.readVar(vars[iV])
			if err == nil {
				dataproducts.AddChartValue(x.data.dbProducts, p.product.Serial, vars[iV], value)
			}
			return nil
		})
		if iV < len(vars)-1 {
			iV++
			return nil
		}
		iV = 0
		if iP < len(products)-1 {
			iP++
		} else {
			iP = 0
		}
		return nil
	})
}

func (x app) blowGas(n ankat.GasCode) error {
	param := "delay_blow_nitrogen"
	what := fmt.Sprintf("продувка газа %d", n)
	if n == ankat.Nitrogen {
		param = "delay_blow_gas"
		what = "продувка азота"
	}
	if err := x.switchGas(n); err != nil {
		return errors.Wrap(err, "не удалось переключить клапан")
	}
	duration := x.data.ConfigDuration(param) * time.Minute
	x.uiWorks.WriteLogf(0, dataworks.Info,
		"%s: в настройках задана длительность %v", what, duration)
	return x.doDelay(what, duration)
}

func (x *app) switchGas(n ankat.GasCode) error {
	port, err := x.comport("gas")
	if err != nil {
		return errors.Wrap(err, "не удалось открыть СОМ порт")
	}
	req := modbus.NewSwitchGasOven(byte(n))
	_, err = port.Fetch(req.Bytes())
	var what string
	if n == ankat.CloseGasBlock {
		what = "закрыть все клапаны"
	} else {
		what = fmt.Sprintf("открыть клапан %d", n)
	}
	if err == nil {
		x.uiWorks.WriteLog(0, dataworks.Info, what)
	} else {
		x.uiWorks.WriteLogf(0, dataworks.Error, "%s: %v", what, err)
	}
	return errors.Wrap(err, "нет связи")
}
