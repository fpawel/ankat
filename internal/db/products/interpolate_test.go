package products

import (
	a "github.com/fpawel/ankat/internal/ankat"
	"github.com/fpawel/numeth"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func TestDBProducts_InterpolateLin(t *testing.T) {
	x := MustOpen(":memory:")
	x.testCreateParty()
	x.testAddInterploateData()

	testInterpolate(t, "LIN", a.Chan1, x.InterpolateLin,
		[]float64{-73.715561, 5.328987, -0.10691, 0.000715})

	testInterpolate(t, "LIN", a.Chan2, x.InterpolateLin,
		[]float64{-27.59523, 1.734398, -0.003078})

	testInterpolate(t, "T0", a.Chan1, x.InterpolateT0,
		[]float64{-10.070443, -0.178915, -0.007194})

	testInterpolate(t, "TK", a.Chan1, x.InterpolateTK,
		[]float64{0.998889, -0.001814, 0.000089 })

	testInterpolate(t, "T0", a.Chan2, x.InterpolateT0,
		[]float64{-129.404838, -0.169998, 0.047562 })

	testInterpolate(t, "TK", a.Chan2, x.InterpolateTK,
		[]float64{ 1.550823, -0.003824, -0.001067 })


}

type interpolateFunc = func(a.ProductSerial, a.AnkatChan) ([]float64, []numeth.Coordinate, error)

func testInterpolate(t *testing.T, what string, chanel a.AnkatChan,
	interpolateFunc interpolateFunc,
	mustBe []float64) {
	xs, _, err := interpolateFunc(1, chanel)
	if err != nil {
		t.Errorf("%s: %s: %v", what, chanel.What(), err)
	}
	mustEq(t, xs, mustBe)
}

func (x DBProducts) testAddInterploateData() {

	for sect, v := range map[a.Sect]map[a.Var][]float64{
		a.Lin1: {
			a.CoutCh1: []float64{23, 34, 55, 69},
		},
		a.Lin2: {
			a.CoutCh2: []float64{17, 49, 87},
		},
		a.T01: {
			a.TppCh1:  []float64{-33, 21, 50},
			a.Var2Ch1: []float64{12, 17, 37},
		},
		a.TK1: {
			a.TppCh1:  []float64{-33, 21, 50},
			a.Var2Ch1: []float64{57, 69, 83},
		},

		a.T02: {
			a.TppCh2:  []float64{-33, 21, 50},
			a.Var2Ch2: []float64{72, 112, 19},
		},
		a.TK2: {
			a.TppCh2:  []float64{-33, 21, 50},
			a.Var2Ch2: []float64{6, 78, 45},
		},
	} {
		for aVar, v := range v {
			for i, v := range v {
				x.SetCurrentPartyProductValue(1, a.ProductVar{
					Var:   aVar,
					Sect:  sect,
					Point: a.Point(i),
				}, v)
			}
		}
	}
}