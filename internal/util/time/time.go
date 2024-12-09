package time

import (
	"time"
)

func GetTime() float64 {
	return UnixPrecise(time.Now())
}

func UnixPrecise(t time.Time) float64 {
	return float64(t.UnixNano()) / 1e9
}
