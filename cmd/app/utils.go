package main

import (
	"github.com/jmoiron/sqlx"
	"encoding/json"
)

func dbMustGet(db *sqlx.DB,  dest interface{}, query string, args ...interface{} ) {
	if err := db.Get(dest, query, args...); err != nil {
		panic(err)
	}
}

func dbMustSelect(db *sqlx.DB, dest interface{}, query string, args ...interface{}){
	if err := db.Select(dest, query, args...); err != nil {
		panic(err)
	}
}

func fmtErr(e error) string {
	if e != nil {
		return e.Error()
	}
	return ""
}

func dbMustOpen(fileName, query string) (db *sqlx.DB) {
	db = sqlx.MustConnect("sqlite3", appFolderFileName(fileName))
	if err := db.Ping(); err != nil {
		panic(err)
	}
	db.MustExec(query)
	return
}




func doNothing() error {
	return nil
}

func mustUnmarshalJson(b []byte, v interface{}) {
	if err := json.Unmarshal(b, v); err != nil {
		panic(err.Error() + ": " + string(b))
	}
}
