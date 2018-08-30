package view

import (
	"math"
	"strconv"
)

func formatFloat6(v float64) string {
	decimalPos := decimalPointPosition(v)
	if decimalPos < 7 {
		n := math.Pow(10, float64(6 - decimalPos) )
		v = math.Round(v * n ) / n
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func decimalPointPosition(x float64) ( pos int ) {
	x = math.Abs(x)
	for x >= 1 {
		x /= 10
		pos ++
	}
	return
}
