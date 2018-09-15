package ankat

import "fmt"

type Cmd int


const (
	CmdCorrectNull1 = 1
	CmdCorrectEnd1 = 2
	CmdCorrectNull2 = 4
	CmdCorrectEnd2 = 5
	CmdSetAddr = 7
	CmdNorm1 = 8
	CmdNorm2 = 9
	CmdSetGas1 = 16
	CmdSetGas2 = 17
	CmdCorrectTemperatureSensorOffset = 20
)

func (x Cmd) What() string {

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
