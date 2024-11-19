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

type NdtConnectionInfo struct {
	Value struct {
		ConnectionInfo struct {
			Server string `json:"Server"`
		} `json:"ConnectionInfo"`
	} `json:"Value"`
}

type NdtDataInfo struct {
	Value struct {
		AppInfo struct {
			NumBytes int64 `json:"NumBytes"`
		} `json:"AppInfo"`
	} `json:"Value"`
}

type NdtFinalInfo struct {
	ServerFQDN string `json:"ServerFQDN"`
	ServerIP   net.IP `json:"ServerIP"`
	Download   struct {
		Throughput     Value `json:"Throughput"`
		Latency        Value `json:"Latency"`
		Retransmission Value `json:"Retransmission"`
	} `json:"Download"`
	Upload struct {
		Throughput Value `json:"Throughput"`
	} `json:"Upload"`
}

type Value struct {
	Value float64 `json:"Value"`
}

func logParserNdt7(logPipe *io.PipeReader) {
	defer close(logParserDone)

	var (
		ndtConnInfo                      NdtConnectionInfo
		ndtDownInfo, ndtUpInfo           NdtDataInfo
		ndtFinalInfo                     NdtFinalInfo
		line, last_download, last_upload string
	)

	scanner := bufio.NewScanner(logPipe)
	for scanner.Scan() {
		line = scanner.Text()

		if strings.Contains(line, "ConnectionInfo") {
			break
		}
	}

	if err := json.Unmarshal([]byte(line), &ndtConnInfo); err == nil {
		ipStr := strings.Split(ndtConnInfo.Value.ConnectionInfo.Server, ":")[0]
		config.ServerIP = net.ParseIP(ipStr)
		close(channel.IPGrabbed)

		log.Println("[speedtest] [log parser] grabbed server ip:", config.ServerIP)
	} else {
		log.Fatalf("[speedtest] [log parser] failed to grab server ip")
	}

	for scanner.Scan() {
		line = scanner.Text()

		if strings.Contains(line, "AppInfo") {
			if strings.Contains(line, "download") {
				last_download = line
			} else {
				last_upload = line
			}
		} else if strings.Contains(line, "ServerFQDN") {
			break
		}
	}

	if err := json.Unmarshal([]byte(line), &ndtFinalInfo); err == nil {
		meta.MNdt = meta.MeasureNdt{
			Download: ndtFinalInfo.Download.Throughput.Value,
			Upload:   ndtFinalInfo.Upload.Throughput.Value,

			DownloadLatency: ndtFinalInfo.Download.Latency.Value,
			DownloadRetrans: ndtFinalInfo.Download.Retransmission.Value,

			Server:   ndtFinalInfo.ServerFQDN,
			ServerIP: ndtFinalInfo.ServerIP,
		}
	} else {
		log.Println("[speedtest] [log parser] failed to parse final info")
	}

	if err := json.Unmarshal([]byte(last_download), &ndtDownInfo); err == nil {
		meta.MBytes += ndtDownInfo.Value.AppInfo.NumBytes
	} else {
		log.Println("[speedtest] [log parser] failed to parse download numBytes")
	}

	if err := json.Unmarshal([]byte(last_upload), &ndtUpInfo); err == nil {
		meta.MBytes += ndtUpInfo.Value.AppInfo.NumBytes
	} else {
		log.Println("[speedtest] [log parser] failed to parse upload numBytes")
	}
}
