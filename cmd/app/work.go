package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/data/dataproducts"
	"github.com/fpawel/ankat/data/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/guartutils/modbus"
	"github.com/fpawel/termochamber"
	"github.com/pkg/errors"
	"time"
)

func (x app) runWork(ordinal int, w uiworks.Work) {
	x.uiWorks.Perform(ordinal, w, func() {
		x.closeOpenedComports(x.sendMessage)
	})
}

func (x app) closeOpenedComports(logger logger) {
	for k, a := range x.comports {
		if a.comport.Opened() {
			if err := a.comport.Close(); err != nil {
				logger(0, dataworks.Error, err.Error())
			}
		}
		delete(x.comports, k)
	}
}

func (x app) comportProduct(p Product, errorLogger errorLogger) (*comport.Port, error) {
	a, existed := x.comports[p.Comport]
	if !existed || a.err != nil {
		portConfig := x.db.ComportSets("products")
		portConfig.Serial.Name = p.Comport
		a.comport = comport.NewPort(portConfig)
		a.err = a.comport.Open(x.uiWorks)
		x.comports[p.Comport] = a
		if a.err != nil {
			errorLogger(p.Serial, a.err.Error())
			notifyProductConnected(p.Ordinal, x.delphiApp, a.err, p.Comport)
		}
	}
	return a.comport, a.err
}

func (x app) comport(name string) (*comport.Port, error) {
	portConfig := x.db.ComportSets(name)
	a, existed := x.comports[portConfig.Serial.Name]
	if !existed || a.err != nil {
		a.comport = comport.NewPort(portConfig)
		a.err = a.comport.Open(x.uiWorks)
		x.comports[portConfig.Serial.Name] = a
	}
	return a.comport, a.err
}

func (x app) sendCmd(cmd uint16, value float64) error {
	x.uiWorks.WriteLogf(0, dataworks.Info, "Отправка команды %s: %v",
		x.db.formatCmd(cmd), value)
	return x.doEachProductDevice(x.uiWorks.WriteError, func(p productDevice) error {
		_ = p.sendCmdLog(cmd, value)
		return nil
	})
}

func (x app) runReadVarsWork() {

	x.runWork(0, uiworks.S("Опрос", func() error {
		dataworks.AddRootWork(x.db.dbProducts, "опрос")
		dataproducts.CreateNewSeries(x.db.dbProducts)
		defer dataproducts.DeleteLastEmptySeries(x.db.dbProducts)

		for {

			if len(x.db.CheckedProducts()) == 0 {
				return errors.New("не выбраны приборы")
			}

			for _, p := range x.db.CheckedProducts() {
				if x.uiWorks.Interrupted() {
					return nil
				}
				x.doProductDevice(p, x.sendErrorMessage, func(p productDevice) error {
					vars := x.db.CheckedVars()
					if len(vars) == 0 {
						vars = x.db.Vars()[:2]
					}
					for _, v := range vars {
						if x.uiWorks.Interrupted() {
							return nil
						}
						value, err := p.readVar(v.Var)
						if err == nil {
							dataproducts.AddChartValue(x.db.dbProducts, p.product.Serial, v.Var, value)
						}
					}
					return nil
				})
			}
		}
		return nil
	}))
}

