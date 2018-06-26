package main

import (
	"fmt"
	"github.com/fpawel/ankat/data/dataproducts"
	"github.com/fpawel/ankat/data/dataworks"
	"github.com/fpawel/ankat/ui/uiworks"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/procmq"
	_ "github.com/mattn/go-sqlite3"
	"sync"
	"time"
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
			dbConfig:   dbMustOpen("config.db", SQLConfigDB),
			dbProducts: dbMustOpen("products.db", dataproducts.SQLProductsDB),
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

const SQLConfigDB = `
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS work_checked (
  work_order INTEGER PRIMARY KEY,
  checked TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS config (
  var NOT NULL PRIMARY KEY CHECK ( var != ''),
  value NOT NULL,
  name TEXT NOT NULL CHECK ( name != ''),
  section TEXT NOT NULL CHECK ( name != ''),
  sort_order INTEGER NOT NULL DEFAULT 0,
  type TEXT NOT NULL CHECK (type in ('bool','integer', 'real', 'text', 'comport_name', 'baud')),
  min, max
);

CREATE TABLE IF NOT EXISTS value_list(
  var NOT NULL CHECK ( var IS NOT ''),
  value NOT NULL,
  UNIQUE (var, value)
);

INSERT OR IGNORE INTO config (section, sort_order, var, name, type, min, max, value) VALUES
  ('Связь с приборами', 0, 'comport_products', 'имя СОМ-порта', 'comport_name', NULL, NULL, 'COM1' ),
  ('Связь с приборами', 1, 'comport_products_baud', 'скорость передачи, бод', 'baud', 2400, 256000, 9600 ),
  ('Связь с приборами', 2, 'comport_products_timeout', 'таймаут, мс', 'integer', 10, 10000, 1000 ),
  ('Связь с приборами', 3, 'comport_products_byte_timeout', 'длительность байта, мс', 'integer', 5, 200, 50 ),
  ('Связь с приборами', 4, 'comport_products_repeat_count', 'количество повторов', 'integer', 0, 10, 0 ),
  ('Связь с приборами', 5, 'comport_products_bounce_timeout', 'таймаут дребезга, мс', 'integer', 0, 1000, 0 ),
  ('Связь с приборами', 6, 'comport_products_bounce_limit', 'предел дребезга', 'integer', 0, 20, 0 )
  ;

INSERT OR IGNORE INTO config (section, sort_order, var, name, type, min, max, value) VALUES
  ('Пневмоблок', 0, 'comport_gas', 'имя СОМ-порта', 'comport_name', NULL, NULL, 'COM1' ),
  ('Пневмоблок', 1, 'comport_gas_baud', 'скорость передачи, бод', 'baud', 2400, 256000, 9600 ),
  ('Пневмоблок', 2, 'comport_gas_timeout', 'таймаут, мс', 'integer', 10, 10000, 1000 ),
  ('Пневмоблок', 3, 'comport_gas_byte_timeout', 'длительность байта, мс', 'integer', 5, 200, 50 ),
  ('Пневмоблок', 4, 'comport_gas_repeat_count', 'количество повторов', 'integer', 0, 10, 0 ),
  ('Пневмоблок', 5, 'comport_gas_bounce_timeout', 'таймаут дребезга, мс', 'integer', 0, 1000, 0 ),
  ('Пневмоблок', 6, 'comport_gas_bounce_limit', 'предел дребезга', 'integer', 0, 20, 0 );

INSERT OR IGNORE INTO config (section, sort_order, var, name, type, min, max, value) VALUES
  ('Термокамера', 0, 'comport_temp', 'имя СОМ-порта', 'comport_name', NULL, NULL, 'COM1' ),
  ('Термокамера', 1, 'comport_temp_baud', 'скорость передачи, бод', 'baud', 2400, 256000, 9600 ),
  ('Термокамера', 2, 'comport_temp_timeout', 'таймаут, мс', 'integer', 10, 10000, 1000 ),
  ('Термокамера', 3, 'comport_temp_byte_timeout', 'длительность байта, мс', 'integer', 5, 200, 50 ),
  ('Термокамера', 4, 'comport_temp_repeat_count', 'количество повторов', 'integer', 0, 10, 0 ),
  ('Термокамера', 5, 'comport_temp_bounce_timeout', 'таймаут дребезга, мс', 'integer', 0, 1000, 0 ),
  ('Термокамера', 6, 'comport_temp_bounce_limit', 'предел дребезга', 'integer', 0, 20, 0 );

INSERT OR IGNORE INTO value_list(var, value) VALUES
  ('gas1', 'CH₄'),
  ('gas1', 'C₃H₈'),
  ('gas1', '∑CH'),
  ('gas1', 'CO₂'),
  ('gas2', 'CH₄'),
  ('gas2', 'C₃H₈'),
  ('gas2', '∑CH'),
  ('gas2', 'CO₂'),
  ('scale1', 2),
  ('scale1', 5),
  ('scale1', 10),
  ('scale1', 100),
  ('scale2', 2),
  ('scale2', 5),
  ('scale2', 10),
  ('scale2', 100),
  ('sensors_count', 1),
  ('sensors_count', 2) ;

`
