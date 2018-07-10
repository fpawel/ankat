package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/data/dataproducts"
	"github.com/fpawel/ankat/data/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/procmq"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"strings"
	"sync"
	"time"
)

type app struct {
	uiWorks   uiworks.Runner
	data      data
	delphiApp *procmq.ProcessMQ
	comports  map[string]comportState
}

type comportState struct {
	comport *comport.Port
	err     error
}

type logger = func(productSerial int, level dataworks.Level, text string)

func runApp() {

	x := &app{
		data: data{
			dbConfig:   dbMustOpen("config.db", SQLConfigDB),
			dbProducts: dbMustOpen("products.db", dataproducts.SQLAnkat),
		},
		delphiApp: procmq.MustOpen("ANKAT"),
		comports:  make(map[string]comportState),
	}

	// внести в базу данных записи технологических процессов
	for _, k := range ankat.TechProcesses() {
		x.data.dbProducts.MustExec(`INSERT OR IGNORE INTO tech_process (tech_process_id, name) VALUES (?,?);`,
			k, ankat.TechProcessName(k))
	}

	x.uiWorks = uiworks.NewRunner(x.delphiApp)

	x.delphiApp.Handle("CURRENT_WORK_CHECKED_CHANGED", func(bytes []byte) {
		var v struct {
			Ordinal    int
			CheckState string
		}
		mustUnmarshalJson(bytes, &v)
		x.data.dbConfig.MustExec(`INSERT OR REPLACE INTO work_checked VALUES ($1, $2);`,
			v.Ordinal, v.CheckState)
	})

	x.delphiApp.Handle("RUN_MAIN_WORK", func(b []byte) {
		n, err := strconv.ParseInt(string(b), 10, 64)
		if err != nil {
			panic(err)
		}
		x.runWork(int(n), x.mainWork())
	})

	x.delphiApp.Handle("READ_VARS", func([]byte) {
		x.runReadVarsWork()
	})
	x.delphiApp.Handle("READ_COEFFICIENTS", func([]byte) {
		x.runReadCoefficientsWork()
	})

	x.delphiApp.Handle("WRITE_COEFFICIENTS", func([]byte) {
		x.runWriteCoefficientsWork()
	})

	x.delphiApp.Handle("MODBUS_CMD", func(b []byte) {
		var a struct {
			Cmd uint16
			Arg float64
		}
		mustUnmarshalJson(b, &a)
		x.runWork(0, uiworks.S("Отправка команды", func() error {
			return x.sendCmd(a.Cmd, a.Arg)
		}))
	})

	x.delphiApp.Handle("SEND_SET_WORK_MODE", func(b []byte) {
		s := string(b)
		s = strings.Replace(s, ",", ".", -1)
		mode, err := strconv.ParseFloat(s, 64)
		if err != nil {
			x.delphiApp.Send("END_WORK", struct {
				Name, Error string
			}{"Отправка команды", err.Error()})
		} else {
			x.runWork(0, x.workSendSetWorkMode(mode))
		}
	})

	fmt.Println("delphiApp connecting...")
	if err := x.delphiApp.Connect(); err != nil {
		panic(err)
	}
	fmt.Println("delphiApp connected")
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		err := x.delphiApp.Run()
		if err != nil {
			panic(err)
		}
		x.uiWorks.Close()
		wg.Done()
	}()

	go func() {
		x.uiWorks.Run(x.data.dbProducts, x.data.dbConfig, x.mainWork().Task())
		wg.Done()
	}()

	x.delphiApp.Send("SETUP_CURRENT_WORKS", x.mainWork().Task().Info(x.data.dbProducts))
	fmt.Println("delphiApp started")
	wg.Wait()
}

func (x *app) sendMessage(productSerial int, level dataworks.Level, text string) {
	x.delphiApp.Send("CURRENT_WORK_MESSAGE", struct {
		Work          int
		CreatedAt     time.Time
		ProductSerial int
		Level         dataworks.Level
		Text          string
	}{
		0, time.Now(), productSerial, level, text,
	})
}
