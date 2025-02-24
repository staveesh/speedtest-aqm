package config

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"

	timeUtil "github.com/internet-equity/traceneck/internal/util/time"
)

var (
	// config flags
	//
	// defaults *may* be specified here (and to be modifiable prior to invocation of Define)
	Interface string           // interface
	Tool      string           // ndt or ookla
	Server    string		   // address for the custom server
	NoPing    bool             // whether to skip pings
	PingType  string           // icmp or udp
	MaxTTL    int              // maximum TTL until which to send pings
	DirectHop int              // hop to ping directly by icmp echo
	OutPath   string = "data/" // out path/directory (may be directory/, file or -)
	TShark    bool             // use tshark
	IdleTime  int              // idle time in seconds
	Force     bool             // whether to confirm
	Quiet     bool             // silence logging
	Terse     bool             // terse rtt metadata

	// other flags
	help    bool
	version bool

	// internal config
	WorkDir     string
	TempWorkDir string

	Timestamp   time.Time
	InterfaceIP []net.IP
	ServerIP    net.IP

	NAME    string
	VERSION string
)

func Define() {
	pflag.StringVarP(&Interface, "interface", "I", defaultInterface(), "Interface")
	pflag.StringVarP(&Tool, "tool", "t", "ndt", "Speedtest tool to use: ndt, ookla or iperf")
	pflag.StringVarP(&Server, "server", "s", "", "IP address and port (<ip>:<port>) for custom server. Optional. If not provided, will use default server.")
	pflag.BoolVarP(&NoPing, "no-ping", "n", false, "Skip pings")
	pflag.StringVarP(&PingType, "ping-type", "p", "icmp", "Ping packet type: icmp or udp")
	pflag.IntVarP(&MaxTTL, "max-ttl", "m", 5, "Maximum TTL until which to send pings")
	pflag.IntVarP(&DirectHop, "direct-hop", "d", 1, "Hop to ping directly by icmp echo [0 to skip]")
	pflag.BoolVarP(&TShark, "tshark", "T", false, "Use TShark")
	pflag.IntVarP(&IdleTime, "idle", "i", 10, "Post speedtest idle time (in secs)")
	pflag.StringVarP(&OutPath, "out-path", "o", OutPath, "Output path [path with trailing slash for directory, file path for tar archive, \"-\" for stdout]")
	pflag.BoolVarP(&Terse, "terse-metadata", "r", false, "Terse rtt metadata")
	pflag.BoolVarP(&Quiet, "quiet", "q", false, "Minimize logging")
	pflag.BoolVarP(&Force, "yes", "y", false, "Do not prompt for confirmation")
	pflag.BoolVarP(&help, "help", "h", false, "Show this help")
	pflag.BoolVarP(&version, "version", "v", false, "Show version")

	pflag.CommandLine.SortFlags = false
}

func Parse() error {
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

	if Quiet {
		// Disable logs
		log.SetOutput(io.Discard)
	}

	err := finish()
	if err != nil {
		return err
	}

	Timestamp = time.Now()
	log.Printf("[config] timestamp: %.9f", timeUtil.UnixPrecise(Timestamp))

	return nil
}

func ShouldArchive() bool {
	return OutPath != WorkDir
}

func GetFilePath(fileName string) string {
	// don't add time to members of archive
	if ShouldArchive() {
		return filepath.Join(WorkDir, fileName)
	}

	// insert time into file name
	fileParts := strings.SplitN(fileName, ".", 2)

	fileBase := fileParts[0]
	fileTimed := fmt.Sprintf("%s-%s", fileBase, Timestamp.Format(time.RFC3339))

	var fileFinal string

	if len(fileParts) == 1 {
		fileFinal = fileTimed
	} else {
		fileExt := fileParts[1]
		fileFinal = fmt.Sprintf("%s.%s", fileTimed, fileExt)
	}

	return filepath.Join(WorkDir, fileFinal)
}

func Teardown() {
	if TempWorkDir != "" {
		os.RemoveAll(TempWorkDir)
	}
}
