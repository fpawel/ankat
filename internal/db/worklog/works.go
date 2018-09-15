package worklog

import (
	"github.com/fpawel/ankat/internal/ankat"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type CurrentWorkMessage = struct {
	WorkIndex     int       `db:"work_index"`
	WorkName      string    `db:"work_name"`
	CreatedAt     time.Time `db:"created_at"`
	ProductSerial int       `db:"product_serial"`
	Level         Level     `db:"level"`
	Text          string    `db:"message"`
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
	ProductSerial ankat.ProductSerial
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
		err := x.Get(&exists, `SELECT exists(SELECT * FROM last_work WHERE work_index=$1)`, work.Index)
		if err != nil {
			panic(err)
		}
		if exists {
			continue
		}
		if i > 0 {
			x.MustExec(`
INSERT INTO work ( work_name, work_index, parent_work_id) 
VALUES ($1, $2, (
  SELECT work_id 
  FROM last_work 
  WHERE work_index = $3) );`,
				work.Name, work.Index, works[i-1].Index)
		} else {
			x.MustExec(`INSERT INTO work ( work_name, work_index) VALUES ($1, $2);`,
				work.Name, work.Index)
		}

	}
}

func AddRootWork(x *sqlx.DB, work string) {
	x.MustExec(`INSERT INTO work ( work_name, work_index ) VALUES ($1, 0 );`, work)
}

func Write(x *sqlx.DB, w WriteRecord) (m CurrentWorkMessage) {

	EnsureCurrentWorks(x, w.Works)

	var productSerial *ankat.ProductSerial
	if w.ProductSerial > 0 {
		productSerial = &w.ProductSerial
	}
	work := w.Works[len(w.Works)-1]

	r := x.MustExec(`
INSERT INTO work_log
  (work_id,  product_serial, level, message)  VALUES
  ( (SELECT work_id FROM last_work WHERE work_index = $1 LIMIT 1), $2, $3 , $4);`,
		work.Index, productSerial, w.Level, w.Text)

	rowID, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}

	dbMustGet(x, &m, `
SELECT 
  message, 
  (CASE WHEN a.product_serial IS NULL THEN 0 ELSE a.product_serial END) AS product_serial, 
  level, 
  a.created_at AS created_at, 
  work_name,
  work_index
FROM work_log a
INNER JOIN work b ON a.work_id = b.work_id
WHERE record_id =  ?`, rowID)
	if err != nil {
		panic(err)
	}
	return

}

type WorkInfo struct {
	HasError   bool `db:"has_error"`
	HasMessage bool `db:"has_message"`
	Found      bool `db:"found"`
}

func GetLastWorkInfo(x *sqlx.DB, workIndex int, workName string) (workInfo WorkInfo) {

	err := x.Get(&workInfo, `
WITH RECURSIVE a(work_id, parent_work_id) AS
    (
    SELECT work_id, parent_work_id
    FROM last_work b
    WHERE b.work_index = $1 AND b.work_name = $2
        UNION
        SELECT w.work_id as record_id, w.parent_work_id as parent_record_id
        FROM a INNER JOIN last_work w ON w.parent_work_id = a.work_id
    ),
    c AS (SELECT * FROM a INNER JOIN work_log b ON a.work_id = b.work_id)
SELECT
  EXISTS( SELECT * FROM a ) AS found,
  EXISTS( SELECT * FROM c WHERE c.level >= 4 ) AS has_error,
  EXISTS( SELECT * FROM c ) AS has_message;
`, workIndex, workName)
	if err != nil {
		panic(err)
	}
	return
}

func dbMustGet(db *sqlx.DB, dest interface{}, query string, args ...interface{}) {
	if err := db.Get(dest, query, args...); err != nil {
		panic(err)
	}
}
