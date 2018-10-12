package main

import (
	"fmt"
	"github.com/fpawel/ankat/internal/ankat"
	"github.com/fpawel/ankat/internal/db/products"
	"github.com/fpawel/ankat/internal/db/worklog"
	"github.com/fpawel/ankat/internal/ui/uiworks"
	"github.com/fpawel/goutils/serial/comport"
	"github.com/fpawel/goutils/serial/modbus"
	"github.com/fpawel/goutils/serial/termochamber"
	"github.com/pkg/errors"
	"time"
)

func (x app) runWork(ordinal int, w uiworks.Work) {
	x.uiWorks.Perform(ordinal, w, func() {
		if err := x.comports.Close(); err != nil {
			x.sendMessage(0, worklog.Error, err.Error())
		}
	})
	x.delphiApp.Send("READ_PRODUCT", struct {
		Product int
	}{-1})
}


func (x app) comportProduct(p products.CurrentProduct, errorLogger errorLogger) (*comport.Port, error) {
	portConfig := x.ConfigSect("comport_products").Comport()
	portConfig.Serial.Name = p.Comport
	port,err := x.comports.Open(portConfig)
	if err != nil {
		x.sendProductConnectionError(p.Ordinal, err.Error())
	}
	return port,err
}

func (x app) comport(name string) (*comport.Port, error) {
	portConfig := x.ConfigSect(name).Comport()
	return x.comports.Open( portConfig )
}

func (x app) sendProductConnectionError(productOrdinal int, text string)  {
	var b struct {
		Product int
		Ok      bool
		Text    string
	}
	b.Product = productOrdinal
	b.Text = text
	x.delphiApp.Send("PRODUCT_CONNECTED", b)
}

func (x app) sendCmd(cmd ankat.Cmd, value float64) error {
	x.uiWorks.WriteLogf(0, worklog.Info, "Отправка команды %s: %v",
		ankat.FormatCmd(cmd), value)
	return x.doEachProductDevice(x.uiWorks.WriteError, func(p productDevice) error {
		_ = p.sendCmdLog(cmd, value)
		return nil
	})
}

