package ankat

import "fmt"

func TemperaturePointDescription(p Point) string {
	switch p {
	case 0:
		return "низкая температура"
	case 1:
		return "НКУ"
	case 2:
		return "высокая температура"
	default:
		panic(fmt.Sprintf("TemperaturePointDescription: %v", p))
	}
}
