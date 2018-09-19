package main

import (
	"fmt"
	"github.com/fpawel/ankat/internal/ankat"
	"github.com/fpawel/ankat/internal/db/config"
	"github.com/fpawel/ankat/internal/db/products"
	"github.com/fpawel/ankat/internal/db/worklog"
	"github.com/fpawel/ankat/internal/ui/uiworks"
	"github.com/fpawel/ankat/internal/view"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/procmq"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

//go:generate go run ./gen_sql_str/main.go

type app struct {
	uiWorks    uiworks.Runner
	dbProducts *sqlx.DB
	dbCfg      *sqlx.DB
	delphiApp  *procmq.ProcessMQ
	comports   comport.Comports

	DBProducts products.DB
}

type errorLogger = func(productSerial ankat.ProductSerial, text string)

func runApp(waitPeer bool) {


	x := &app{
		dbProducts: products.MustOpen(ankat.AppDataFileName( "products.db")),
		dbCfg:      config.MustOpen(ankat.AppDataFileName( "config.db")),
		delphiApp:  procmq.MustOpen("ANKAT"),
		comports:   comport.Comports{},
	}
	x.DBProducts.DB = x.dbProducts


	if !x.DBProducts.PartyExists(){
		fmt.Println("must create party", ankat.AppFileName("ankat_newparty.exe"))
		cmd := exec.Command(ankat.AppFileName("ankat_newparty.exe") )
		if err := cmd.Start(); err != nil {
			panic(err)
		}
		if err := cmd.Wait(); err != nil {
			panic(err)
		}
		if !x.DBProducts.PartyExists(){
			fmt.Println("not created")
			return
		}
	}

	x.uiWorks = uiworks.NewRunner(x.delphiApp)

	// роуты, через которые приходят команды от клиентского приложения
	x.registerRoutes()

	fmt.Print("peer: connecting...")

	if !waitPeer {
		if err := exec.Command(ankat.AppFileName("ankatui.exe")).Start(); err != nil {
			panic(err)
		}
	}

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
		x.uiWorks.Run(x.dbProducts, x.dbCfg, x.mainWork().Task())
		wg.Done()
	}()

	fmt.Println("peer application started")
	wg.Wait()
}

func (x *app) ConfigSect(sect string) config.Section {
	return config.Section{
		Section:sect,
		DB:x.dbCfg,
	}
}

func (x *app) sendErrorMessage(productSerial ankat.ProductSerial, text string) {
	x.sendMessage(productSerial, worklog.Error, text)
}

func (x *app) sendMessage(productSerial ankat.ProductSerial, level worklog.Level, text string) {
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
		Level         worklog.Level
		Text          string
	}{
		workIndex, work, time.Now(), productSerial, level, text,
	})
}

// registerRoutes создаётроуты, через которые приходят команды от клиентского приложения
func (x *app) registerRoutes(){
	for msg,fun := range map[string] func([]byte) interface{} {

		"MODBUS_COMMANDS": func([]byte) interface {}{
			type a = struct {
				Cmd ankat.Cmd
				Str string
			}
			var payload struct {
				Items []a
			}
			for _,v := range ankat.Commands() {
				payload.Items = append(payload.Items, a{v, ankat.FormatCmd(v)})
			}
			return payload
		},

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
		"SET_COEFFICIENT": func(b []byte) interface {}{
			var a struct {
				Product, Coefficient int
			}
			mustUnmarshalJson(b, &a)
			x.runSetCoefficient(a.Product, a.Coefficient)
			return nil
		},
		"MODBUS_CMD": func(b []byte) interface {}{
			var a struct {
				Cmd ankat.Cmd
				Arg float64
			}
			mustUnmarshalJson(b, &a)
			x.runWork(0, uiworks.S("Отправка команды", func() error {
				return x.sendCmd(a.Cmd, a.Arg)
			}))
			return nil
		},
		"CURRENT_WORKS": func(bytes []byte) interface {} {
			return x.mainWork().Task().Info(x.dbProducts)
		},
		"PARTY_INFO": func(bytes []byte) interface {} {
			partyID := ankat.PartyID(mustParseInt64(bytes))
			str := view.Party( x.dbProducts, partyID )
			return str
		},
	} {
		x.delphiApp.Handle(msg, fun)
	}
}
