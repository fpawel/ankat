package ankat

import "fmt"

type PartyID int64

type ProductSerial int

type Sect string

type Var int

type GasCode int

type Point int

type Coefficient int

type SectInfo struct {
	What string
	Coefficient0 Coefficient
}

type ProductVar struct {
	Sect  Sect
	Var   Var
	Point Point
}

type SectPoint struct{
	Sect
	Point
}


const (

	Lin1 Sect = "LIN1"
	Lin2 Sect = "LIN2"

	T01 Sect = "T01"
	T02 Sect = "T02"

	TK1 Sect = "TK1"
	TK2 Sect = "TK2"

	PT Sect = "PT"
)

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
	Var2Ch1 Var = 658
	Var3Ch1 Var = 660

	VdatP Var  =    18

	CoutCh2 Var = 2
	TppCh2  Var = 674
	UwCh2   Var = 680
	UrCh2   Var = 682
	WorkCh2 Var = 684
	RefCh2  Var = 686
	Var1Ch2 Var = 688
	Var2Ch2 Var = 690
	Var3Ch2 Var = 692
)

func (x Sect) Description() string {
	return sectInfo[x].What
}

func (x Sect) Coefficient0() Coefficient {
	return sectInfo[x].Coefficient0
}

func Sects() (xs []Sect) {
	for s := range sectInfo {
		xs = append(xs, s)
	}
	return
}

func (x GasCode) Description() string {
	if s, ok := gases[x]; ok {
		return s
	}
	panic(fmt.Sprintf("unknown gas code: %d", x))
}

func (x Sect) PointDescription(point Point) string {
	if s, ok := pointSectDescription[SectPoint{x, point}]; ok {
		return s
	}
	return fmt.Sprintf("№%d", point+1)
}




var sectInfo = map[Sect]SectInfo{
	Lin1: { "линеаризация к.1", 23},
	Lin2: {"линеаризация к.2", 33},
}

var gases = map[GasCode]string{
	GasNitrogen:     "ПГС1 азот",
	GasChan1Middle1: "ПГС2 середина к.1",
	GasChan1Middle2: "ПГС3 середина доп.CO2",
	GasChan1End:     "ПГС4 шкала к.1",
	GasChan2Middle:  "ПГС5 середина к.2",
	GasChan2End:     "ПГС6 шкала к.2",
}

var pointSectDescription = map[struct{
	Sect
	Point
}]string{
	{Lin1, 0}: gases[GasNitrogen],
	{Lin1, 1}: gases[GasChan1Middle1],
	{Lin1, 2}: gases[GasChan1Middle2],
	{Lin1, 3}: gases[GasChan1End],

	{Lin2, 0}: gases[GasNitrogen],
	{Lin2, 1}: gases[GasChan2Middle],
	{Lin2, 2}: gases[GasChan2End],
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

func (x GasCode) Var() string {
	return fmt.Sprintf("cgas%d", x)
}


