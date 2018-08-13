package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/termochamber"
	"github.com/pkg/errors"
	"time"
)

func (x app) mainWork() uiworks.Work {

	return uiworks.L("Настройка Анкат",
		x.workSetupAndHoldTemperature(20),

		uiworks.S("Корректировка температуры CPU", func() error {
			portTemperature, err := x.comport("temp")
			if err != nil {
				return errors.Wrap(err, "не удалось открыть СОМ порт термокамеры")
			}
			return x.doEachProductDevice(x.uiWorks.WriteLog, func(p productDevice) error {
				err := p.doAdjustTemperatureCPU(portTemperature, 0)
				if err == nil {
					p.writeInfo("температура CPU откорректирована успешно")
					return nil
				}

				if termochamber.IsHardwareError(err) {
					return err
				}
				p.writeError(err.Error())
				return nil
			})
		}),

		x.workSendSetWorkMode(2),
		x.workEachProduct("Установка коэффициентов", func(p productDevice) error {
			err := p.writeInitCoefficients()
			if err == nil {
				for _, k := range x.db.Coefficients() {
					if _, err = p.readCoefficient(k.Coefficient); err != nil {
						break
					}
				}
			}
			if err != nil {
				p.writeError(err.Error())
			}
			return nil
		}),

		uiworks.S("Нормировка", func() error {
			if err := x.blowGas(ankat.GasNitrogen); err != nil {
				return errors.Wrap(err,
					"не удалось продуть азот")
			}
			if err := x.sendCmd(8, 100); err != nil {
				return errors.Wrap(err,
					"не удалось выполнить команду для нормировки канала 1")
			}
			if x.db.IsTwoConcentrationChannels() {
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
			nitrogenConcentration := x.db.CurrentPartyValue("c_gas1")
			if err := x.sendCmd(1, nitrogenConcentration); err != nil {
				return errors.Wrap(err,
					"не удалось выполнить команду калибровки начала шкалы канала 1")
			}
			if x.db.IsTwoConcentrationChannels() {
				if err := x.sendCmd(4, nitrogenConcentration); err != nil {
					return errors.Wrap(err,
						"не удалось выполнить команду калибровки начала шкалы канала 2")
				}
			}
			x.doPause("калибровка начала шкалы", 10*time.Second)
			return nil
		}),

		uiworks.S("Калибровка чувствительности", func() error {
			if err := x.blowGas(ankat.GasChan1Middle1); err != nil {
				return errors.Wrap(err,
					"не удалось продуть середину шкалы канала 1")
			}
			concentration := x.db.CurrentPartyValue("c_gas2ch1")
			if err := x.sendCmd(2, concentration); err != nil {
				return errors.Wrap(err,
					"не удалось выполнить команду калибровки чувствительности канала 1")
			}
			if x.db.IsTwoConcentrationChannels() {
				if err := x.blowGas(ankat.GasChan2Middle); err != nil {
					return errors.Wrap(err,
						"не удалось продуть середину шкалы канала 2")
				}
				concentration = x.db.CurrentPartyValue("c_gas2ch2")
				if err := x.sendCmd(5, concentration); err != nil {
					return errors.Wrap(err,
						"не удалось выполнить команду калибровки чувствительности канала 2")
				}
			}
			x.doPause("калибровка чувствительности", 10*time.Second)

			if err := x.blowGas(ankat.GasNitrogen); err != nil {
				return errors.Wrap(err,
					"не удалось продуть азот после калибровки чувствительности")
			}

			return nil
		}),
		x.workLin(),
	)
}

func (x *app) workLin() (r uiworks.Work) {
	r.Name = "Снятие для линеаризации"

	gases := []ankat.GasCode{
		ankat.GasNitrogen,
		ankat.GasChan1Middle1,
	}
	if x.db.IsCO2() {
		gases = append(gases, ankat.GasChan1Middle2)
	}
	gases = append(gases, ankat.GasChan1End)
	if x.db.IsTwoConcentrationChannels() {
		gases = append(gases, ankat.GasChan2Middle)
		gases = append(gases, ankat.GasChan2End)
	}
	for _, gas := range gases {
		gas := gas
		r.Children = append(r.Children, uiworks.S(ankat.GasCodeDescription(gas), func() error {
			if err := x.blowGas(gas); err != nil {
				return err
			}
			return x.doEachProductDevice(x.uiWorks.WriteLog, func(p productDevice) error {
				return p.fixVarsValues(ankat.LinProductVars(gas))
			})
		}))
	}
	return
}

func (x app) workEachProduct(name string, work func(p productDevice) error) uiworks.Work {
	return uiworks.S(name, func() error {
		return x.doEachProductDevice(x.sendMessage, work)
	})
}

func (x app) workSendSetWorkMode(mode float64) uiworks.Work {
	return x.workEachProduct(fmt.Sprintf("Установка режима работы: %v", mode), func(p productDevice) error {
		_ = p.sendSetWorkModeCmd(mode)
		return nil
	})
}

func (x app) workNorming() uiworks.Work {

	return uiworks.S("Нормировка каналов измерения", func() error {
		if err := x.blowGas(ankat.GasNitrogen); err != nil {
			return err
		}
		if err := x.sendCmd(8, 100); err != nil {
			return err
		}
		if x.db.IsTwoConcentrationChannels() {
			return x.sendCmd(9, 100)
		}
		return nil
	})
}

func (x app) workSetupAndHoldTemperature(temperature float64) uiworks.Work {
	return uiworks.S(fmt.Sprintf("Установка термокамеры %v\"C", temperature), func() error {
		return x.setupAndHoldTemperature(temperature)
	})
}
