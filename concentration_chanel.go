package ankat

import "fmt"

type ConcentrationChannel int

const (
	ConcentrationChannel1 ConcentrationChannel = 1
	ConcentrationChannel2 ConcentrationChannel = 2
)

func MustValidConcentrationChannel(c ConcentrationChannel) {
	if c != ConcentrationChannel1 && c != ConcentrationChannel2{
		panic(fmt.Sprintf("канал концентрации должен быть 1 или 2: %d", c))

	}
}

func (x ConcentrationChannel) What() string {
	MustValidConcentrationChannel(x)
	return fmt.Sprintf("Канал %d", x)
}

func (x ConcentrationChannel) Lin() Sect {
	MustValidConcentrationChannel(x)
	if x == ConcentrationChannel1 {
		return Lin1
	} else {
		return Lin2
	}
}

func (x ConcentrationChannel) LinPoints(isCO2 bool) (xs []LinPoint) {
	MustValidConcentrationChannel(x)
	if x == ConcentrationChannel1 {

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

func (x ConcentrationChannel) T0() Sect{
	MustValidConcentrationChannel(x)
	if x == ConcentrationChannel1  {
		return T01
	} else {
		return T02
	}
}

func (x ConcentrationChannel) TK() Sect{
	MustValidConcentrationChannel(x)
	if x == ConcentrationChannel1  {
		return TK1
	} else {
		return TK2
	}
}

func (x ConcentrationChannel) CoutCh() Var{
	MustValidConcentrationChannel(x)
	if x == ConcentrationChannel1  {
		return CoutCh1
	} else {
		return CoutCh2
	}
}

func (x ConcentrationChannel) Var2() Var{
	MustValidConcentrationChannel(x)
	if x == ConcentrationChannel1  {
		return Var2Ch1
	} else {
		return Var2Ch2
	}
}

func (x ConcentrationChannel) Tpp() Var{
	MustValidConcentrationChannel(x)
	if x == ConcentrationChannel1  {
		return TppCh1
	} else {
		return TppCh2
	}
}
