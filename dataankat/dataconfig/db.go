package dataconfig

import (
	"github.com/jmoiron/sqlx"
)


func Initialize(db *sqlx.DB, fileName string)  {
	db.MustExec(`ATTACH DATABASE ? AS dbconfig`, fileName)
	db.MustExec(SQLConfig)
	for _,s := range []string{
		"comport_products", "comport_gas", "comport_temperature",
	} {
		_,err := db.NamedExec(SQLComport, map[string]interface{}{"section_name": s} )
		if err != nil {
			panic(err)
		}
	}
}

