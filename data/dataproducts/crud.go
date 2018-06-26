package dataproducts

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type PartyID int64

type Party struct {
	PartyID     PartyID     `db:"party_id"`
	CreatedAt   time.Time   `db:"created_at"`
	ProductType string      `db:"product_type"`
	Products    []int `db:"-"`
}

type DB struct {
	*sqlx.DB
}

type YearMonth struct {
	Year, Month int
}

type YearMonthDay struct {
	Year, Month, Day int
}

func (x DB) ProductTypes() (productTypes []string) {
	err := x.Select(&productTypes, `SELECT product_type FROM product_types;`)
	if err != nil {
		panic(err)
	}
	return
}

func (x DB) CurrentParty() (party Party) {

	err := x.Get(&party, `SELECT * FROM current_party;`)
	if err != nil {
		panic(err)
	}
	party.Products = x.Products(party.PartyID)
	return
}

func (x DB) Party(partyID PartyID) (party Party) {
	err := x.Get(&party, `
SELECT party_id, created_at, product_type
FROM parties
WHERE party_id = $1;`, partyID)
	if err != nil {
		panic(err)
	}
	party.Products = x.Products(partyID)
	return
}

func (x DB) Products(partyID PartyID) ( products []int) {
	err := x.Select(&products, `
SELECT serial 
FROM products 
WHERE party_id = $1 
ORDER BY serial ASC;`, partyID)
	if err != nil {
		panic(err)
	}
	return
}

func (x DB) Years() (xs []int) {
	err := x.Select(&xs, `
SELECT cast(strftime('%Y', created_at) AS INT) AS year FROM parties GROUP BY year;`)
	if err != nil {
		panic(err)
	}
	return
}

func (x DB) MonthsOfYear(year int) (xs []int) {
	err := x.Select(&xs, `
SELECT cast( strftime('%m', created_at) AS INT) AS month FROM parties
WHERE cast(strftime('%Y', created_at) AS INT) = $1
GROUP BY month;
`, year)
	if err != nil {
		panic(err)
	}
	return
}

func (x DB) DaysOfMonth(v YearMonth) (xs []int) {
	err := x.Select(&xs, `
SELECT cast( strftime('%d', created_at) AS INT) AS day FROM parties
WHERE  cast(strftime('%Y', created_at) AS INT) = $1 AND cast(strftime('%m', created_at) AS INT) = $2
GROUP BY day;
`, v.Year, v.Month)
	if err != nil {
		panic(err)
	}
	return
}

func (x DB) PartiesOfDay(v YearMonthDay) (xs []Party) {
	err := x.Select(&xs, `
SELECT party_id, created_at, product_type FROM parties
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

func (x DB) GetPartyValue(partyID PartyID, param interface{}, value interface{}) error {
	return x.Get(value, `SELECT value FROM party_value WHERE party_id = ? AND param = ?;`, partyID, param)
}

type KeysValues map [interface{}] interface{}


type NewPartyConfig struct {
	ProductsCount int
	ProductType string
	KeysValues KeysValues

}

func (x DB) NewParty(c NewPartyConfig) (result Party) {

	t := time.Now()


	sqlInsertParty := fmt.Sprintf(`INSERT INTO parties (product_type) VALUES ('%s');`, c.ProductType )

	sqlInsertProducts := `INSERT INTO products (party_id, serial) VALUES `

	for i:=0; i < c.ProductsCount; i++ {
		sqlInsertProducts += fmt.Sprintf(
			"((SELECT * FROM current_party_id), %d)",
			i + 1)
		if i < c.ProductsCount-1 {
			sqlInsertProducts += ", "
		} else {
			sqlInsertProducts += ";\n"
		}
	}
	sql := "BEGIN TRANSACTION;" + sqlInsertParty + "\n" + sqlInsertProducts + "COMMIT;"
	x.MustExec(sql)

	for k,v := range c.KeysValues{
		x.SetPartyValue( KeyValue{k,v})
	}

	result = x.CurrentParty()
	fmt.Println("NEW PARTY:", time.Since(t))
	return
}

func (x DB) DeleteParty(partyID PartyID) {
	x.MustExec(`DELETE FROM parties WHERE party_id=$1;`, partyID)
}

func (x DB) PartyExists(partyID PartyID) (r bool) {
	dbMustGet(x.DB, &r, `SELECT exists( SELECT * FROM parties WHERE party_id=$1)`, partyID)
	return
}

type KeyValue struct {
	Key,Value interface{}
}

func (x DB) SetPartyValue(kv KeyValue) (err error ) {
	_,err = x.Exec( `
INSERT OR REPLACE INTO party_value (party_id, param, value) 
VALUES ((SELECT * FROM current_party_id),?,?)`, kv.Key, kv.Key)
	return
}

func (x DB) MustSetPartyValue(kv KeyValue)  {
	if err := x.SetPartyValue(kv); err != nil {
		panic(err)
	}
}

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

