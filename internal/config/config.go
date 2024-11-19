package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/internet-equity/traceneck/internal/util"
)

var (
	Interface string // interface
	Tool      string // ndt or ookla
	PingType  string // icmp or udp
	MaxTTL    int    // maximum TTL until which to send pings
	DirectHop int    // hop to ping directly by icmp echo
	OutDir    string // out directory
	TShark    bool   // use tshark
	IdleTime  int    // idle time in seconds
	Archive   bool   // create archive tar

	help    bool
	version bool

	Timestamp   time.Time
	InterfaceIP []net.IP
	ServerIP    net.IP

	NAME    string
	VERSION string
)

func Define() {
	pflag.StringVarP(&Interface, "interface", "I", defaultInterface(), "Interface")
	pflag.StringVarP(&Tool, "tool", "t", "ndt", "Speedtest tool to use: ndt or ookla")
	pflag.StringVarP(&PingType, "ping-type", "p", "icmp", "Ping packet type: icmp or udp")
	pflag.IntVarP(&MaxTTL, "max-ttl", "m", 5, "Maximum TTL until which to send pings")
	pflag.IntVarP(&DirectHop, "direct-hop", "d", 1, "Hop to ping directly by icmp echo [0 to skip]")
	pflag.StringVarP(&OutDir, "out-dir", "o", "data", "Output directory")
	pflag.BoolVarP(&TShark, "tshark", "T", false, "Use TShark")
	pflag.IntVarP(&IdleTime, "idle", "i", 10, "Post speedtest idle time (in secs)")
	pflag.BoolVarP(&Archive, "archive", "a", false, "Create output archive")
	pflag.BoolVarP(&help, "help", "h", false, "Show this help")
	pflag.BoolVarP(&version, "version", "v", false, "Show version")

	pflag.CommandLine.SortFlags = false
}

func Parse() {
	pflag.Parse()

	if help {
		fmt.Printf("\nUsage: %s [OPTIONS]\n\nOptions:\n", NAME)
		pflag.PrintDefaults()
		os.Exit(0)
	}

	if version {
		fmt.Printf("%s %s\n", NAME, VERSION)
		os.Exit(0)
	}

	verifyFlags()

	Timestamp = time.Now()
	log.Printf("[config] timestamp: %.9f", util.GetTime(Timestamp))
}
