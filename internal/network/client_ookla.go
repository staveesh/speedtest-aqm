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

type OoklaStartInfo struct {
	Server struct {
		IP net.IP `json:"ip"`
	} `json:"server"`
}

type OoklaResultInfo struct {
	Ping struct {
		Jitter  float64 `json:"jitter"`
		Latency float64 `json:"latency"`
	} `json:"ping"`
	Download struct {
		Bandwidth int   `json:"bandwidth"`
		Bytes     int64 `json:"bytes"`
	} `json:"download"`
	Upload struct {
		Bandwidth int   `json:"bandwidth"`
		Bytes     int64 `json:"bytes"`
	} `json:"upload"`
	Server struct {
		ID   int    `json:"id"`
		Host string `json:"host"`
		Name string `json:"name"`
	} `json:"server"`
	PktLoss float64 `json:"packetLoss"`
}

func logParserOokla(logPipe *io.PipeReader) {
	defer close(logParserDone)

	var (
		ooklaStartInfo  OoklaStartInfo
		ooklaResultInfo OoklaResultInfo
		line            string
	)

	scanner := bufio.NewScanner(logPipe)
	for scanner.Scan() {
		line = scanner.Text()

		if strings.Contains(line, "testStart") {
			break
		}
	}

	if err := json.Unmarshal([]byte(line), &ooklaStartInfo); err == nil {
		config.ServerIP = ooklaStartInfo.Server.IP
		close(channel.IPGrabbed)

		log.Println("[speedtest] [log parser] grabbed server ip:", config.ServerIP)
	} else {
		log.Fatalf("[speedtest] [log parser] failed to grab server ip")
	}

	for scanner.Scan() {
		line = scanner.Text()

		if strings.Contains(line, "result") {
			break
		}
	}

	if err := json.Unmarshal([]byte(line), &ooklaResultInfo); err == nil {
		meta.MOokla = meta.MeasureOokla{
			Download: float64(ooklaResultInfo.Download.Bandwidth) * 8 / 1000000,
			Upload:   float64(ooklaResultInfo.Upload.Bandwidth) * 8 / 1000000,

			Jitter:   ooklaResultInfo.Ping.Jitter,
			Latency:  ooklaResultInfo.Ping.Latency,
			PktLoss2: ooklaResultInfo.PktLoss,

			ServerHost: ooklaResultInfo.Server.Host,
			ServerId:   ooklaResultInfo.Server.ID,
			ServerName: ooklaResultInfo.Server.Name,
		}

		meta.MBytes = ooklaResultInfo.Download.Bytes + ooklaResultInfo.Upload.Bytes
	} else {
		log.Println("[speedtest] [log parser] failed to parse result info")
	}
}
