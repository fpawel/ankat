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
	"strings"
	"sync"
	"time"
)

//go:generate go run ./gen_sql_str/main.go

type app struct {
	uiWorks   uiworks.Runner
	db        dataankat.DBAnkat
	delphiApp *procmq.ProcessMQ
	comports  map[string]comportState
}

type comportState struct {
	comport *comport.Port
	err     error
}

type logger = func(productSerial ankat.ProductSerial, level dataworks.Level, text string)
type errorLogger = func(productSerial ankat.ProductSerial, text string)

func runApp() {

	x := &app{
		db: dataankat.MustOpen(),
		delphiApp: procmq.MustOpen("ANKAT"),
		comports:  make(map[string]comportState),
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




	//x.db.EnsurePartyExists()

	x.uiWorks = uiworks.NewRunner(x.delphiApp)

	x.delphiApp.Handle("CURRENT_WORK_STOP", func([]byte) interface {}{
		x.uiWorks.Interrupt()
		for _,serialPort := range x.comports{
			serialPort.comport.Interrupt()
		}
		return nil
	})

	x.delphiApp.Handle("CURRENT_WORK_CHECKED_CHANGED", func(bytes []byte) interface{}{
		var v struct {
			Ordinal    int
			CheckState string
		}
		mustUnmarshalJson(bytes, &v)
		x.db.DBConfig.DB.MustExec(`INSERT OR REPLACE INTO work_checked VALUES ($1, $2);`,
			v.Ordinal, v.CheckState)
		return nil
	})

	x.delphiApp.Handle("RUN_MAIN_WORK", func(b []byte) interface {}{
		n, err := strconv.ParseInt(string(b), 10, 64)
		if err != nil {
			panic(err)
		}
		x.runWork(int(n), x.mainWork())
		return nil
	})

	x.delphiApp.Handle("READ_VARS", func([]byte) interface {} {
		x.runReadVarsWork()
		return nil
	})
	x.delphiApp.Handle("READ_COEFFICIENTS", func([]byte) interface {}{
		x.runReadCoefficientsWork()
		return nil
	})

	x.delphiApp.Handle("WRITE_COEFFICIENTS", func([]byte) interface {}{
		x.runWriteCoefficientsWork()
		return nil
	})

	x.delphiApp.Handle("MODBUS_CMD", func(b []byte) interface {}{
		var a struct {
			Cmd uint16
			Arg float64
		}
		mustUnmarshalJson(b, &a)
		x.runWork(0, uiworks.S("Отправка команды", func() error {
			return x.sendCmd(a.Cmd, a.Arg)
		}))
		return nil
	})

	x.delphiApp.Handle("SEND_SET_WORK_MODE", func(b []byte) interface {} {
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
		return nil
	})

	x.delphiApp.Handle("CURRENT_WORKS", func(bytes []byte) interface {} {
		return x.mainWork().Task().Info(x.db.DBProducts.DB)
	})

	x.delphiApp.Handle("PARTY_INFO", func(bytes []byte) interface {} {
		partyID := ankat.PartyID(mustParseInt64(bytes))
		str := view.Party( x.db.PartyInfo(partyID), x.db.VarName )

		return str
	})



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