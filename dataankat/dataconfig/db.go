package dataconfig

import (
	"github.com/fpawel/ankat/dataankat/dbutils"
	"github.com/fpawel/guartutils/comport"
	"github.com/jmoiron/sqlx"
	"time"
)

type DBConfig struct {
	DB *sqlx.DB
}

func MustOpen(fileName string) DBConfig {

	db := DBConfig{dbutils.MustOpen(fileName, "sqlite3", SQLConfig) }

	for _,s := range []string{
		"comport_products", "comport_gas", "comport_temperature",
	} {
		_,err := db.DB.NamedExec(SQLComport, map[string]interface{}{"section_name": s} )
		if err != nil {
			panic(err)
		}
	}

	return db
}

func (x DBConfig) ConfigValue( dest interface{}, sectionName, propertyName string ) {
	dbutils.MustGet( x.DB, dest, `SELECT value FROM config WHERE section_name = ? AND property_name = ?;`,
		sectionName, propertyName, )
}

func (x DBConfig) ConfigComport(section string) (c comport.Config) {
	c.Serial.ReadTimeout = time.Millisecond
	c.Serial.Name = x.ConfigString( section, "port")
	c.Serial.Baud = x.ConfigInt(section, "baud")
	c.Fetch.ReadTimeout = x.ConfigMillisecond(section, "timeout")
	c.Fetch.ReadByteTimeout = x.ConfigMillisecond(section, "byte_timeout")
	c.Fetch.MaxAttemptsRead = x.ConfigInt(section, "repeat_count")
	c.BounceTimeout = x.ConfigMillisecond(section, "bounce_timeout")
	return
}

func (x DBConfig) ConfigHour(sectionName, propertyName string) time.Duration {
	return x.ConfigDuration(sectionName, propertyName, time.Hour )
}

func (x DBConfig) ConfigMinute(sectionName, propertyName string) time.Duration {
	return x.ConfigDuration(sectionName, propertyName, time.Minute )
}

func (x DBConfig) ConfigMillisecond(sectionName, propertyName string) time.Duration {
	return x.ConfigDuration(sectionName, propertyName, time.Millisecond )
}

func (x DBConfig) ConfigDuration(sectionName, propertyName string, k time.Duration) time.Duration {
	var v time.Duration
	x.ConfigValue(&v, sectionName, propertyName)
	v *= k
	return v
}

func (x DBConfig) ConfigFloat64(sectionName, propertyName string) float64 {
	var v float64
	x.ConfigValue(&v, sectionName, propertyName)
	return v
}

func (x DBConfig) ConfigString(sectionName, propertyName string) string {
	var v string
	x.ConfigValue(&v, sectionName, propertyName)
	return v
}

func (x DBConfig) ConfigInt(sectionName, propertyName string) int {
	var v int
	x.ConfigValue(&v, sectionName, propertyName)
	return v
}