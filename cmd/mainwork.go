package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/numeth"
	"github.com/fpawel/termochamber"
	"github.com/pkg/errors"
	"math"
	"time"
)


type interpolateChanSect = struct {
	sect ankat.Sect
	interpolateFunc func(ankat.ProductSerial, ankat.AnkatChan) ([]float64, []numeth.Coordinate, error)
}
type interpolateChanSectFunc func(ankat.AnkatChan) interpolateChanSect
type interpolateSectFunc func(ankat.ProductSerial) ([]float64, []numeth.Coordinate, error)

func (x app) mainWork() (result uiworks.Work) {
	t20gc := func() float64 {
		return 20
	}

	result = uiworks.L("Настройка Анкат",
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
			if err := x.blowGas(ankat.GasChan1End); err != nil {
				return errors.Wrap(err,
					"не удалось продуть конец шкалы канала 1")
			}
			concentration := x.db.CurrentPartyValue(ankat.GasChan1End.Var())
			if err := x.sendCmd(2, concentration); err != nil {
				return errors.Wrap(err,
					"не удалось выполнить команду калибровки чувствительности канала 1")
			}
			if x.db.IsTwoConcentrationChannels() {
				if err := x.blowGas(ankat.GasChan2End); err != nil {
					return errors.Wrap(err,
						"не удалось продуть конец шкалы канала 2")
				}
				concentration = x.db.CurrentPartyValue(ankat.GasChan2End.Var())
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
		x.workSaveCalculateLinSourceValues(),
		x.workInterpolateChanSect("Линеаризация", func(channel ankat.AnkatChan) interpolateChanSect {
			return interpolateChanSect{
				channel.Lin(),
				x.db.InterpolateLin,
			}
		}),

		x.workTemperaturePoint("низкая температура", func() float64 {
			return x.db.CurrentPartyValue("t-")
		}, 0),
		x.workTemperaturePoint("высокая температура", func() float64 {
			return x.db.CurrentPartyValue("t+")
		}, 2),
		x.workTemperaturePoint("НКУ", t20gc, 1),

		x.workInterpolateChanSect("Термокомпенсация нуля шкалы", func(channel ankat.AnkatChan) interpolateChanSect {
			return interpolateChanSect{
				sect:            channel.T0(),
				interpolateFunc: x.db.InterpolateT0,
			}
		}),

		x.workInterpolateChanSect("Термокомпенсация конца шкалы", func(channel ankat.AnkatChan) interpolateChanSect {
			return interpolateChanSect{
				sect:            channel.TK(),
				interpolateFunc: x.db.InterpolateTK,
			}
		}),
	)

	if x.db.IsPressureSensor(){
		result.AddChild( x.workInterpolateSect("Термокомпенсация давления", ankat.PT, x.db.InterpolatePT ) )
	}

	return
}

func (x *app) workInterpolateSect(what string, sect ankat.Sect, interpolateFunc interpolateSectFunc) (result uiworks.Work) {

	result = uiworks.L(what)

	fstr := func (s string) string{
		return fmt.Sprintf("%s: %s к-тов %s", sect, s, sect.CoefficientsStr(), )
	}

	result.AddChild(uiworks.S(fstr("расчёт"), func() error {
		x.doEachProductData(func(p productData) {
			values, xs, err := interpolateFunc(p.product.Serial)
			if err != nil {
				p.writeErrorf("расчёт %v не удался: %v", sect, err)
			} else {
				for  i := range values{
					values[i] = math.Round(values[i] * 1000000.) / 1000000.
				}
				p.writeInfof("расчёт %v: %v: [%s] = [%v]", sect, xs, sect.CoefficientsStr(), values)
			}
		})
		return nil
	}))
	result.AddChild(uiworks.S(fstr("ввод"), func() error {
		return x.doEachProductDevice(x.uiWorks.WriteError, func(p productDevice) error {
			return p.writeSectCoefficients(sect)
		})
		return nil
	}))

	return
}

func (x *app) workInterpolateChanSect(what string, ccs interpolateChanSectFunc) (result uiworks.Work) {

	result = uiworks.L(what)

	for _, ankatChan := range x.db.AnkatChannels() {
		concentrationChannel := ankatChan
		cs := ccs(concentrationChannel)

		fstr := func (s string) string{
			return fmt.Sprintf("%s: %s к-тов %s, канал %d", cs.sect, s, cs.sect.CoefficientsStr(), ankatChan)
		}

		result.AddChild(uiworks.S(fstr("расчёт"), func() error {
			x.doEachProductData(func(p productData) {
				values, xs, err := cs.interpolateFunc(p.product.Serial, concentrationChannel)
				if err != nil {
					p.writeErrorf("расчёт %v не удался: %v", cs.sect, err)
				} else {
					for  i := range values{
						values[i] = math.Round(values[i] * 1000000.) / 1000000.
					}
					p.writeInfof("расчёт %v: %v: [%s] = [%v]", cs.sect, xs, cs.sect.CoefficientsStr(), values)
				}
			})
			return nil
		}))
		result.AddChild(uiworks.S(fstr("ввод"), func() error {
			return x.doEachProductDevice(x.uiWorks.WriteError, func(p productDevice) error {
				return p.writeSectCoefficients(cs.sect)
			})
			return nil
		}))
	}
	return
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
			s += fmt.Sprintf("%s[%d]%s", a.Sect, a.Point, x.db.VarName(a.Var))
		}

		return x.workEachProduct(fmt.Sprintf("Снятие %s: %s: %s", what,
			gas.Description(), s), func(p productDevice) error {
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
	if x.db.IsTwoConcentrationChannels() {
		nitrogenVars = append(nitrogenVars, ankat.ProductVar{
			Sect: ankat.T02, Var: ankat.TppCh2,
		}, ankat.ProductVar{
			Sect: ankat.T02, Var: ankat.Var2Ch2,
		})
	}

	if x.db.IsPressureSensor() {
		nitrogenVars = append(nitrogenVars, ankat.ProductVar{
			Sect: ankat.PT, Var: ankat.VdatP,
		}, ankat.ProductVar{
			Sect: ankat.PT, Var: ankat.TppCh1,
		})
	}

	r.Name = fmt.Sprintf("Cнятие на температуре: %s", what)
	r.Children = append(r.Children,
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
	if x.db.IsTwoConcentrationChannels() {
		r.Children = append(r.Children,
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
	r.Children = append(r.Children, x.workBlowGas(ankat.GasNitrogen))

	return
}

func (x *app) workSaveCalculateLinSourceValues() (r uiworks.Work) {
	r.Name = "Снятие линеаризации"

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
		r.Children = append(r.Children, uiworks.S(gas.Description(), func() error {
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
		if x.db.IsTwoConcentrationChannels() {
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
