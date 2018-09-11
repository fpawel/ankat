package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dataworks"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/guartutils/fetch"
	"github.com/fpawel/guartutils/modbus"
	"github.com/fpawel/procmq"
	"github.com/fpawel/termochamber"
	"github.com/pkg/errors"
	"math"
	"time"
)

type productDevice struct {
	productData
	port *comport.Port
}

type readProductVarResult struct {
	ProductOrder,
	VarOrder int
	Var           ankat.Var
	ProductSerial ankat.ProductSerial
	VarName       string `json:",omitempty"`
	Value         float64
	Error         string `json:",omitempty"`
}

type readProductCoefficientResult struct {
	ProductOrder,
	CoefficientOrder int
	Coefficient     ankat.Coefficient
	ProductSerial   ankat.ProductSerial
	Value           float64
	Error           string `json:",omitempty"`
}

type CoefficientValues = map[ankat.Coefficient]float64

func notifyProductConnected(productOrdinal int, pipe *procmq.ProcessMQ, err error, format string, a ...interface{}) {
	if fetch.Canceled(err) {
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
	x.app.delphiApp.Send("READ_COEFFICIENT", readProductCoefficientResult{
		CoefficientOrder: x.app.DBProducts.GetCoefficientOrdinal(coefficient),
		Coefficient:      coefficient,
		ProductOrder:     x.app.DBProducts.CurrentProductOrderBySerial(x.ProductSerial),
		Error:            fmtErr(err),
		Value:            value,
	})
}

func (x productDevice) fixVarsValues(vars []ankat.ProductVar) error {
	for _, pv := range vars {
		value, err := x.readVar(pv.Var)

		s := fmt.Sprintf("%s:%s[%d]", pv.Sect, x.app.DBProducts.VarName(pv.Var), pv.Point)

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

		s := fmt.Sprintf("%s:%s[%d]", pv.Sect, x.app.DBProducts.VarName(pv.Var), pv.Point)

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
		x.writeLogf(dataworks.Warning, "запись К%d: значение не задано", coefficient)
		return nil
	}

	err := x.sendCmd(ankat.Cmd((0x80<<8)+coefficient), value)

	if fetch.Canceled(err) {
		return nil
	}
	x.notifyConnected(err, "K%d:=%v", coefficient, value)
	if err == nil {
		x.writeInfof("запись K%d:=%v", coefficient, value)
	} else {
		x.writeErrorf("запись K%d:=%v: %v", coefficient, value, err)
	}

	x.notifyCoefficient(coefficient, value, err)

	return err
}

func (x productDevice) readCoefficient(coefficient ankat.Coefficient) (value float64, err error) {

	req := modbus.NewReadCoefficient(1, int(coefficient))
	var bytes []byte
	bytes, err = x.port.Fetch(req.Bytes())

	if fetch.Canceled(err) {
		return 0, err
	}
	if err == nil {
		value, err = req.ParseBCDValue(bytes)
		if err == nil {
			x.SetCoefficientValue(coefficient, value)
		}
	}
	x.notifyConnected(err, "K%d=%v", coefficient, value)

	if err == nil {
		x.writeInfof("считывание K%d=%v", coefficient, value)
	} else {
		x.writeErrorf("считывание K%d: %v", coefficient, err)
	}

	x.notifyCoefficient(coefficient, value, err)
	return value, err
}

func (x productDevice) readVar(ankatVar ankat.Var) (value float64, err error) {
	req := modbus.NewReadBCD(1, uint16(ankatVar))

	var bytes []byte
	bytes, err = x.port.Fetch(req.Bytes())
	if err != nil {
		return
	}
	value, err = req.ParseBCDValue(bytes)
	if err != nil {
		return
	}
	//x.notifyConnected(err, "$%d=%v", v, value)
	for _, a := range x.app.DBProducts.Vars() {
		if a.Var == ankatVar {
			x.app.delphiApp.Send("READ_VAR", readProductVarResult{
				VarOrder:     a.Ordinal,
				Var: ankatVar,
				VarName: x.app.DBProducts.VarName(ankatVar),
				ProductOrder: x.Ordinal,
				Error:        fmtErr(err),
				Value:        value,
			})
			break
		}
	}

	return value, err
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
	b, err := x.port.Fetch(req.Bytes())
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

	req := modbus.NewWriteCmdBCD(1, 0x16, uint16(cmd), value)
	b, err := x.port.Fetch(req.Bytes())
	if err == nil {
		err = req.CheckResponse16(b)
	}
	if fetch.NoAnswer(err) || modbus.ProtocolError(err) {
		//x.workCtrl.WriteLog(x.product.Serial, dataworks.Warning, err.Error())
		req = modbus.NewWriteCmdBCD(1, 0x10, uint16(cmd), value)
		b, err = x.port.Fetch(req.Bytes())
		if err == nil {
			err = req.CheckResponse16(b)
		}
	}
	if fetch.Canceled(err) {
		return nil
	}
	return err
}

func (x productDevice) sendCmdLog(cmd ankat.Cmd, value float64) error {
	err := x.sendCmd(cmd, value)
	if err == nil {
		x.writeInfof("%s: %v", cmd.What(), value)
	} else {
		x.writeErrorf("%s: %v: %v", cmd.What(), value, err)
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

	err := x.sendCmd(ankat.Cmd((0x80<<8)+coefficient), value)

	if fetch.Canceled(err) {
		return nil
	}
	x.notifyConnected(err, "K%d:=%v", coefficient, float6(value))
	if err == nil {
		x.SetCoefficientValue(coefficient, float6(value))
		x.writeInfof("K%d:=%v", coefficient, value)
	} else {
		x.writeErrorf("запись K%d:=%v: %v", coefficient, float6(value), err)
	}
	for _, a := range x.app.DBProducts.Coefficients() {
		if a.Coefficient == coefficient {
			x.app.delphiApp.Send("READ_COEFFICIENT", readProductResult{
				VarOrder:     a.Ordinal,
				ProductOrder: x.Ordinal,
				Error:        fmtErr(err),
				Value:        value,
			})
			break
		}
	}
	return err
}

func (x productDevice) doAdjustTemperatureCPU(portTermo *comport.Port, attemptNumber int) error {
	const maxAttemptsLimit = 10

	wrapErr := func(err error) error {
		return errors.Wrapf(err, "не удалось откалибровать датчик температуры (попытка %d из %d)",
			attemptNumber+1, maxAttemptsLimit)
	}

	k49, err := x.readCoefficient(49)
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
		Cmd:  0x16,
		Addr: 1,
		Data: append([]byte{0xA0, 0, 0, 2, 4}, modbus.BCD6(mode)...),
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
