package ankat

import "fmt"

type AnkatChan int

const (
	Chan1 AnkatChan = 1
	Chan2 AnkatChan = 2
)

func MustValidChan(c AnkatChan) {
	if c != Chan1 && c != Chan2 {
		panic(fmt.Sprintf("канал концентрации должен быть 1 или 2: %d", c))

	}
}

func (x AnkatChan) What() string {
	MustValidChan(x)
	return fmt.Sprintf("Канал %d", x)
}

func (x AnkatChan) Lin() Sect {
	MustValidChan(x)
	if x == Chan1 {
		return Lin1
	} else {
		return Lin2
	}
}

func (x AnkatChan) LinPoints(isCO2 bool) (xs []LinPoint) {
	MustValidChan(x)
	if x == Chan1 {

		xs = []LinPoint{
			{
				ProductVar{Point: 0},
				GasNitrogen,
			},
			{
				ProductVar{Point: 1},
				GasChan1Middle1,
			},
		}
		if isCO2 {
			xs = append(xs, LinPoint{ProductVar{Point: 2}, GasChan1Middle2})
		}
		xs = append(xs, LinPoint{ProductVar{Point: 3}, GasChan1End})
	} else {
		xs = []LinPoint{
			{
				ProductVar{Point: 0},
				GasNitrogen,
			},
			{
				ProductVar{Point: 1},
				GasChan2Middle,
			},
			{
				ProductVar{Point: 2},
				GasChan2End,
			},
		}
	}

	for i := range xs {
		xs[i].Sect = x.Lin()
		xs[i].ProductVar.Var = x.CoutCh()
	}
	return
}

func (x AnkatChan) T0() Sect{
	MustValidChan(x)
	if x == Chan1 {
		return T01
	} else {
		return T02
	}
}

func (x AnkatChan) TK() Sect{
	MustValidChan(x)
	if x == Chan1 {
		return TK1
	} else {
		return TK2
	}
}

func (x AnkatChan) CoutCh() Var{
	MustValidChan(x)
	if x == Chan1 {
		return CoutCh1
	} else {
		return CoutCh2
	}
}

func (x AnkatChan) Var2() Var{
	MustValidChan(x)
	if x == Chan1 {
		return Var2Ch1
	} else {
		return Var2Ch2
	}
}

func (x AnkatChan) Tpp() Var{
	MustValidChan(x)
	if x == Chan1 {
		return TppCh1
	} else {
		return TppCh2
	}
}
