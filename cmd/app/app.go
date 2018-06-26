package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/procmq"
	_ "github.com/mattn/go-sqlite3"
	"sync"
	"time"
	"github.com/fpawel/ankat/data/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
)

const uiExe = `F:\DELPHIPATH\src\github.com\fpawel\delphianalit\ankat\Win32\Debug\ankat.exe`

type app struct {
	workCtrl  uiworks.Runner
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
			dbConfig:   dbMustOpen("config.db", ankat.SQLConfigDB),
			dbProducts: dbMustOpen("products.db", ankat.SQLAnkat),
		},
		delphiApp: procmq.MustOpen("ANKAT"),
		comports:  make(map[string]comportState),
	}
	x.workCtrl = uiworks.NewRunner(x.delphiApp)

	x.delphiApp.Handle("CURRENT_WORK_CHECKED_CHANGED", func(bytes []byte) {
		var v struct {
			Ordinal    int
			CheckState string
		}
		mustUnmarshalJson(bytes, &v)
		x.data.dbConfig.MustExec(`INSERT OR REPLACE INTO work_checked VALUES ($1, $2);`,
			v.Ordinal, v.CheckState)
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
		x.runWork(uiworks.S(x.data.formatCmd(a.Cmd), func() error {
			return x.sendCmd(a.Cmd, a.Arg)
		}))
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
		x.workCtrl.Close()
		wg.Done()
	}()

	go func() {
		x.workCtrl.Run(x.data.dbProducts, x.data.dbConfig, x.mainWork().Task())
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
