package products

type CurrentProduct struct {
	Product
	Checked       bool                `db:"checked"`
	Comport       string              `db:"comport"`
	Ordinal       int                 `db:"ordinal"`
}