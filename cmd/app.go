package main

import (
	"fmt"
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat"
	"github.com/fpawel/ankat/dataankat/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/ankat/view"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/procmq"
	_ "github.com/mattn/go-sqlite3"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

//go:generate go run ./gen_sql_str/main.go

type app struct {
	uiWorks   uiworks.Runner
	db        dataankat.DBAnkat
	delphiApp *procmq.ProcessMQ
	comports  comport.Comports
}


type logger = func(productSerial ankat.ProductSerial, level dataworks.Level, text string)
type errorLogger = func(productSerial ankat.ProductSerial, text string)

func runApp() {

	x := &app{
		db: dataankat.MustOpen(),
		delphiApp: procmq.MustOpen("ANKAT"),
		comports:  comport.Comports{},
	}
	if !x.db.PartyExists(){
		fmt.Println("must create party", ankat.AppFileName("ankat_newparty.exe"))
		cmd := exec.Command(ankat.AppFileName("ankat_newparty.exe") )
		if err := cmd.Start(); err != nil {
			panic(err)
		}
		if err := cmd.Wait(); err != nil {
			panic(err)
		}
		if !x.db.PartyExists(){
			fmt.Println("not created")
			return
		}
	}

	x.uiWorks = uiworks.NewRunner(x.delphiApp)

	for msg,fun := range map[string] func([]byte) interface{} {
		"CURRENT_WORK_STOP":  func([]byte) interface {}{
			x.uiWorks.Interrupt()
			x.comports.Interrupt()
			return nil
		},
		"RUN_MAIN_WORK": func(b []byte) interface {}{
			n, err := strconv.ParseInt(string(b), 10, 64)
			if err != nil {
				panic(err)
			}
			x.runWork(int(n), x.mainWork())
			return nil
		},
		"READ_VARS": func([]byte) interface {} {
			x.runReadVarsWork()
			return nil
		},
		"READ_COEFFICIENTS": func([]byte) interface {}{
			x.runReadCoefficientsWork()
			return nil
		},
		"WRITE_COEFFICIENTS": func([]byte) interface {}{
			x.runWriteCoefficientsWork()
			return nil
		},
		"MODBUS_CMD": func(b []byte) interface {}{
			var a struct {
				Cmd uint16
				Arg float64
			}
			mustUnmarshalJson(b, &a)
			x.runWork(0, uiworks.S("Отправка команды", func() error {
				return x.sendCmd(a.Cmd, a.Arg)
			}))
			return nil
		},
		"CURRENT_WORKS": func(bytes []byte) interface {} {
			return x.mainWork().Task().Info(x.db.DBProducts.DB)
		},
		"PARTY_INFO": func(bytes []byte) interface {} {
			partyID := ankat.PartyID(mustParseInt64(bytes))
			str := view.Party( x.db.Party(partyID), x.db.VarName )
			return str
		},
	} {
		x.delphiApp.Handle(msg, fun)
	}

	fmt.Print("peer: connecting...")
	if err := x.delphiApp.Connect(); err != nil {
		panic(err)
	}
	fmt.Println("peer: connected")

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
		x.uiWorks.Run(x.db.DBProducts.DB, x.db.DBConfig.DB, x.mainWork().Task())
		wg.Done()
	}()

	fmt.Println("peer application started")
	wg.Wait()
}

func (x *app) sendErrorMessage(productSerial ankat.ProductSerial, text string) {
	x.sendMessage(productSerial, dataworks.Error, text)
}

func (x *app) sendMessage(productSerial ankat.ProductSerial, level dataworks.Level, text string) {
	workIndex := 0
	work := ""
	if t := x.uiWorks.CurrentRunTask(); t != nil {
		workIndex = t.Ordinal()
		work = t.Name()
	}

	x.delphiApp.Send("CURRENT_WORK_MESSAGE", struct {
		WorkIndex     int
		Work          string
		CreatedAt     time.Time
		ProductSerial ankat.ProductSerial
		Level         dataworks.Level
		Text          string
	}{
		workIndex, work, time.Now(), productSerial, level, text,
	})
}