func (x app) runReadVarsWork() {

	x.runWork(0, uiworks.S("Опрос", func() error {

		series := products.NewSeries()
		defer series.Save(x.dbProducts, "Опрос")

		for {

			if len(x.DBProducts.CheckedProducts()) == 0 {
				return errors.New("не выбраны приборы")
			}

			for _, p := range x.DBProducts.CheckedProducts() {
				if x.uiWorks.Interrupted() {
					return nil
				}
				x.doProductDevice(p, x.sendErrorMessage, func(p productDevice) error {
					vars := x.DBProducts.CheckedVars()
					if len(vars) == 0 {
						vars = x.DBProducts.Vars()[:2]
					}
					for _, v := range vars {
						if x.uiWorks.Interrupted() {
							return nil
						}
						value, err := p.readVar(v.Var)
						if err == nil {
							series.AddRecord(p.ProductSerial, v.Var, value)
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
		var read []products.ProductCoefficientValue
		x.doEachProductDevice(x.sendErrorMessage, func(p productDevice) error {
			for _, v := range x.DBProducts.CheckedOrAllCoefficients() {
				if x.uiWorks.Interrupted() {
					return nil
				}
				value, err := p.readCoefficient(v.Coefficient)
				if err == nil {
					read = append(read, products.ProductCoefficientValue{
						ProductSerial:p.ProductSerial,
						Coefficient:v.Coefficient,
						Value:value,
					})
				}
			}
			return nil
		})
		x.DBProducts.CurrentParty().SetCoefficientsValues(read)
		
		return nil
	}))
}

func (x *app) runSetCoefficient(productOrder, coefficientOrder int) {
	coefficient := x.DBProducts.Coefficients()[coefficientOrder].Coefficient
	p := x.DBProducts.CurrentProductAt(productOrder)
	s := fmt.Sprintf("Прибор %d: K%d: ", p.ProductSerial, coefficient)
	value,okValue := p.CoefficientValue(coefficient)
	if okValue {
		s += fmt.Sprintf("запись %v", value)
	} else {
		s += "считывание значения"
	}

	x.runWork(0, uiworks.S(s, func() ( err error) {

		var pd productDevice
		pd.app = x
		pd.CurrentProduct = p
		pd.port, err = x.comportProduct(p, x.sendErrorMessage)
		if err != nil {
			pd.notifyCoefficient(coefficient, 0, err)
			return
		}

		if okValue {
			if err := pd.writeCoefficient(coefficient); err != nil {
				return err
			}
		}
		_,err = pd.readAndSaveCoefficient(coefficient)
		return
	}))
}

func (x app) runWriteCoefficientsWork() {

	x.runWork(0, uiworks.S("Запись коэффициентов", func() error {
		return x.doEachProductDevice(x.sendErrorMessage, func(p productDevice) error {
			for _, v := range x.DBProducts.CheckedOrAllCoefficients() {
				if x.uiWorks.Interrupted() {
					return nil
				}
				_ = p.writeCoefficient(v.Coefficient)
			}
			return nil
		})
	}))
}

func (x *app) doEachProductData(w func(p productData)) {
	for _, p := range x.DBProducts.CheckedProducts() {
		w(productData{app: x, CurrentProduct: p,})
	}
}

func (x app) doEachProductDevice(errorLogger errorLogger, w func(p productDevice) error) error {
	if len(x.DBProducts.CheckedProducts()) == 0 {
		return errors.New("не выбраны приборы")
	}

	for _, p := range x.DBProducts.CheckedProducts() {
		if x.uiWorks.Interrupted() {
			return errors.New("прервано")
		}
		x.doProductDevice(p, errorLogger, w)
	}
	return nil
}

func (x *app) doProductDevice(p products.CurrentProduct, errorLogger errorLogger, w func(p productDevice) error) {
	x.delphiApp.Send("READ_PRODUCT", struct {
		Product int
	}{p.Ordinal})


	port, err := x.comportProduct(p, errorLogger)
	if err == nil {
		err = w(productDevice{
			productData{
				app:            x,
				CurrentProduct: p,
			},
			port,
		})
	}

	if err != nil {
		errorLogger(p.ProductSerial, err.Error())
	}

}

func (x app) doDelayWithReadProducts(what string, duration time.Duration) error {

	series := products.NewSeries()
	defer series.Save(x.dbProducts, what)

	vars := ankat.MainVars1()
	if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {
		vars = append(vars, ankat.MainVars2()...)
	}
	iV, iP := 0, 0

	type ProductError struct {
		Serial ankat.ProductSerial
		Error  string
	}

	productErrors := map[ProductError]struct{}{}

	return x.uiWorks.Delay(what, duration, func() error {
		checkedProducts := x.DBProducts.CheckedProducts()
		if len(checkedProducts) == 0 {
			return errors.New(what + ": " + "не отмечено ни одного прибора")
		}

		if iP >= len(checkedProducts) {
			iP, iV = 0, 0
		}
		x.doProductDevice(checkedProducts[iP], func(productSerial ankat.ProductSerial, text string) {
			k := ProductError{productSerial, text}
			if _, exists := productErrors[k]; !exists {
				x.uiWorks.WriteError(productSerial, what+": "+text)
				productErrors[k] = struct{}{}
			}
		}, func(p productDevice) error {
			value, err := p.readVar(vars[iV])
			if err == nil {
				series.AddRecord(p.ProductSerial, vars[iV], value)
			}
			return nil
		})
		if iV < len(vars)-1 {
			iV++
			return nil
		}
		iV = 0
		if iP < len(checkedProducts)-1 {
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

func (x app) blowGas(gas ankat.GasCode) error {
	param := "delay_blow_nitrogen"
	what := fmt.Sprintf("продувка газа %s", gas.Description())
	if gas == ankat.GasNitrogen {
		param = "delay_blow_gas"
		what = "продувка азота"
	}
	if err := x.switchGas(gas); err != nil {
		return errors.Wrapf(err, "не удалось переключить клапан %s", gas.Description())
	}
	duration := x.ConfigSect("automatic_work").Minute(param)
	return x.doDelayWithReadProducts(what, duration)
}

func (x app) switchGas(n ankat.GasCode) error {
	port, err := x.comport("comport_gas")
	if err != nil {
		return errors.Wrap(err, "не удалось открыть СОМ порт газового блока")
	}
	req := modbus.NewSwitchGasOven(byte(n))
	_, err = port.GetResponse(req.Bytes())
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
	x.uiWorks.WriteLogf(0, worklog.Warning, "ошибка автоматической настройки была проигнорирована: %v", err)
	return nil
}

func (x app) setupTemperature(temperature float64) error {
	port, err := x.comport("comport_temperature")
	if err != nil {
		return errors.Wrap(err, "не удалось открыть СОМ порт термокамеры")
	}
	deltaTemperature := x.ConfigSect("automatic_work").Float64("delta_temperature")

	return termochamber.WaitForSetupTemperature(
		temperature-deltaTemperature, temperature+deltaTemperature,
		x.ConfigSect("automatic_work").Minute("timeout_temperature"),
		func() (float64, error) {
			return termochamber.T800Read(port)
		})
}

func (x app) holdTemperature(temperature float64) error {
	if err := x.setupTemperature(temperature); err != nil {
		errA := errors.Wrapf(err, "не удалось установить температуру %v\"С в термокамере", temperature)
		if err = x.promptErrorStopWork(errA); err != nil {
			return err
		}
	}
	duration := x.ConfigSect("automatic_work").Hour( "delay_temperature")
	x.uiWorks.WriteLogf(0, worklog.Info,
		"выдержка термокамеры на %v\"C: в настройках задана длительность %v", temperature, duration)
	return x.doDelayWithReadProducts(fmt.Sprintf("выдержка термокамеры на %v\"C", temperature), duration)
}
