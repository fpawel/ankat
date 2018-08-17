package dataproducts

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func MustSetPartyValue(x *sqlx.DB, kv KeyStr) {
	if err := SetPartyValue(x, kv); err != nil {
		panic(err)
	}
}

func dbMustSelect(db *sqlx.DB, dest interface{}, query string, args ...interface{}) {
	if err := db.Select(dest, query, args...); err != nil {
		panic(err)
	}
}

func SetPartyValue(x *sqlx.DB, kv KeyStr) (err error) {
	_, err = x.Exec(`
INSERT OR REPLACE INTO party_value (party_id, param, value)
VALUES ((SELECT * FROM current_party_id),?,?)`, kv.Key, kv.Key)
	return
}

func CreateNewParty(x *sqlx.DB, c NewPartyConfig) (result Party) {

	t := time.Now()

	sqlInsertParty := fmt.Sprintf(`INSERT INTO parties (product_type) VALUES ('%s');`, c.ProductType)

	sqlInsertProducts := `INSERT INTO products (party_id, serial) VALUES `

	for i := 0; i < c.ProductsCount; i++ {
		sqlInsertProducts += fmt.Sprintf(
			"((SELECT * FROM current_party_id), %d)",
			i+1)
		if i < c.ProductsCount-1 {
			sqlInsertProducts += ", "
		} else {
			sqlInsertProducts += ";\n"
		}
	}
	sql := "BEGIN TRANSACTION;" + sqlInsertParty + "\n" + sqlInsertProducts + "COMMIT;"
	x.MustExec(sql)

	for k, v := range c.KeysValues {
		SetPartyValue(x, KeyStr{k, v})
	}

	result = CurrentParty(x)
	fmt.Println("NEW PARTY:", time.Since(t))
	return
}

func DeleteParty(x *sqlx.DB, partyID PartyID) {
	x.MustExec(`DELETE FROM parties WHERE party_id=$1;`, partyID)
}

func PartyExists(x *sqlx.DB, partyID PartyID) (r bool) {
	dbMustGet(x, &r, `SELECT exists( SELECT * FROM parties WHERE party_id=$1)`, partyID)
	return
}

type YearMonth struct {
	Year, Month int
}

type YearMonthDay struct {
	Year, Month, Day int
}

func ProductTypes(x *sqlx.DB) (productTypes []string) {
	err := x.Select(&productTypes, `SELECT product_type FROM product_types;`)
	if err != nil {
		panic(err)
	}
	return
}

func CurrentParty(x *sqlx.DB) (party Party) {

	err := x.Get(&party, `SELECT * FROM current_party;`)
	if err != nil {
		panic(err)
	}
	party.Products = GetProducts(x, party.PartyID)
	return
}



func GetProducts(x *sqlx.DB, partyID PartyID) (products []int) {
	err := x.Select(&products, `
SELECT product_serial
FROM product
WHERE party_id = $1
ORDER BY product_serial ASC;`, partyID)
	if err != nil {
		panic(err)
	}
	return
}

func GetYears(x *sqlx.DB) (xs []int) {
	err := x.Select(&xs, `
SELECT cast(strftime('%Y', created_at) AS INT) AS year FROM party GROUP BY year;`)
	if err != nil {
		panic(err)
	}
	return
}

func GetMonthsOfYear(x *sqlx.DB, year int) (xs []int) {
	err := x.Select(&xs, `
SELECT cast( strftime('%m', created_at) AS INT) AS month FROM party
WHERE cast(strftime('%Y', created_at) AS INT) = $1
GROUP BY month;
`, year)
	if err != nil {
		panic(err)
	}
	return
}

func GetDaysOfMonth(x *sqlx.DB, v YearMonth) (xs []int) {
	err := x.Select(&xs, `
SELECT cast( strftime('%d', created_at) AS INT) AS day FROM party
WHERE  cast(strftime('%Y', created_at) AS INT) = $1 AND cast(strftime('%m', created_at) AS INT) = $2
GROUP BY day;
`, v.Year, v.Month)
	if err != nil {
		panic(err)
	}
	return
}

func GetPartiesOfDay(x *sqlx.DB, v YearMonthDay) (xs []Party) {
	err := x.Select(&xs, `
SELECT party_id, created_at, product_type FROM party
WHERE
 cast(strftime('%Y', created_at) AS INT) = $1 AND
 cast(strftime('%m', created_at) AS INT) = $2 AND
 cast(strftime('%d', created_at) AS INT) = $3
ORDER BY created_at;
`, v.Year, v.Month, v.Day)
	if err != nil {
		panic(err)
	}
	return
}

