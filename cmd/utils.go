package main

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
)



func fmtErr(err error) string {
	if err != nil && err != context.Canceled {
		return err.Error()
	}
	return ""
}

func mustUnmarshalJson(b []byte, v interface{}) {
	if err := json.Unmarshal(b, v); err != nil {
		panic(err.Error() + ": " + string(b))
	}
}

func mustParseInt64(b []byte) int64 {
	v,err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		panic(err.Error() + ": " + string(b))
	}
	return v
}

func float6(x float64) float64{
	return math.Round(x * 1000000.) / 1000000.
}
