package dataworks

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"time"
)


type CurrentWorkMessage = struct {
	Work      int       `db:"work_index"`
	CreatedAt time.Time `db:"created_at"`
	ProductSerial int `db:"product_serial"`
	Level     Level     `db:"level"`
	Text      string    `db:"message"`
}

type Level int

const (
	Trace Level = iota
	Debug
	Info
	Warning
	Error
)


type WriteRecord struct {
	Works         []Work
	ProductSerial int
	Level         Level
	Text          string
}

type Work struct {
	Name  string
	Index int
}

func EnsureCurrentWorks(x *sqlx.DB, works []Work) {

	for i, work := range works {

		var exists bool
		err := x.Get(&exists, `SELECT exists(SELECT * FROM last_work_log WHERE work_index=$1)`, work.Index)
		if err != nil {
			panic(err)
		}
		if exists {
			continue
		}
		if i > 0 {
			x.MustExec(`
INSERT INTO work_log ( work, work_index, parent_record_id) 
VALUES ($1, $2, (
  SELECT record_id 
  FROM last_work_log 
  WHERE work_index = $3) );`,
				work.Name, work.Index, works[i-1].Index)
		} else {
			x.MustExec(`INSERT INTO work_log ( works, work_index) VALUES ($1, $2);`,
				work.Name, work.Index)
		}

	}
}

func LastWork(x *sqlx.DB) (s string) {
	dbMustGet(x, &s, `SELECT work FROM last_work_log WHERE parent_record_id ISNULL LIMIT 1;`)
	return
}

func AddRootWork(x *sqlx.DB, work string){
	x.MustExec(`INSERT INTO work_log ( work, work_index ) VALUES ($1, 0 );`, work )
}

func Write(x *sqlx.DB, w WriteRecord) (m CurrentWorkMessage) {

	EnsureCurrentWorks(x, w.Works)

	var productSerial *int
	if w.ProductSerial > 0 {
		productSerial = &w.ProductSerial
	}
	work := w.Works[len(w.Works)-1]

	r := x.MustExec(`
INSERT INTO work_log
  (parent_record_id,  product_serial, level, message)  values
  ( (SELECT record_id FROM last_work_log WHERE work_index = $1 LIMIT 1), $2, $3 , $4);`,
		work.Index, productSerial, w.Level, w.Text)

	rowID, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	err = x.Get(&m, `
SELECT 
  a.message AS message, 
  (CASE WHEN a.product_serial IS NULL THEN 0 ELSE a.product_serial END) AS product_serial, 
  a.level AS level, 
  a.created_at AS created_at, 
  b.work_index AS work_index
FROM last_work_log a
INNER JOIN last_work_log b ON a.parent_record_id = b.record_id
WHERE a.record_id = $1`, rowID)
	if err != nil {
		panic(err)
	}
	return

}

func GetMessagesOfLastWork(x *sqlx.DB, workIndex int) (xs []CurrentWorkMessage) {

	err := x.Select(&xs, `
WITH RECURSIVE acc(record_id, parent_record_id) AS (
  SELECT
    record_id, parent_record_id
  FROM last_work_log WHERE last_work_log.work_index = $1
  UNION
  SELECT
    w.record_id, w.parent_record_id
  FROM acc
    INNER JOIN last_work_log w ON w.parent_record_id = acc.record_id
)
SELECT 
  l.created_at as created_at, 
  p.work_index as work_index, 
  l.level as level, 
  l.message as message 
FROM acc
  INNER JOIN last_work_log p ON acc.parent_record_id = p.record_id
  INNER JOIN last_work_log l ON acc.record_id = l.record_id
WHERE l.message NOT NULL AND l.level NOT NULL;`, workIndex)

	if err != nil {
		panic(err)
	}
	return
}

type WorkInfo struct{
	HasError bool `db:"has_error"`
	HasMessage bool `db:"has_message"`
}

func GetLastWorkInfo(x *sqlx.DB, workIndex int) (workInfo WorkInfo ) {

	err := x.Get(&workInfo, `
WITH RECURSIVE acc(record_id, parent_record_id, level) AS (
  SELECT
    record_id, parent_record_id, level
  FROM last_work_log WHERE last_work_log.work_index = $1
  UNION
  SELECT
    w.record_id as record_id, 
    w.parent_record_id as parent_record_id, 
    w.level as level
  FROM acc
    INNER JOIN last_work_log w ON w.parent_record_id = acc.record_id
)

SELECT 
  EXISTS( SELECT * FROM acc WHERE level >= 4) as has_error, 
  EXISTS( SELECT * FROM acc ) as has_message;`, workIndex)

	if err != nil {
		panic(err)
	}

	return
}

func dbMustGet(db *sqlx.DB,  dest interface{}, query string, args ...interface{} ) {
	if err := db.Get(dest, query, args...); err != nil {
		panic(err)
	}
}

func dbMustSelect(db *sqlx.DB, dest interface{}, query string, args ...interface{}){
	if err := db.Select(dest, query); err != nil {
		panic(err)
	}
}

