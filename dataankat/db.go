package dataankat

import (
	"github.com/fpawel/ankat"
	"github.com/fpawel/ankat/dataankat/dataconfig"
	"github.com/fpawel/ankat/dataankat/dataproducts"
)

type DBAnkat struct {
	dataconfig.DBConfig
	dataproducts.DBProducts
}


func MustOpen() DBAnkat {
	return DBAnkat{
		DBProducts:dataproducts.MustOpen(ankat.AppDataFileName( "products.db")),
		DBConfig:dataconfig.MustOpen(ankat.AppDataFileName("config.db")),

	}
}