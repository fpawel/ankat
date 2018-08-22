package ankat

import "fmt"

type LinPoint = struct {
	ProductVar
	GasCode
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
