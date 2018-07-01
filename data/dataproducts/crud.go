package dataproducts

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"time"
	"database/sql"
)

type PartyID int64

type Party struct {
	PartyID     PartyID   `db:"party_id"`
	CreatedAt   time.Time `db:"created_at"`
	ProductType string    `db:"product_type"`
	Products    []int     `db:"-"`
}

//type YearMonth struct {
//	Year, Month int
//}
//
//type YearMonthDay struct {
//	Year, Month, Day int
//}
//
//func ProductTypes(x *sqlx.DB) (productTypes []string) {
//	err := x.Select(&productTypes, `SELECT product_type FROM product_types;`)
//	if err != nil {
//		panic(err)
//	}
//	return
//}
//
//func CurrentParty(x *sqlx.DB) (party Party) {
//
//	err := x.Get(&party, `SELECT * FROM current_party;`)
//	if err != nil {
//		panic(err)
//	}
//	party.Products = GetProducts(x, party.PartyID)
//	return
//}

//func GetParty(x *sqlx.DB, partyID PartyID) (party Party) {
//	err := x.Get(&party, `
//SELECT party_id, created_at, product_type
//FROM parties
//WHERE party_id = $1;`, partyID)
//	if err != nil {
//		panic(err)
//	}
//	party.Products = GetProducts(x, partyID)
//	return
//}

//func GetProducts(x *sqlx.DB, partyID PartyID) (products []int) {
//	err := x.Select(&products, `
//SELECT product_serial
//FROM product
//WHERE party_id = $1
//ORDER BY product_serial ASC;`, partyID)
//	if err != nil {
//		panic(err)
//	}
//	return
//}

//func GetYears(x *sqlx.DB) (xs []int) {
//	err := x.Select(&xs, `
//SELECT cast(strftime('%Y', created_at) AS INT) AS year FROM party GROUP BY year;`)
//	if err != nil {
//		panic(err)
//	}
//	return
//}

//func GetMonthsOfYear(x *sqlx.DB, year int) (xs []int) {
//	err := x.Select(&xs, `
//SELECT cast( strftime('%m', created_at) AS INT) AS month FROM party
//WHERE cast(strftime('%Y', created_at) AS INT) = $1
//GROUP BY month;
//`, year)
//	if err != nil {
//		panic(err)
//	}
//	return
//}

//func GetDaysOfMonth(x *sqlx.DB, v YearMonth) (xs []int) {
//	err := x.Select(&xs, `
//SELECT cast( strftime('%d', created_at) AS INT) AS day FROM party
//WHERE  cast(strftime('%Y', created_at) AS INT) = $1 AND cast(strftime('%m', created_at) AS INT) = $2
//GROUP BY day;
//`, v.Year, v.Month)
//	if err != nil {
//		panic(err)
//	}
//	return
//}

//func GetPartiesOfDay(x *sqlx.DB, v YearMonthDay) (xs []Party) {
//	err := x.Select(&xs, `
//SELECT party_id, created_at, product_type FROM party
//WHERE
//  cast(strftime('%Y', created_at) AS INT) = $1 AND
//  cast(strftime('%m', created_at) AS INT) = $2 AND
//  cast(strftime('%d', created_at) AS INT) = $3
//ORDER BY created_at;
//`, v.Year, v.Month, v.Day)
//	if err != nil {
//		panic(err)
//	}
//	return
//}

type KeysValues map[interface{}]interface{}

type NewPartyConfig struct {
	ProductsCount int
	ProductType   string
	KeysValues    KeysValues
}

//func CreateNewParty(x *sqlx.DB, c NewPartyConfig) (result Party) {
//
//	t := time.Now()
//
//	sqlInsertParty := fmt.Sprintf(`INSERT INTO parties (product_type) VALUES ('%s');`, c.ProductType)
//
//	sqlInsertProducts := `INSERT INTO products (party_id, serial) VALUES `
//
//	for i := 0; i < c.ProductsCount; i++ {
//		sqlInsertProducts += fmt.Sprintf(
//			"((SELECT * FROM current_party_id), %d)",
//			i+1)
//		if i < c.ProductsCount-1 {
//			sqlInsertProducts += ", "
//		} else {
//			sqlInsertProducts += ";\n"
//		}
//	}
//	sql := "BEGIN TRANSACTION;" + sqlInsertParty + "\n" + sqlInsertProducts + "COMMIT;"
//	x.MustExec(sql)
//
//	for k, v := range c.KeysValues {
//		SetPartyValue(x, KeyValue{k, v})
//	}
//
//	result = CurrentParty(x)
//	fmt.Println("NEW PARTY:", time.Since(t))
//	return
//}

//func DeleteParty(x *sqlx.DB, partyID PartyID) {
//	x.MustExec(`DELETE FROM parties WHERE party_id=$1;`, partyID)
//}
//
//func PartyExists(x *sqlx.DB, partyID PartyID) (r bool) {
//	dbMustGet(x, &r, `SELECT exists( SELECT * FROM parties WHERE party_id=$1)`, partyID)
//	return
//}

type KeyValue struct {
	Key, Value interface{}
}

//func SetPartyValue(x *sqlx.DB, kv KeyValue) (err error) {
//	_, err = x.Exec(`
//INSERT OR REPLACE INTO party_value (party_id, param, value)
//VALUES ((SELECT * FROM current_party_id),?,?)`, kv.Key, kv.Key)
//	return
//}

func SetCoefficientValue(x *sqlx.DB,  productSerial,coefficient int, value float64) {
	x.MustExec(`
INSERT OR REPLACE INTO product_coefficient_value (party_id, product_serial, coefficient_id, value)
VALUES ((SELECT * FROM current_party_id),
        $1, $2, $3); `, productSerial, coefficient, value)
}

func CoefficientValue(x *sqlx.DB, productSerial,coefficient int) ( value sql.NullFloat64) {
	dbMustGet(x, &value,`
SELECT value FROM current_party_coefficient_value 
WHERE product_serial=$1 AND coefficient_id = $2;`, productSerial, coefficient,)
	return
}

func CurrentPartyValue(x *sqlx.DB, name string) ( value float64) {
	dbMustGet(x, &value, `
SELECT value FROM party_value 
WHERE var=$1 AND party_id IN ( SELECT * FROM current_party_id);`, name)
	return
}

func CurrentPartyValueStr(x *sqlx.DB, name string) ( value string) {
	dbMustGet(x, &value, `
SELECT value FROM party_value 
WHERE var=$1 AND party_id IN ( SELECT * FROM current_party_id);`, name)
	return
}

//func MustSetPartyValue(x *sqlx.DB, kv KeyValue) {
//	if err := SetPartyValue(x, kv); err != nil {
//		panic(err)
//	}
//}

func dbMustGet(db *sqlx.DB, dest interface{}, query string, args ...interface{}) {
	if err := db.Get(dest, query, args...); err != nil {
		panic(err)
	}
}

//func dbMustSelect(db *sqlx.DB, dest interface{}, query string, args ...interface{}) {
//	if err := db.Select(dest, query, args...); err != nil {
//		panic(err)
//	}
//}
