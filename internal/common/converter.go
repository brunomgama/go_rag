package common

import (
	"math"
	"strconv"
)

func Itoa(i int) string {
	return strconv.Itoa(i)
}

func AsInt(v any) int {
	switch t := v.(type) {
	case float64:
		return int(math.Round(t))
	case int:
		return t
	default:
		return 0
	}
}
