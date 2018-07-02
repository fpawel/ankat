package main

import (
	"fmt"
	"github.com/fpawel/ankat/data/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"time"
)

func (x *app) mainWork() uiworks.Work {

	sleep10 := delay(x.uiWorks, "Задержка 10с", time.Second*10)

	return uiworks.L("Настройка Анкат",
		x.workSendSetWorkMode(2),
		x.eachProductWork("Установка значений коэффициентов по умолчанию", func(p productDevice) error {
			return p.sendSetWorkModeCmd(2)
		}),

		uiworks.L("Продувка воздухом",
			uiworks.S("Подать воздух", sleep10),
			uiworks.S("Задержка 5 минут", sleep10),
			uiworks.S("Выключить пневмоблок", sleep10),
			x.uiWorks.WorkDelay("Задержка 10 с", func() time.Duration {
				return 10 * time.Second
			}, nil),
		),
		uiworks.L("Продувка воздухом2",
			uiworks.S("Подать воздух2", sleep10),
			uiworks.S("Задержка 5 минут2", sleep10),
			uiworks.S("Выключить пневмоблок2", sleep10),
			x.uiWorks.WorkDelay("Задержка 10 с2", func() time.Duration {
				return 10 * time.Second
			}, nil),
		),
	)
}

func (x *app) workSendSetWorkMode(mode float64) uiworks.Work {
	return x.eachProductWork(fmt.Sprintf("Установка режима работы: %v", mode), func(p productDevice) error {
		_ = p.sendSetWorkModeCmd(mode)
		return nil
	})
}

func delay(u uiworks.Runner, what string, duration time.Duration) func() error {
	return func() error {
		u.WriteLog(0, dataworks.Debug, fmt.Sprintf("СОМ порт приборов: %v", "COM1"))
		return u.Delay(what, duration, doNothing)
	}
}
