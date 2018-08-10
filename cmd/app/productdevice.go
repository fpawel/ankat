package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/data/dataproducts"
	"github.com/fpawel/ankat/data/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/guartutils/fetch"
	"github.com/fpawel/guartutils/modbus"
	"github.com/fpawel/procmq"
	"github.com/fpawel/termochamber"
	"github.com/pkg/errors"
	"log"
	"math"
	"time"
)

type productDevice struct {
	port     *comport.Port
	product  Product
	pipe     *procmq.ProcessMQ
	workCtrl uiworks.Runner
	data     db
}

type readProductResult struct {
	Product, Var int
	Value        float64
	Error        string
}

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

func (x productDevice) fixVarsValues(vars []ankat.ProductVar) error {
	for _, pv := range vars {
		value, err := x.readVar(pv.Var)
		if err != nil {
			return errors.Wrapf(err, "сохранение: %s: РЕГ:%d ТОЧКА:%d", pv.Sect, pv.Var, pv.Point)
		}
		x.data.SetProductValue(x.product.Serial, pv, value)
		x.writeInfof("сохранение: %s: РЕГ:%d ТОЧКА:%d = %v", pv.Sect, pv.Var, pv.Point, value)
	}
	return nil
}

func (x productDevice) notifyConnected(err error, format string, a ...interface{}) {
	notifyProductConnected(x.product.Ordinal, x.pipe, err, format, a...)
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
			x.data.SetCoefficientValue(x.product.Serial, coefficient, value)
		}
	}
	x.notifyConnected(err, "K%d=%v", coefficient, value)

	for _, a := range x.data.Coefficients() {
		if a.Coefficient == coefficient {
			x.pipe.Send("READ_COEFFICIENT", readProductResult{
				Var:     a.Ordinal,
				Product: x.product.Ordinal,
				Error:   fmtErr(err),
				Value:   value,
			})
			break
		}
	}
	if err == nil {
		x.writeInfof("K%d=%v", coefficient, value)
	} else {
		x.writeErrorf("считывание K%d: %v", coefficient, err)
	}
	return value, err
}

func (x productDevice) readVar(v ankat.Var) (value float64, err error) {
	req := modbus.NewReadBCD(1, uint16(v))

	var bytes []byte
	bytes, err = x.port.Fetch(req.Bytes())
	if fetch.Canceled(err) {
		return 0, err
	}
	if err == nil {
		value, err = req.ParseBCDValue(bytes)
	}
	x.notifyConnected(err, "$%d=%v", v, value)
	for _, a := range x.data.Vars() {
		if a.Var == v {
			x.pipe.Send("READ_VAR", readProductResult{
				Var:     a.Ordinal,
				Product: x.product.Ordinal,
				Error:   fmtErr(err),
				Value:   value,
			})
			break
		}
	}

	return value, err
}

