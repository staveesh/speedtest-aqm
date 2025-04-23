package meta

import (
	"encoding/json"
	"log"
	"maps"
	"net"
	"os"
	"slices"

	"github.com/internet-equity/traceneck/internal/config"
	timeUtil "github.com/internet-equity/traceneck/internal/util/time"
)

type MeasureNdt struct {
	Download        float64 `json:"speedtest_ndt7_download"`
	DownloadLatency float64 `json:"speedtest_ndt7_downloadlatency"`
	DownloadRetrans float64 `json:"speedtest_ndt7_downloadretrans"`
	Server          string  `json:"speedtest_ndt7_server"`
	ServerIP        net.IP  `json:"speedtest_ndt7_server_ip"`
	Upload          float64 `json:"speedtest_ndt7_upload"`
}

type MeasureOokla struct {
	Download   float64 `json:"speedtest_ookla_download"`
	Jitter     float64 `json:"speedtest_ookla_jitter"`
	Latency    float64 `json:"speedtest_ookla_latency"`
	PktLoss2   float64 `json:"speedtest_ookla_pktloss2"`
	ServerHost string  `json:"speedtest_ookla_server_host"`
	ServerId   int     `json:"speedtest_ookla_server_id"`
	ServerName string  `json:"speedtest_ookla_server_name"`
	Upload     float64 `json:"speedtest_ookla_upload"`
}

type MeasureOoklaHttp struct {
	Download   float64 `json:"speedtest_ookla_download"`
	Latency    float64 `json:"speedtest_ookla_latency"`
	ServerHost string  `json:"speedtest_ookla_server_host"`
	ServerId   string  `json:"speedtest_ookla_server_id"`
	ServerName string  `json:"speedtest_ookla_server_name"`
	Upload     float64 `json:"speedtest_ookla_upload"`
}

type MeasureIperf struct {
	Download float64 `json:"speedtest_iperf_download"`
	Upload   float64 `json:"speedtest_iperf_upload"`
}

type RttSample struct {
	TTL         int     `json:"ttl"`
	Round       int     `json:"round"`
	ReplyIP     net.IP  `json:"reply_ip"`
	SendTime    float64 `json:"send_time"`
	RecvTime    float64 `json:"recv_time"`
	RTT         float64 `json:"rtt"`
	IcmpSeqNo   *int    `json:"icmp_seq_no,omitempty"`
	UdpDestPort *int    `json:"udp_dest_port,omitempty"`
}

type Measurements struct {
	Ndt7          *MeasureNdt   	`json:"ndt7,omitempty"`
	Ookla         *MeasureOokla 	`json:"ookla,omitempty"`
	OoklaHttp     *MeasureOoklaHttp `json:"ooklahttp,omitempty"`
	Iperf         *MeasureIperf 	`json:"iperf,omitempty"`
	RttSamples    []RttSample   	`json:"rtt_samples"`
	BytesConsumed int64         	`json:"test_bytes_consumed"`
}

type Meta struct {
	ID                 string   `json:"Id"`
	Time               float64  `json:"Time"`
	ToolStartTime      float64  `json:"Tool_start_time"`
	ToolEndTime        float64  `json:"Tool_end_time"`
	SpeedtestStartTime float64  `json:"Speedtest_start_time"`
	SpeedtestEndTime   float64  `json:"Speedtest_end_time"`
	PingStartTime      float64  `json:"Ping_start_time"`
	PingEndTime        float64  `json:"Ping_end_time"`
	Interface          string   `json:"Interface"`
	InterfaceIP        []net.IP `json:"Interface_ip"`
}

type Metadata struct {
	Measurements Measurements `json:"Measurements"`
	Meta         Meta         `json:"Meta"`
}

var (
	MNdt   MeasureNdt
	MOokla MeasureOokla
	MOoklaHttp MeasureOoklaHttp
	MIperf MeasureIperf

	MSamples       = make(map[int]RttSample)
	MBytes   int64 = 0

	MMeta Meta
	MetaD Metadata

	MetaFile string
)

func Init() {
	MMeta = Meta{
		Time:          timeUtil.UnixPrecise(config.Timestamp),
		ToolStartTime: timeUtil.UnixNow(),
		Interface:     config.Interface,
		InterfaceIP:   config.InterfaceIP,
	}

	log.Println("[metadata] init")
}

func Collect() {
	MMeta.ToolEndTime = timeUtil.UnixNow()

	MetaD = Metadata{
		Measurements: Measurements{
			BytesConsumed: MBytes,
		},
		Meta: MMeta,
	}

	if config.Tool == "ndt" {
		MetaD.Measurements.Ndt7 = &MNdt
	} else if config.Tool == "ookla" {	
		MetaD.Measurements.Ookla = &MOokla
	} else if config.Tool == "ookla-http" {	
		MetaD.Measurements.OoklaHttp = &MOoklaHttp
	}else if config.Tool == "iperf" {
		MetaD.Measurements.Iperf = &MIperf
	}

	if !config.Terse {
		MetaD.Measurements.RttSamples = slices.Collect(maps.Values(MSamples))
	}

	log.Println("[metadata] collected")
}

func Write() {
	MetaFile = config.GetFilePath("metadata.json")
	metaWriter, err := os.Create(MetaFile)
	if err != nil {
		log.Fatalln("[metadata] error opening metadata file:", err.Error())
	}

	if err := json.NewEncoder(metaWriter).Encode(MetaD); err != nil {
		log.Fatalln("[metadata] error writing metadata:", err.Error())
	} else {
		log.Println("[metadata] metadata written to:", MetaFile)
	}
}

func ToString() (string, error) {
	if metaBytes, err := json.Marshal(MetaD); err == nil {
		return string(metaBytes), nil
	} else {
		return "", err
	}
}
