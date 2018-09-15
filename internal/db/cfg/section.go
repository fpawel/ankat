package cfg

import (
	"github.com/fpawel/ankat/internal/dbutils"
	"github.com/fpawel/guartutils/comport"
	"github.com/jmoiron/sqlx"
	"time"
)

type Section struct {
	DB *sqlx.DB
	Section string
}

func (x Section) GetValue( dest interface{}, propertyName string ) {
	dbutils.MustGet( x.DB, dest, `SELECT value FROM cfg WHERE section_name = ? AND property_name = ?;`,
		x.Section, propertyName, )
}

func (x Section) Comport() (c comport.Config) {
	c.Serial.ReadTimeout = time.Millisecond
	c.Serial.Name = x.String("port")
	c.Serial.Baud = x.Int("baud")
	c.Fetch.ReadTimeout = x.Millisecond("timeout")
	c.Fetch.ReadByteTimeout = x.Millisecond("byte_timeout")
	c.Fetch.MaxAttemptsRead = x.Int("repeat_count")
	c.BounceTimeout = x.Millisecond("bounce_timeout")
	return
}

func (x Section) Hour(propertyName string) time.Duration {
	return x.Duration(propertyName, time.Hour )
}

func (x Section) Minute(propertyName string) time.Duration {
	return x.Duration(propertyName, time.Minute )
}

func (x Section) Millisecond(propertyName string) time.Duration {
	return x.Duration(propertyName, time.Millisecond )
}

func (x Section) Duration(propertyName string, k time.Duration) time.Duration {
	var v time.Duration
	x.GetValue( &v, propertyName)
	v *= k
	return v
}

func (x Section) Float64(propertyName string) float64 {
	var v float64
	x.GetValue(&v, propertyName)
	return v
}

func (x Section) String(propertyName string) string {
	var v string
	x.GetValue(&v, propertyName)
	return v
}

func (x Section) Int( propertyName string) int {
	var v int
	x.GetValue(&v, propertyName)
	return v
}