func TestNewParty(t *testing.T) {
	filename := filepath.Join(goanalit.appFolder("tests").Path(), "products.db")

	db := MustOpen(filename, []string{"035", "035-1", "100-1"})

	p0 := db.CurrentParty()

	assert.True(t, db.PartyExists(p0.PartyID), "current party should exist")
	assert.Equal(t, p0, db.Party(p0.PartyID), "current parties should be equal")
	assert.Equal(t, p0, db.CurrentParty(), "current party should be equal")

	p1 := db.NewParty(NewPartyConfig{4, "035", KeysValues{"gas1": 123.456}})

	assert.True(t, db.PartyExists(p1.PartyID), "new party should exist")

	p2 := db.CurrentParty()
	assert.Equal(t, p1, p2, "new party should be current")
	assert.Equal(t, len(p1.Products), len(p2.Products), "lengths should be equal")
	for i := range p1.Products {
		assert.Equal(t, p1.Products[i], p2.Products[i], "products should be equal")
	}

	db.SetPartyValue("gas2", 789.012)

	var v float64
	err := db.GetPartyValue(p1.PartyID, "gas1", &v)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 123.456, v, "param value should be equal")

	err = db.GetPartyValue(p1.PartyID, "gas2", &v)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 789.012, v, "param value should be equal")

	db.DeleteParty(p2.PartyID)

	assert.True(t, !db.PartyExists(p2.PartyID), "deleted party should not exist")

	assert.Equal(t, p0, db.CurrentParty(), "parties should be equal")
}

func TestConcurrentRead(t *testing.T) {
	filename := filepath.Join(goanalit.appFolder("tests").Path(), "products.db")

	db := MustOpen(filename, []string{"035", "035-1", "100-1"})

	party0 := db.CurrentParty()
	ch := make(chan Party)
	for i := 0; i < 1000; i++ {
		go func() {
			ch <- db.CurrentParty()
		}()
	}

	n := 0
	for p := range ch {
		assert.Equal(t, party0, p, "parties should be equal")
		if n++; n == 1000 {
			break
		}
	}
}

func TestConcurrentWrite(t *testing.T) {
	filename := filepath.Join(goanalit.appFolder("tests").Path(), "products.db")
	db := MustOpen(filename, []string{"035", "035-1", "100-1"})
	var mu sync.Mutex
	done := make(chan time.Duration)
	const iterationCount = 100
	for i := 0; i < iterationCount; i++ {
		go func() {

			mu.Lock()
			start := time.Now()
			p := db.NewParty(NewPartyConfig{4, "035", KeysValues{"gas1": 123.456}})

			assert.True(t, db.PartyExists(p.PartyID), "party should exist")
			assert.Equal(t, p, db.Party(p.PartyID), "parties should be equal")
			assert.Equal(t, p, db.CurrentParty(), "current party should be equal")

			db.DeleteParty(p.PartyID)

			assert.True(t, !db.PartyExists(p.PartyID), "party should not exist")

			done <- time.Since(start)

			mu.Unlock()
		}()
	}

	i := 0
	for t := range done {
		i++
		fmt.Println("[", i, "]", t)
		if i == iterationCount {
			break
		}
	}
}

type KeysValues map[interface{}]interface{}

type NewPartyConfig struct {
	ProductsCount int
	ProductType   string
	KeysValues    KeysValues
}


func EnsurePartyExists(x *sqlx.DB) {
	var exists bool
	dbMustGet(x, &exists, `SELECT exists(SELECT party_id FROM party);`)
	if exists {
		return
	}
	x.MustExec(`
INSERT INTO party(party_id)  VALUES(1);
INSERT INTO product(party_id, product_serial) VALUES (1,1), (1,2), (1,3), (1,4), (1,5);`)

	var vars []string
	dbMustSelect(x, &vars, `SELECT var FROM party_var`)

	const (
		sqlDefVal = `SELECT def_val FROM party_var WHERE var = ?`
		sqlSet    = `INSERT INTO party_value (party_id, var, value) VALUES (1, ?, ?);`
	)
	for _, aVar := range vars {

		var strType string
		dbMustGet(x, &strType, `SELECT type FROM party_var WHERE var = ?`, aVar)
		switch strType {
		case "integer":
			var value int
			dbMustGet(x, &value, sqlDefVal, aVar)
			x.MustExec(sqlSet, aVar, value)
		case "text":
			var value string
			dbMustGet(x, &value, sqlDefVal, aVar)
			x.MustExec(sqlSet, aVar, value)
		case "real":
			var value float64
			dbMustGet(x, &value, sqlDefVal, aVar)
			x.MustExec(sqlSet, aVar, value)
		default:
			panic(strType)
		}
	}
}