func (x productDevice) writeInitCoefficients() error {
	type cv struct {
		Coefficient ankat.Coefficient
		Value       float64
	}

	sensorCode := func(gas string, scale float64) float64 {
		switch gas {
		case "CH₄":
			return 16
		case "C₃H₈":
			return 14
		case "∑CH":
			return 15
		case "CO₂":
			switch scale {
			case 2:
				return 11
			case 5:
				return 12
			case 10:
				return 13
			}
		}
		log.Panicf("unknown sensor: gas: %q, scale: %v", gas, scale)
		return 0
	}

	val := func(name string) float64 {
		return x.data.CurrentPartyValue(name)
	}
	str := func(name string) string {
		return x.data.CurrentPartyValueStr(name)
	}

	xs := []cv{
		{2, float64(time.Now().Year())},
		{6, sensorCode(
			str("gas1"),
			val("scale1")),
		},
		{10, val("c_gas1")},
		{11, val("c_gas3ch1")},

		{23, 0},
		{24, 1},
		{25, 0},
		{26, 0},
		{27, 0},
		{28, 0},
		{29, 0},
		{30, 1},
		{31, 0},
		{32, 0},

		{43, 740},
		{44, 0},
		{45, 0},
		{46, 1},
		{47, 0},
	}

	if x.data.IsTwoConcentrationChannels() {

		xs2 := []cv{
			{15, sensorCode(
				str("gas2"),
				val("scale2")),
			},
			{19, val("c_gas1")},
			{20, val("c_gas3ch2")},
			{33, 0},
			{34, 1},
			{35, 0},
			{36, 0},
			{37, 0},
			{38, 0},
			{39, 0},
			{40, 1},
			{41, 0},
			{42, 0},
		}
		xs = append(xs, xs2...)
	}

	for _, c := range xs {
		//if x.workCtrl.Interrupted() {
		//	return errors.New("прервано")
		//}
		if err := x.writeCoefficientValue(c.Coefficient, c.Value); err != nil {
			return err
		}
	}
	return nil
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

func (x productDevice) sendCmd(cmd uint16, value float64) error {

	req := modbus.NewWriteCmdBCD(1, 0x16, cmd, value)
	b, err := x.port.Fetch(req.Bytes())
	if err == nil {
		err = req.CheckResponse16(b)
	}
	if fetch.NoAnswer(err) || modbus.ProtocolError(err) {
		//x.workCtrl.WriteLog(x.product.Serial, dataworks.Warning, err.Error())
		req = modbus.NewWriteCmdBCD(1, 0x10, cmd, value)
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

func (x productDevice) sendCmdLog(cmd uint16, value float64) error {
	err := x.sendCmd(cmd, value)
	if err == nil {
		x.writeInfof("%s: %v", x.data.formatCmd(cmd), value)
	} else {
		x.writeErrorf("%s: %v: %v", x.data.formatCmd(cmd), value, err)
	}
	return err
}

func (x productDevice) writeCoefficient(coefficient ankat.Coefficient) error {
	v, exists := x.data.CoefficientValue(x.product.Serial, coefficient)
	if !exists {
		x.workCtrl.WriteLog(x.product.Serial, dataworks.Warning, fmt.Sprintf(
			"запись К%d: значение коэффициента не задано", coefficient))
		return nil
	}

	err := x.sendCmd(uint16((0x80<<8)+coefficient), v)

	if fetch.Canceled(err) {
		return nil
	}
	x.notifyConnected(err, "K%d:=%v", coefficient, v)
	if err == nil {
		x.writeInfof("K%d:=%v", coefficient, v)
	} else {
		x.writeErrorf("запись K%d:=%v: %v", coefficient, v, err)
	}
	for _, a := range x.data.Coefficients() {
		if a.Coefficient == coefficient {
			x.pipe.Send("READ_COEFFICIENT", readProductResult{
				Var:     a.Ordinal,
				Product: x.product.Ordinal,
				Error:   fmtErr(err),
				Value:   v,
			})
			break
		}
	}
	return err
}

func (x productDevice) writeCoefficientValue(coefficient ankat.Coefficient, value float64) error {

	err := x.sendCmd(uint16((0x80<<8)+coefficient), value)

	if fetch.Canceled(err) {
		return nil
	}
	x.notifyConnected(err, "K%d:=%v", coefficient, value)
	if err == nil {
		dataproducts.SetCoefficientValue(x.data.dbProducts, x.product.Serial, coefficient, value)
		x.writeInfof("K%d:=%v", coefficient, value)
	} else {
		x.writeErrorf("запись K%d:=%v: %v", coefficient, value, err)
	}
	for _, a := range x.data.Coefficients() {
		if a.Coefficient == coefficient {
			x.pipe.Send("READ_COEFFICIENT", readProductResult{
				Var:     a.Ordinal,
				Product: x.product.Ordinal,
				Error:   fmtErr(err),
				Value:   value,
			})
			break
		}
	}
	return err
}

func (x productDevice) writeInfo(str string) {
	x.workCtrl.WriteLog(x.product.Serial, dataworks.Info, str)
}

func (x productDevice) writeInfof(format string, a ...interface{}) {
	x.writeInfo(fmt.Sprintf(format, a...))
}

func (x productDevice) writeError(str string) {
	x.workCtrl.WriteLog(x.product.Serial, dataworks.Error, str)
}

func (x productDevice) writeErrorf(format string, a ...interface{}) {
	x.writeError(fmt.Sprintf(format, a...))
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

	temperatureTermoChamber, err := termochamber.T800Read(portTermo)
	if err != nil {
		return wrapErr(errors.Wrap(err, "не удалось считать температуру термокамеры"))
	}

	temperatureCPU, err := x.readVar(10)
	if err != nil {
		return wrapErr(errors.Wrap(err, "не удалось считать температуру микроконтроллера"))
	}

	err = x.writeCoefficientValue(49, k49+temperatureTermoChamber-temperatureCPU)
	if err != nil {
		return wrapErr(errors.Wrap(err, "не удалось записать коэффициент 49"))
	}

	if math.Abs(temperatureTermoChamber-temperatureCPU) > 3 {
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