func (x app) runReadCoefficientsWork() {

	x.runWork(0, uiworks.S("Считывание коэффициентов", func() error {
		return x.doEachProductDevice(x.sendErrorMessage, func(p productDevice) error {
			xs := x.db.CheckedCoefficients()
			if len(xs) == 0 {
				xs = x.db.Coefficients()
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
		return x.doEachProductDevice(x.sendErrorMessage, func(p productDevice) error {
			xs := x.db.CheckedCoefficients()
			if len(xs) == 0 {
				xs = x.db.Coefficients()
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

func (x app) doEachProductData(w func(p productData)) {

	for _, p := range x.db.CheckedProducts() {
		w(productData{
			product:  p,
			pipe:     x.delphiApp,
			workCtrl: x.uiWorks,
			db:       x.db,
		})
	}
}

func (x app) doEachProductDevice(errorLogger errorLogger, w func(p productDevice) error) error {
	if len(x.db.CheckedProducts()) == 0 {
		return errors.New("не выбраны приборы")
	}

	for _, p := range x.db.CheckedProducts() {
		if x.uiWorks.Interrupted() {
			return errors.New("прервано")
		}
		x.doProductDevice(p, errorLogger, w)
	}
	return nil
}

func (x app) doProductDevice(p Product, errorLogger errorLogger, w func(p productDevice) error) {
	x.delphiApp.Send("READ_PRODUCT", struct {
		Product int
	}{p.Ordinal})
	port, err := x.comportProduct(p, errorLogger)
	if err != nil {
		return
	}
	err = w(productDevice{
		productData{
			product:  p,
			pipe:     x.delphiApp,
			workCtrl: x.uiWorks,
			db:       x.db,
		},
		port,
	})
	if err != nil {
		errorLogger(p.Serial, err.Error())
	}
	x.delphiApp.Send("READ_PRODUCT", struct {
		Product int
	}{-1})
}

func (x app) doDelayWithReadProducts(what string, duration time.Duration) error {
	dataproducts.CreateNewSeries(x.db.dbProducts)
	vars := ankat.MainVars1()
	if x.db.IsTwoConcentrationChannels() {
		vars = append(vars, ankat.MainVars2()...)
	}
	iV, iP := 0, 0

	type ProductError struct {
		Serial ankat.ProductSerial
		Error  string
	}

	productErrors := map[ProductError]struct{}{}

	return x.uiWorks.Delay(what, duration, func() error {
		products := x.db.CheckedProducts()
		if len(products) == 0 {
			return errors.New(what + ": " + "не отмечено ни одного прибора")
		}
		if iP >= len(products) {
			iP, iV = 0, 0
		}
		x.doProductDevice(products[iP], func(productSerial ankat.ProductSerial, text string) {
			k := ProductError{productSerial, text}
			if _, exists := productErrors[k]; !exists {
				x.uiWorks.WriteError(productSerial, what+": "+text)
				productErrors[k] = struct{}{}
			}
		}, func(p productDevice) error {
			value, err := p.readVar(vars[iV])
			if err == nil {
				dataproducts.AddChartValue(x.db.dbProducts, p.product.Serial, vars[iV], value)
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

func (x app) doPause(what string, duration time.Duration) {
	_ = x.uiWorks.Delay(what, duration, func() error {
		return nil
	})
}

func (x app) blowGas(n ankat.GasCode) error {
	param := "delay_blow_nitrogen"
	what := fmt.Sprintf("продувка газа %d", n)
	if n == ankat.GasNitrogen {
		param = "delay_blow_gas"
		what = "продувка азота"
	}
	if err := x.switchGas(n); err != nil {
		return errors.Wrap(err, "не удалось переключить клапан")
	}
	duration := x.db.ConfigDuration(param) * time.Minute
	return x.doDelayWithReadProducts(what, duration)
}

func (x app) switchGas(n ankat.GasCode) error {
	port, err := x.comport("gas")
	if err != nil {
		return errors.Wrap(err, "не удалось открыть СОМ порт газового блока")
	}
	req := modbus.NewSwitchGasOven(byte(n))
	_, err = port.Fetch(req.Bytes())
	if err != nil {
		return x.promptErrorStopWork(errors.Wrapf(err, "нет связи c газовым блоком через %s", port.Config().Serial.Name))
	}
	return nil
}

func (x app) promptErrorStopWork(err error) error {
	s := x.delphiApp.SendAndGetAnswer("PROMPT_ERROR_STOP_WORK", err.Error())
	if s != "IGNORE" {
		return err
	}
	x.uiWorks.WriteLogf(0, dataworks.Warning, "ошибка автоматической настройки была проигнорирована: %v", err)
	return nil
}

func (x app) doSetupTemperature(temperature float64) error {
	port, err := x.comport("temp")
	if err != nil {
		return errors.Wrap(err, "не удалось открыть СОМ порт термокамеры")
	}
	deltaTemperature := x.db.ConfigValue("delta_temperature")

	cancelWaitTemperature, chWaitTemperature := termochamber.RunWaitForTemperature(termochamber.WaitTemperatureConfig{
		ReadTemperature: func() (float64, error) {
			return termochamber.T800Read(port)
		},
		Timeout:        x.db.ConfigDuration("timeout_temperature") * time.Minute,
		TemperatureMax: temperature - deltaTemperature,
		TemperatureMin: temperature + deltaTemperature,
	})
	chInterrupted := make(chan struct{})
	x.uiWorks.SubscribeInterrupted(chInterrupted, true)
	defer x.uiWorks.SubscribeInterrupted(chInterrupted, false)
	for {
		select {
		case <-chInterrupted:
			cancelWaitTemperature()
			return errors.New("перервано")
		case err := <-chWaitTemperature:
			return err
		}
	}
}

func (x app) setupAndHoldTemperature(temperature float64) error {
	if err := x.doSetupTemperature(temperature); err != nil {
		if err = x.promptErrorStopWork(
			errors.Wrapf(err,
				"не удалось установить температуру %v\"С в термокамере", temperature)); err != nil {
			return err
		}
	}
	duration := x.db.ConfigDuration("delay_temperature") * time.Hour
	x.uiWorks.WriteLogf(0, dataworks.Info,
		"выдержка термокамеры на %v\"C: в настройках задана длительность %v", temperature, duration)
	return x.doDelayWithReadProducts(fmt.Sprintf("выдержка термокамеры на %v\"C", temperature), duration)
}
