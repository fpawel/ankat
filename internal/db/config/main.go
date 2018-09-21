package config

import (
	"github.com/fpawel/goutils/dbutils"
	"github.com/jmoiron/sqlx"
)

func MustOpen(fileName string) (db *sqlx.DB) {
	db = dbutils.MustOpen(fileName, "sqlite3", )
	db.MustExec(SQLConfig)
	for _,s := range []string{
		"comport_products", "comport_gas", "comport_temperature",
	} {
		_,err := db.NamedExec(SQLComport, map[string]interface{}{"section_name": s} )
		if err != nil {
			panic(err)
		}
	}
	return
}

