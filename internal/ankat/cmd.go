package ankat

import (
	"fmt"
	"github.com/fpawel/goutils/serial/modbus"
)

type Cmd = modbus.DeviceCommandCode


const (
	CmdCorrectNull1 Cmd = 1
	CmdCorrectEnd1 Cmd = 2
	CmdCorrectNull2 Cmd = 4
	CmdCorrectEnd2 Cmd = 5
	CmdSetAddr Cmd = 7
	CmdNorm1 Cmd = 8
	CmdNorm2 Cmd = 9
	CmdSetGas1 Cmd = 16
	CmdSetGas2 Cmd = 17
	CmdCorrectTemperatureSensorOffset Cmd = 20
)

func FormatCmd(x Cmd) string {

	if x > 0x8000 {
		return fmt.Sprintf("команда %d: запись коэффициента %d", x, x - 0x8000)
	}

	if s,ok := commandStr[x]; ok {
		return s
	}
	return fmt.Sprintf("команда %d", x)
}

func Commands() (commands []Cmd){
	for cmd := range commandStr{
		commands = append(commands, cmd)
	}
	return
}

var commandStr = map[Cmd] string {
	CmdCorrectNull1:"Коррекция нуля 1",
	CmdCorrectNull2:"Коррекция нуля 2",
	CmdCorrectEnd1:"Коррекция конца шкалы 1",
	CmdCorrectEnd2:"Коррекция конца шкалы 2",
	CmdSetAddr:"Установка адреса MODBUS",
	CmdNorm1:"Нормировать канал 1 ИКД",
	CmdNorm2:"Нормировать канал 2 ИКД",
	CmdSetGas1:"Установить тип газа 1",
	CmdSetGas2:"Установить тип газа 2",
	CmdCorrectTemperatureSensorOffset:"Коррекция смещения датчика температуры",
}
