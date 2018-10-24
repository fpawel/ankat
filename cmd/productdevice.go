package main

import (
	"context"
	"fmt"
	"github.com/fpawel/ankat/internal/ankat"
	"github.com/fpawel/ankat/internal/db/worklog"
	"github.com/fpawel/goutils/procmq"
	"github.com/fpawel/goutils/serial/comport"
	"github.com/fpawel/goutils/serial/modbus"
	"github.com/fpawel/goutils/serial/termochamber"
	"github.com/pkg/errors"
	"math"
	"time"
)

type productDevice struct {
	productData
	port *comport.Port
}

type readProductVarResult struct {
	ProductOrder  int
	ProductSerial ankat.ProductSerial
	VarOrder      int
	Var           ankat.Var
	VarName       string `json:",omitempty"`
	Value         float64
	Error         string `json:",omitempty"`
}

type readProductCoefficientResult struct {
	ProductOrder     int
	ProductSerial    ankat.ProductSerial
	CoefficientOrder int
	Coefficient      ankat.Coefficient
	CoefficientName  string
	Value            float64
	Error            string `json:",omitempty"`
}

type CoefficientValues = map[ankat.Coefficient]float64

func notifyProductConnected(productOrdinal int, pipe *procmq.ProcessMQ, err error, format string, a ...interface{}) {
	if err == context.Canceled {
		return
	}
	var b struct {
		Product int
		Ok      bool
		Text    string
	}
	b.Product = productOrdinal
	if err == nil {
		b.Ok = true
		b.Text = fmt.Sprintf(format, a...)
	} else {
		b.Text = err.Error()
	}
	pipe.Send("PRODUCT_CONNECTED", b)
}

func (x productDevice) notifyCoefficient(coefficient ankat.Coefficient, value float64, err error) {
	coefficientInfo := x.app.DBProducts.Coefficient(coefficient)
	x.app.delphiApp.Send("READ_COEFFICIENT", readProductCoefficientResult{
		CoefficientName:  coefficientInfo.Name,
		CoefficientOrder: coefficientInfo.Ordinal,
		Coefficient:      coefficient,
		ProductOrder:     x.Ordinal,
		ProductSerial:    x.ProductSerial,
		Error:            fmtErr(err),
		Value:            value,
	})
}

func (x productDevice) fixVarsValues(vars []ankat.ProductVar) error {
	for _, pv := range vars {
		value, err := x.readVar(pv.Var)

		s := fmt.Sprintf("%s:%s[%d]", pv.Sect, x.app.DBProducts.Var(pv.Var).Name, pv.Point)

		if err != nil {
			return errors.Wrapf(err, "сохранение: %s", s)
		}
		x.SetValue(pv, value)
		x.writeInfof("сохранение: %s = %v", s, value)
	}
	return nil
}

func (x productDevice) fixMainError(vars []ankat.ProductVar) error {
	for _, pv := range vars {
		value, err := x.readVar(pv.Var)

		s := fmt.Sprintf("%s:%s[%d]", pv.Sect, x.app.DBProducts.Var(pv.Var).Name, pv.Point)

		if err != nil {
			return errors.Wrapf(err, "сохранение: %s", s)
		}
		x.SetValue(pv, value)
		x.writeInfof("сохранение: %s = %v", s, value)
	}
	return nil
}

func (x productDevice) notifyConnected(err error, format string, a ...interface{}) {
	notifyProductConnected(x.Ordinal, x.app.delphiApp, err, format, a...)
}

func (x productDevice) writeCoefficient(coefficient ankat.Coefficient) error {
	value, exists := x.CoefficientValue(coefficient)
	if !exists {
		x.writeLogf(worklog.Warning, "запись К%d: значение не задано", coefficient)
		return nil
	}

	err := x.sendCmd(ankat.Cmd((0x80<<8)+coefficient), value)

	if err == context.Canceled {
		return nil
	}

	if err == nil {
		x.writeInfof("запись K%d:=%v", coefficient, value)
	} else {
		x.writeErrorf("запись K%d:=%v: %v", coefficient, value, err)
	}

	x.notifyConnected(err, "K%d:=%v", coefficient, value)
	x.notifyCoefficient(coefficient, value, err)
	return err
}

func (x productDevice) readCoefficient(coefficient ankat.Coefficient) ( float64,  error) {
	value, err := modbus.ReadCoefficient(x.port, 1, coefficient)
	x.notifyConnected(err, "K%d=%v", coefficient, value)
	x.notifyCoefficient(coefficient, value, err)

	return value, err
}

func (x productDevice) readAndSaveCoefficient(coefficient ankat.Coefficient) (value float64, err error) {

	value,err = x.readCoefficient(coefficient)
	if err == nil {
		x.SetCoefficientValue(coefficient, value)
		x.writeInfof("считывание K%d=%v", coefficient, value)
	} else {
		x.writeErrorf("считывание K%d: %v", coefficient, err)
	}


	return value, err
}

func (x productDevice) readVar(ankatVar ankat.Var) (value float64, err error) {
	value,err = modbus.Read3BCD(x.port, 1, ankatVar)
	value += math.Sin(float64(time.Now().Second()) / 60)
	varInfo := x.app.DBProducts.Var(ankatVar)
	x.notifyConnected(err, "%s=%v", varInfo.Name, value)
	x.app.delphiApp.Send("READ_VAR", readProductVarResult{
		VarOrder:      varInfo.Ordinal,
		Var:           ankatVar,
		VarName:       varInfo.Name,
		ProductOrder:  x.Ordinal,
		ProductSerial: x.ProductSerial,
		Error:         fmtErr(err),
		Value:         value,
	})
	return
}

