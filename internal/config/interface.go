package config

import (
	"bufio"
	"os"
	"strings"
)

const (
	procNetRoutePath = "/proc/net/route"
	zeros            = "00000000"
)

func readNetRoute() <-chan []string {
	ch := make(chan []string)

	go func() {
		defer close(ch)

		f, err := os.Open(procNetRoutePath)
		if err != nil {
			return
		}
		defer f.Close()

		bs := bufio.NewScanner(f)
		for bs.Scan() {
			fields := strings.Fields(bs.Text())
			if len(fields) != 0 {
				ch <- fields
			}
		}
	}()

	return ch
}

func defaultInterface() string {
	for route := range readNetRoute() {
		if route[1] == zeros && route[7] == zeros {
			return route[0]
		}
	}

	return ""
}
