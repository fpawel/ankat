package ankat

type Scale int

type GasComponent string

type ConcentrationUnits string

const (
	UnitsNkpr                   ConcentrationUnits = "%, НКПР"
	UnitsPercentVolumeFractions ConcentrationUnits = "объемная доля, %"

	GasCH4   GasComponent = "CH₄"
	GasC3H8  GasComponent = "C₃H₈"
	GasSumCH GasComponent = "∑CH"
	GasCO2   GasComponent = "CO₂"
)

func (x Scale) Code() float64 {
	switch x {
	case 2:
		return 2
	case 5:
		return 6
	case 10:
		return 7
	case 100:
		return 21
	}
	return 0
}

func (x ConcentrationUnits) Code() float64 {
	switch x {
	case "объемная доля, %":
		return 3
	case "%, НКПР":
		return 4
	}
	return 0
}

func (x GasComponent) Code(scale Scale) float64 {
	switch x {
	case "CH₄":
		return 16
	case "C₃H₈":
		return 14
	case "∑CH":
		return 15
	case "CO₂":
		switch scale {
		case 2:
			return 11
		case 5:
			return 12
		case 10:
			return 13
		}
	}
	return 0
}
