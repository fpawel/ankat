package main

import (
	"fmt"
	"github.com/fpawel/ankat/internal/ankat"
	"github.com/fpawel/ankat/internal/ui/uiworks"
	"github.com/fpawel/goutils/serial/termochamber"
	"github.com/pkg/errors"
	"os"
	"time"
)

func (x app) mainWork() uiworks.Work {
	t20gc := func() float64 {
		return 20
	}

	workInterpolateSect := func(sect ankat.Sect) uiworks.Work {

		return uiworks.S(fmt.Sprintf("%s: расчёт к-тов %s", sect, sect.CoefficientsStr()),
			func() error {
				x.doEachProductData(func(p productData) {
					p.interpolateSect(sect)
				})
				return nil
			})
	}

	workWriteSectCoefficients := func(sect ankat.Sect) uiworks.Work {
		return uiworks.S(fmt.Sprintf("%s: ввод к-тов %s", sect, sect.CoefficientsStr()), func() error {
			return x.doEachProductDevice(x.uiWorks.WriteError,
				func(p productDevice) error {
					return p.writeSectCoefficients(sect)
				})
			return nil
		})
	}



	workTemperature := uiworks.L("Термокомпенсация",
		uiworks.L("Снятие",
			x.workTemperaturePoint("Низкая температура", func() float64 {
				return x.DBProducts.CurrentParty().TemperatureMinus
			}, 0),
			x.workTemperaturePoint("Высокая температура", func() float64 {
				return x.DBProducts.CurrentParty().TemperaturePlus
			}, 2),
			x.workTemperaturePoint("НКУ", t20gc, 1),
		),
		uiworks.L("Расчёт",
			workInterpolateSect(ankat.T01),
			workInterpolateSect(ankat.TK1),
		),

		uiworks.L("Ввод",
			workWriteSectCoefficients(ankat.T01),
			workWriteSectCoefficients(ankat.TK1),
		),
	)

	workLin := uiworks.L("Линеаризация",
		x.workSaveCalculateLinSourceValues(),
		uiworks.L("Расчёт",
			workInterpolateSect(ankat.Lin1),
		),
		uiworks.L("Ввод",
			workWriteSectCoefficients(ankat.Lin1),
		),
	)

	if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {

		workTemperature.Children[1].AddChildren(
			workInterpolateSect(ankat.T02),
			workInterpolateSect(ankat.TK2))

		workTemperature.Children[2].AddChildren(
			workWriteSectCoefficients(ankat.T02),
			workWriteSectCoefficients(ankat.TK2))

		workLin.Children[1].AddChildren(
			workInterpolateSect(ankat.Lin2))

		workLin.Children[2].AddChildren(
			workWriteSectCoefficients(ankat.Lin2))
	}

	if x.DBProducts.CurrentParty().PressureSensor {
		workTemperature.Children[1].AddChildren(
			workInterpolateSect(ankat.PT))
		workTemperature.Children[2].AddChildren(
			workWriteSectCoefficients(ankat.PT))
	}

	return uiworks.L("Настройка Анкат",
		x.workHoldTemperature("НКУ", t20gc),
		uiworks.S("Корректировка температуры CPU", func() error {
			portTemperature, err := x.comport("comport_temperature")
			if err != nil {
				return errors.Wrap(err, "не удалось открыть СОМ порт термокамеры")
			}
			return x.doEachProductDevice(x.uiWorks.WriteError, func(p productDevice) error {
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
				for _, k := range x.DBProducts.Coefficients() {
					if _, err = p.readAndSaveCoefficient(k.Coefficient); err != nil {
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
			if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {
				err := x.sendCmd(9, 100)
				return errors.Wrap(err,
					"не удалось выполнить команду для нормировки канала 2")
			}
			return nil
		}),

		uiworks.L("Калибровка",
			uiworks.S("Начало шкалы", func() error {
				if err := x.blowGas(ankat.GasNitrogen); err != nil {
					return errors.Wrap(err,
						"не удалось продуть азот")
				}
				nitrogenConcentration := x.DBProducts.CurrentParty().VerificationGasConcentration(ankat.GasNitrogen)
				if err := x.sendCmd(1, nitrogenConcentration); err != nil {
					return errors.Wrap(err,
						"не удалось выполнить команду калибровки начала шкалы канала 1")
				}
				if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {
					if err := x.sendCmd(4, nitrogenConcentration); err != nil {
						return errors.Wrap(err,
							"не удалось выполнить команду калибровки начала шкалы канала 2")
					}
				}
				x.doPause("калибровка начала шкалы", 10*time.Second)
				return nil
			}),

			uiworks.S("Чувствительность", func() error {
				if err := x.blowGas(ankat.GasChan1End); err != nil {
					return errors.Wrap(err,
						"не удалось продуть конец шкалы канала 1")
				}
				concentration := x.DBProducts.CurrentParty().VerificationGasConcentration(ankat.GasChan1End)
				if err := x.sendCmd(2, concentration); err != nil {
					return errors.Wrap(err,
						"не удалось выполнить команду калибровки чувствительности канала 1")
				}
				if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {
					if err := x.blowGas(ankat.GasChan2End); err != nil {
						return errors.Wrap(err,
							"не удалось продуть конец шкалы канала 2")
					}
					concentration = x.DBProducts.CurrentParty().VerificationGasConcentration(ankat.GasChan2End)
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
		),
		workLin,

		workTemperature,
	)
}

func (x *app) workTemperaturePoint(what string, temperature func() float64, point ankat.Point) (r uiworks.Work) {

	workSave := func(gas ankat.GasCode, vars []ankat.ProductVar) uiworks.Work {
		for i := range vars {
			vars[i].Point = point
		}
		s := ""
		for _, a := range vars {
			if s != "" {
				s += ", "
			}
			s += fmt.Sprintf("%s[%d]%s", a.Sect, a.Point, x.DBProducts.Var(a.Var).Name)
		}

		return x.workEachProduct(fmt.Sprintf("Снятие %s: %s", gas.Description(), s), func(p productDevice) error {
			return p.fixVarsValues(vars)
		})
	}

	nitrogenVars := []ankat.ProductVar{
		{
			Sect: ankat.T01, Var: ankat.TppCh1,
		},
		{
			Sect: ankat.T01, Var: ankat.Var2Ch1,
		},
	}

	if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {
		nitrogenVars = append(nitrogenVars, ankat.ProductVar{
			Sect: ankat.T02, Var: ankat.TppCh2,
		}, ankat.ProductVar{
			Sect: ankat.T02, Var: ankat.Var2Ch2,
		})
	}

	if x.DBProducts.CurrentParty().PressureSensor {
		nitrogenVars = append(nitrogenVars, ankat.ProductVar{
			Sect: ankat.PT, Var: ankat.VdatP,
		}, ankat.ProductVar{
			Sect: ankat.PT, Var: ankat.TppCh1,
		})
	}

	r = uiworks.L(what,
		x.workHoldTemperature(what, temperature),
		x.workBlowGas(ankat.GasNitrogen),
		workSave(ankat.GasNitrogen, nitrogenVars),
		x.workBlowGas(ankat.GasChan1End),
		workSave(ankat.GasChan1End, []ankat.ProductVar{
			{
				Sect: ankat.TK1, Var: ankat.TppCh1,
			},
			{
				Sect: ankat.TK1, Var: ankat.Var2Ch1,
			},
		}),
	)

	if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {
		r.AddChildren(
			x.workBlowGas(ankat.GasChan2End),
			workSave(ankat.GasChan2End, []ankat.ProductVar{
				{
					Sect: ankat.TK2, Var: ankat.TppCh2,
				},
				{
					Sect: ankat.TK2, Var: ankat.Var2Ch2,
				},
			}))
	}
	r.AddChild(x.workBlowGas(ankat.GasNitrogen))

	return
}

func (x *app) workSaveCalculateLinSourceValues() (r uiworks.Work) {
	r.Name = "Снятие"

	gases := []ankat.GasCode{
		ankat.GasNitrogen,
		ankat.GasChan1Middle1,
	}

	if x.DBProducts.CurrentParty().IsCO2() {
		gases = append(gases, ankat.GasChan1Middle2)
	}
	gases = append(gases, ankat.GasChan1End)
	if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {
		gases = append(gases, ankat.GasChan2Middle)
		gases = append(gases, ankat.GasChan2End)
	}
	for _, gas := range gases {
		gas := gas
		r.AddChild(uiworks.S(gas.Description(), func() error {
			if err := x.blowGas(gas); err != nil {
				return err
			}
			return x.doEachProductDevice(x.uiWorks.WriteError, func(p productDevice) error {
				return p.fixVarsValues(ankat.LinProductVars(gas))
			})
		}))
	}
	return
}

func (x app) workEachProduct(name string, work func(p productDevice) error) uiworks.Work {
	return uiworks.S(name, func() error {
		return x.doEachProductDevice(x.sendErrorMessage, work)
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
		if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {
			return x.sendCmd(9, 100)
		}
		return nil
	})
}

func (x app) workHoldTemperature(what string, temperature func() float64) uiworks.Work {
	return uiworks.S(fmt.Sprintf("Установка термокамеры: %s", what), func() error {
		return x.holdTemperature(temperature())
	})
}

func (x app) workBlowGas(gas ankat.GasCode) uiworks.Work {
	return uiworks.S(fmt.Sprintf("Продувка %s", gas.Description()), func() error {
		return x.blowGas(gas)
	})
}

func (x *app) workMainErrorTemperaturePoint(what string, temperature func() float64, point ankat.Point) (r uiworks.Work) {

	r = uiworks.L(what,
		x.workHoldTemperature(what, temperature),
	)
	os.ErrClosed = nil

	for i, gas := range []ankat.GasCode{ankat.GasNitrogen, ankat.GasChan1Middle1, ankat.GasChan1End} {
		r.AddChild(uiworks.S(gas.Description(), func() error {
			if err := x.blowGas(gas); err != nil {
				return err
			}
			return x.doEachProductDevice(x.uiWorks.WriteError, func(p productDevice) error {
				return p.fixVarsValues([]ankat.ProductVar{
					{
						Point: ankat.Point(i),
						Var:   ankat.CoutCh1,
					},
				})
			})

		}))
	}

	workSave := func(gas ankat.GasCode, vars []ankat.ProductVar) uiworks.Work {
		for i := range vars {
			vars[i].Point = point
		}
		s := ""
		for _, a := range vars {
			if s != "" {
				s += ", "
			}
			s += fmt.Sprintf("%s[%d]%s", a.Sect, a.Point, x.DBProducts.Var(a.Var).Name)
		}

		return x.workEachProduct(fmt.Sprintf("Снятие %s: %s", gas.Description(), s), func(p productDevice) error {
			return p.fixVarsValues(vars)
		})
	}

	nitrogenVars := []ankat.ProductVar{
		{
			Sect: ankat.T01, Var: ankat.TppCh1,
		},
		{
			Sect: ankat.T01, Var: ankat.Var2Ch1,
		},
	}

	if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {
		nitrogenVars = append(nitrogenVars, ankat.ProductVar{
			Sect: ankat.T02, Var: ankat.TppCh2,
		}, ankat.ProductVar{
			Sect: ankat.T02, Var: ankat.Var2Ch2,
		})
	}

	if x.DBProducts.CurrentParty().PressureSensor {
		nitrogenVars = append(nitrogenVars, ankat.ProductVar{
			Sect: ankat.PT, Var: ankat.VdatP,
		}, ankat.ProductVar{
			Sect: ankat.PT, Var: ankat.TppCh1,
		})
	}

	if x.DBProducts.CurrentParty().IsTwoConcentrationChannels() {
		r.AddChildren(
			x.workBlowGas(ankat.GasChan2End),
			workSave(ankat.GasChan2End, []ankat.ProductVar{
				{
					Sect: ankat.TK2, Var: ankat.TppCh2,
				},
				{
					Sect: ankat.TK2, Var: ankat.Var2Ch2,
				},
			}))
	}
	r.AddChild(x.workBlowGas(ankat.GasNitrogen))

	return
}
