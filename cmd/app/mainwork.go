package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/data/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/termochamber"
	"github.com/pkg/errors"
	"time"
)

func (x *app) mainWork() uiworks.Work {

	return uiworks.L("Настройка Анкат",
		x.workSetupTemperature(20),

		uiworks.S("Корректировка температуры CPU", func() error {
			portTermo, err := x.comport("temp")
			if err != nil {
				return errors.Wrap(err, "не удалось открыть СОМ порт термокамеры")
			}
			return x.doEachProductDevice(x.uiWorks.WriteLog, func(p productDevice) error {
				err := p.doAdjustTemperatureCPU(portTermo, 0)
				if err == nil {
					p.writeInfo("температура CPU откорректирована успешно")
					return nil
				}
				p.writeError(err.Error())
				if termochamber.IsHardwareError(err) {
					return err
				}
				return nil
			})
		}),

		x.workEachProduct("Корректировка температуры CPU", func(p productDevice) error {
			err := p.writeInitCoefficients()
			if err != nil {
				x.uiWorks.WriteLogf(p.product.Serial, dataworks.Error,
					"не удалось записать коэффициенты по умолчанию: %v", err)
			}
			return nil
		}),

		x.workSendSetWorkMode(2),
		x.workEachProduct("Установка значений коэффициентов по умолчанию", func(p productDevice) error {
			err := p.writeInitCoefficients()
			if err != nil {
				x.uiWorks.WriteLogf(p.product.Serial, dataworks.Error,
					"не удалось записать коэффициенты по умолчанию: %v", err)
			}
			return nil
		}),

		uiworks.S("Нормировка каналов измерения", func() error {
			if err := x.blowGas(ankat.GasNitrogen); err != nil {
				return errors.Wrap(err,
					"не удалось продуть азот")
			}
			if err := x.sendCmd(8, 100); err != nil {
				return errors.Wrap(err,
					"не удалось выполнить команду для нормировки канала 1")
			}
			if x.data.IsTwoConcentrationChannels() {
				err := x.sendCmd(9, 100)
				return errors.Wrap(err,
					"не удалось выполнить команду для нормировки канала 2")
			}
			return nil
		}),

		uiworks.S("Калибровка начала шкалы", func() error {
			if err := x.blowGas(ankat.GasNitrogen); err != nil {
				return errors.Wrap(err,
					"не удалось продуть азот")
			}
			nitrogenConcentration := x.data.CurrentPartyValue("c_gas1")
			if err := x.sendCmd(1, nitrogenConcentration); err != nil {
				return errors.Wrap(err,
					"не удалось выполнить команду калибровки начала шкалы канала 1")
			}
			if x.data.IsTwoConcentrationChannels() {
				if err := x.sendCmd(4, nitrogenConcentration); err != nil {
					return errors.Wrap(err,
						"не удалось выполнить команду калибровки начала шкалы канала 2")
				}
			}
			return nil
		}),

		uiworks.S("Калибровка чувствительности", func() error {
			if err := x.blowGas(ankat.GasChan1Middle1); err != nil {
				return errors.Wrap(err,
					"не удалось продуть середину шкалы канала 1")
			}
			concentration := x.data.CurrentPartyValue("c_gas2ch1")
			if err := x.sendCmd(2, concentration); err != nil {
				return errors.Wrap(err,
					"не удалось выполнить команду калибровки чувствительности канала 1")
			}
			if x.data.IsTwoConcentrationChannels() {
				if err := x.blowGas(ankat.GasChan2Middle); err != nil {
					return errors.Wrap(err,
						"не удалось продуть середину шкалы канала 2")
				}
				concentration = x.data.CurrentPartyValue("c_gas2ch2")
				if err := x.sendCmd(5, concentration); err != nil {
					return errors.Wrap(err,
						"не удалось выполнить команду калибровки чувствительности канала 2")
				}
			}
			x.doPause("калибровка чувствительности", 10*time.Second)
			return nil
		}),
	)
}

func (x app) workEachProduct(name string, work func(p productDevice) error) uiworks.Work {
	return uiworks.S(name, func() error {
		return x.doEachProductDevice(x.sendMessage, work)
	})
}

func (x *app) workSendSetWorkMode(mode float64) uiworks.Work {
	return x.workEachProduct(fmt.Sprintf("Установка режима работы: %v", mode), func(p productDevice) error {
		_ = p.sendSetWorkModeCmd(mode)
		return nil
	})
}

func (x *app) workNorming() uiworks.Work {

	return uiworks.S("Нормировка каналов измерения", func() error {
		if err := x.blowGas(ankat.GasNitrogen); err != nil {
			return err
		}
		if err := x.sendCmd(8, 100); err != nil {
			return err
		}
		if x.data.IsTwoConcentrationChannels() {
			return x.sendCmd(9, 100)
		}
		return nil
	})
}

func (x *app) workSetupTemperature(temperature float64) uiworks.Work {
	return uiworks.S(fmt.Sprintf("Установка термокамеры %v\"C", temperature), func() error {
		return x.setupTemperature(temperature)
	})
}
