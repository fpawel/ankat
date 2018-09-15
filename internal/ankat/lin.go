package ankat

import "fmt"

type LinPoint = struct {
	ProductVar
	GasCode
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
