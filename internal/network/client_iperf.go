package network

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"strings"

	"github.com/internet-equity/traceneck/internal/channel"
	"github.com/internet-equity/traceneck/internal/config"
	"github.com/internet-equity/traceneck/internal/meta"
)

type IperfResult struct {
	Start struct {
		Connected []struct {
			RemoteHost string `json:"remote_host"`
		} `json:"connected"`
	} `json:"start"`
	End struct {
		SumSent struct {
			BitsPerSecond float64 `json:"bits_per_second"`
		} `json:"sum_sent"`
		SumReceived struct {
			BitsPerSecond float64 `json:"bits_per_second"`
		} `json:"sum_received"`
	} `json:"end"`
}

func logParserIperf(logPipe *io.PipeReader) {
	defer close(logParserDone)

	var (
		iperfResult IperfResult
		line        string
	)

	scanner := bufio.NewScanner(logPipe)
	var jsonOutput strings.Builder

	for scanner.Scan() {
		line = scanner.Text()
		jsonOutput.WriteString(line)
	}

	if err := json.Unmarshal([]byte(jsonOutput.String()), &iperfResult); err == nil {
		if len(iperfResult.Start.Connected) > 0 {
			config.ServerIP = net.ParseIP(iperfResult.Start.Connected[0].RemoteHost)
			close(channel.IPGrabbed)
			log.Println("[iperf] [log parser] grabbed server ip:", config.ServerIP)
		} else {
			log.Println("[iperf] [log parser] failed to grab server ip")
		}

		meta.MIperf = meta.MeasureIperf{
			Download: iperfResult.End.SumReceived.BitsPerSecond,
			Upload:   iperfResult.End.SumSent.BitsPerSecond,
		}
	} else {
		log.Println("[iperf] [log parser] failed to parse iperf result")
	}
}