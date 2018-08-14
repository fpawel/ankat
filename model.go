package ankat

import "fmt"

type PartyID int64

type ProductSerial int

type Sect string

type Var int

type GasCode int

type Point int

type Coefficient int

type ProductVar struct {
	Sect  Sect
	Var   Var
	Point Point
}

const (
	GasClose GasCode = iota
	GasNitrogen
	GasChan1Middle1
	GasChan1Middle2
	GasChan1End
	GasChan2Middle
	GasChan2End
)

const (
	CoutCh1 Var = 0
	TppCh1  Var = 642
	UwCh1   Var = 648
	UrCh1   Var = 650
	WorkCh1 Var = 652
	RefCh1  Var = 654
	Var1Ch1 Var = 656
	Var3Ch1 Var = 660

	CoutCh2 Var = 2
	TppCh2  Var = 674
	UwCh2   Var = 680
	UrCh2   Var = 682
	WorkCh2 Var = 684
	RefCh2  Var = 686
	Var1Ch2 Var = 688
	Var3Ch2 Var = 692
)

const (
	Lin1 Sect = "LIN1"
	Lin2 Sect = "LIN2"
)

func SectDescription(s Sect) string {
	return varSects[s]
}

func Sects() (xs []Sect) {
	for s := range varSects {
		xs = append(xs, s)
	}
	return
}

func GasCodeDescription(gasCode GasCode) string {
	if s, ok := gases[gasCode]; ok {
		return s
	}
	panic(fmt.Sprintf("unknown gas code: %d", gasCode))
}

var varSects = map[Sect]string{
	Lin1: "линеаризация к.1",
	Lin2: "линеаризация к.2",
}

var gases = map[GasCode]string{
	GasNitrogen:     "ПГС1 азот",
	GasChan1Middle1: "ПГС2 середина к.1",
	GasChan1Middle2: "ПГС3 середина доп.CO2",
	GasChan1End:     "ПГС4 шкала к.1",
	GasChan2Middle:  "ПГС5 середина к.2",
	GasChan2End:     "ПГС6 шкала к.2",
}

func MainVars1() []Var {
	return []Var{
		CoutCh1,
		TppCh1,
		UwCh1,
		UrCh1,
		WorkCh1,
		RefCh1,
		Var1Ch1,
		Var3Ch1,
	}
}

func MainVars2() []Var {
	return []Var{
		CoutCh2,
		TppCh2,
		UwCh2,
		UrCh2,
		WorkCh2,
		RefCh2,
		Var1Ch2,
		Var3Ch2,
	}
}

type LinPoint = struct {
	ProductVar
	GasCode
}

func (x GasCode) Var() string {
	return fmt.Sprintf("cgas%d", x)
}

func Lin1Points(isCO2 bool) (xs []LinPoint) {
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
	for i := range xs {
		xs[i].Sect = Lin1
		xs[i].ProductVar.Var = CoutCh1
	}
	return
}

func Lin2Points() (xs []LinPoint) {
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
	for i := range xs {
		xs[i].Sect = Lin2
		xs[i].ProductVar.Var = CoutCh2
	}
	return
}

func LinProductVars(gas GasCode) []ProductVar {
	switch gas {
	case GasNitrogen:
		return []ProductVar{
			{
				Var:   CoutCh1,
				Sect:  Lin1,
				Point: 0,
			},
			{
				Var:   CoutCh2,
				Sect:  Lin2,
				Point: 0,
			},
		}
	case GasChan1Middle1:
		return []ProductVar{{
			Var:   CoutCh1,
			Sect:  Lin1,
			Point: 1,
		}}
	case GasChan1Middle2:
		return []ProductVar{{
			Var:   CoutCh1,
			Sect:  Lin1,
			Point: 2,
		}}
	case GasChan1End:
		return []ProductVar{{
			Var:   CoutCh1,
			Sect:  Lin1,
			Point: 3,
		}}
	case GasChan2Middle:
		return []ProductVar{{
			Var:   CoutCh2,
			Sect:  Lin2,
			Point: 1,
		}}
	case GasChan2End:
		return []ProductVar{{
			Var:   CoutCh2,
			Sect:  Lin2,
			Point: 2,
		}}
	default:
		panic(fmt.Sprintf("bad gas: %d", gas))

	}
}
