package dbutils

import (
	"github.com/jmoiron/sqlx"
)

func MustGet(db *sqlx.DB,  dest interface{}, query string, args ...interface{} ) {
	if err := db.Get(dest, query, args...); err != nil {
		panic(err)
	}
}

func MustSelect(db *sqlx.DB, dest interface{}, query string, args ...interface{}){
	if err := db.Select(dest, query, args...); err != nil {
		panic(err)
	}
}



func MustOpen(fileName, driverName, query string) (db *sqlx.DB) {
	db = sqlx.MustConnect(driverName,  fileName)
	if err := db.Ping(); err != nil {
		panic(err)
	}
	db.MustExec(query)
	return
}
