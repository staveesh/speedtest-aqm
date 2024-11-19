package util

import (
	"fmt"
	"strings"
	"time"
)

func GetTime(args ...time.Time) float64 {
	t := time.Now()
	if len(args) > 0 {
		t = args[0]
	}

	return float64(t.UnixNano()) / 1000000000
}

func GetFilePath(dir string, fileName string, tstamp time.Time) string {
	parts := strings.SplitN(fileName, ".", 2)
	if len(parts) < 2 {
		parts = append(parts, "")
	}

	return fmt.Sprintf("%s/%s-%s.%s", dir, parts[0], tstamp.Format(time.RFC3339), parts[1])
}
