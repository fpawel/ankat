package dbutils

import (
	"database/sql"
	"fmt"
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

func MustNamedExec(db *sqlx.DB, query string, arg interface{}) sql.Result {
	r,err := db.NamedExec(query, arg)
	if err != nil {
		panic(err)
	}
	return r
}

func MustOpen(fileName, driverName string) (db *sqlx.DB) {
	db = sqlx.MustConnect(driverName,  fileName)
	if err := db.Ping(); err != nil {
		panic(err)
	}
	if err := db.Close(); err != nil {
		panic(err)
	}
	db = sqlx.MustConnect(driverName,  fileName)
	if err := db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("DB:", fileName)
	return
}
