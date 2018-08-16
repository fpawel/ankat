package main

import (
	"github.com/jmoiron/sqlx"
	"encoding/json"
	"github.com/fpawel/ankat"
	"strconv"
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
	db = sqlx.MustConnect("sqlite3",  ankat.AppDataFileName(fileName))
	if err := db.Ping(); err != nil {
		panic(err)
	}
	db.MustExec(query)
	return
}


func mustUnmarshalJson(b []byte, v interface{}) {
	if err := json.Unmarshal(b, v); err != nil {
		panic(err.Error() + ": " + string(b))
	}
}

func mustParseInt64(b []byte) int64 {
	v,err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		panic(err.Error() + ": " + string(b))
	}
	return v
}
