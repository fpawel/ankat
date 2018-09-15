package products

import (
	_ "github.com/mattn/go-sqlite3"
	"math"
	"testing"
)

func (x DBProducts) testCreateParty() {
	x.DB.MustExec(`
INSERT INTO party (sensors_count,
                   pressure_sensor,
                   cgas1,
                   cgas2,
                   cgas3,
                   cgas4,
                   cgas5,
                   cgas6,
                   gas1,
                   gas2,
                   scale1,
                   scale2,
                   units1,
                   units2)
VALUES (2, 1, 1, 12, 15, 20, 50, 100, 'CO₂', 'CH₄', 10, 100, 'объемная доля, %', 'объемная доля, %');
INSERT INTO product(party_id, product_serial) VALUES (1,1);
`)

}



func round6(x float64) float64 {
	return math.Round(x*1000000.) / 1000000.
}

func mustEq(t *testing.T, x, y []float64) {
	for i := range x {
		if round6(x[i]) != round6(y[i]) {
			t.Errorf("%v != %v", x, y)
			return
		}
	}
}
