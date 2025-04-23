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

type OoklaHttpStartInfo struct {
	Server struct {
		IP net.IP `json:"host"`
	} `json:"server"`
}

type OoklaHttpResultInfo struct {
	Download      float64 `json:"download"`
	Upload        float64 `json:"upload"`
	Ping          float64 `json:"ping"`
	Timestamp     string  `json:"timestamp"`
	BytesSent     int64   `json:"bytes_sent"`
	BytesReceived int64   `json:"bytes_received"`

	Server struct {
		Name    string  `json:"name"`
		Sponsor string  `json:"sponsor"`
		ID      string	`json:"id"`
		Host    string  `json:"host"`
	} `json:"server"`

}

func logParserOoklaHttp(logPipe *io.PipeReader) {
	defer close(logParserDone)

	var (
		ooklaStartInfo  OoklaHttpStartInfo
		ooklaResultInfo OoklaHttpResultInfo
		line            string
	)

	scanner := bufio.NewScanner(logPipe)
	for scanner.Scan() {
    	line = scanner.Text()
	    if strings.Contains(line, "download") {
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

	if err := json.Unmarshal([]byte(line), &ooklaResultInfo); err == nil {
		meta.MOoklaHttp = meta.MeasureOoklaHttp{
			Download: float64(ooklaResultInfo.Download) * 8 / 1000000,
			Upload:   float64(ooklaResultInfo.Upload) * 8 / 1000000,

			Latency:  ooklaResultInfo.Ping,

			ServerHost: ooklaResultInfo.Server.Host,
			ServerId:   ooklaResultInfo.Server.ID,
			ServerName: ooklaResultInfo.Server.Name,
		}

		meta.MBytes = ooklaResultInfo.BytesReceived + ooklaResultInfo.BytesReceived
	} else {
		log.Println("[speedtest] [log parser] failed to parse result info")
	}
}