func (x productDevice) writeInitCoefficients() error {

	p := x.app.DBProducts.CurrentParty()

	xs := CoefficientValues{
		2: float64(time.Now().Year()),

		5:  p.Units1.Code(),
		6:  p.Gas1.Code(p.Scale1),
		7:  p.Scale1.Code(),
		10: p.VerificationGasConcentration(ankat.GasNitrogen),
		11: p.VerificationGasConcentration(ankat.GasChan1End),

		23: 0,
		24: 1,
		25: 0,
		26: 0,
		27: 0,
		28: 0,
		29: 0,
		30: 1,
		31: 0,
		32: 0,

		43: 740,
		44: 0,
		45: 0,
		46: 1,
		47: 0,

		14: p.Units2.Code(),
		15: p.Gas2.Code(p.Scale2),
		16: p.Scale2.Code(),
		19: p.VerificationGasConcentration(ankat.GasNitrogen),
		20: p.VerificationGasConcentration(ankat.GasChan2End),
		33: 0,
		34: 1,
		35: 0,
		36: 0,
		37: 0,
		38: 0,
		39: 0,
		40: 1,
		41: 0,
		42: 0,
	}
	return x.writeCoefficientValues(xs)
}

func (x productDevice) sendSetWorkModeCmd(mode float64) error {
	req := newAnkatSetWorkModeRequest(mode)
	b, err := x.port.GetResponse(req.Bytes())
	if err == nil {
		err = checkResponseAnkatSetWorkMode(req, b)
	}
	if err == nil {
		x.writeInfof("установка режима работы: %v", mode)
	} else {
		x.writeErrorf("установка режима работы: %v: %v", mode, err)
	}
	return err
}

func (x productDevice) sendCmd(cmd ankat.Cmd, value float64) error {
	return modbus.Write32Float1016(x.port, 1, cmd, value)
}

func (x productDevice) sendCmdLog(cmd ankat.Cmd, value float64) error {
	err := x.sendCmd(cmd, value)
	if err == nil {
		x.writeInfof("%s: %v", ankat.FormatCmd(cmd), value)
	} else {
		x.writeErrorf("%s: %v: %v", ankat.FormatCmd(cmd), value, err)
	}
	return err
}

func (x productDevice) writeCoefficientValues(coefficientValues CoefficientValues) error {

	for k, value := range coefficientValues {
		if err := x.writeCoefficientValue(k, value); err != nil {
			return err
		}
	}
	return nil
}

func (x productDevice) writeSectCoefficients(sect ankat.Sect) error {
	x.writeInfof("%v: ввод коэффициентов %s", sect, sect.CoefficientsStr())
	for i := sect.Coefficient0(); i < sect.Coefficient0()+sect.CoefficientsCount(); i++ {
		if err := x.writeCoefficient(i); err != nil {
			return err
		}
	}
	return nil
}

func (x productDevice) writeCoefficientValue(coefficient ankat.Coefficient, value float64) error {

	err := modbus.WriteCoefficient1016(x.port, 1, coefficient, value)
	if err == context.Canceled {
		return nil
	}

	x.notifyConnected(err, "K%d:=%v", coefficient, float6(value))
	if err == nil {
		x.SetCoefficientValue(coefficient, float6(value))
		x.writeInfof("K%d:=%v", coefficient, value)
	} else {
		x.writeErrorf("запись K%d:=%v: %v", coefficient, float6(value), err)
	}
	x.notifyCoefficient(coefficient, value, err)
	return err
}

func (x productDevice) doAdjustTemperatureCPU(portTermo *comport.Port, attemptNumber int) error {
	const maxAttemptsLimit = 10

	wrapErr := func(err error) error {
		return errors.Wrapf(err, "не удалось откалибровать датчик температуры (попытка %d из %d)",
			attemptNumber+1, maxAttemptsLimit)
	}

	k49, err := x.readAndSaveCoefficient(49)
	if err != nil {
		return wrapErr(errors.Wrap(err, "не удалось считать коэффициент 49"))
	}

	temperatureChamber, err := termochamber.T800Read(portTermo)
	if err != nil {
		return wrapErr(errors.Wrap(err, "не удалось считать температуру термокамеры"))
	}

	temperatureCPU, err := x.readVar(10)
	if err != nil {
		return wrapErr(errors.Wrap(err, "не удалось считать температуру микроконтроллера"))
	}

	err = x.writeCoefficientValue(49, k49+temperatureChamber-temperatureCPU)
	if err != nil {
		return wrapErr(errors.Wrap(err, "не удалось записать коэффициент 49"))
	}

	if math.Abs(temperatureChamber-temperatureCPU) > 3 {
		if attemptNumber < maxAttemptsLimit {
			return x.doAdjustTemperatureCPU(portTermo, attemptNumber+1)
		}
		return wrapErr(errors.New("превышено максимальное число попыток"))
	}
	return nil
}

func newAnkatSetWorkModeRequest(mode float64) modbus.Request {
	return modbus.Request{
		ProtocolCommandCode: 0x16,
		Addr:                1,
		Data:                append([]byte{0xA0, 0, 0, 2, 4}, modbus.BCD6(mode)...),
	}
}

func checkResponseAnkatSetWorkMode(x modbus.Request, b []byte) error {
	if err := x.CheckResponse(b); err != nil {
		return err
	}
	a := []byte{0xA0, 0, 0, 0}
	for i := range a {
		if a[i] != b[i+2] {
			return errors.Errorf("ошибка формата ответа на запрпос установки режима работы АНКАТ: % X != % X",
				a, b[2:6])
		}
	}
	return nil
}
