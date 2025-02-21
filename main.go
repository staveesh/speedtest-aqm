package main

import (
	"log"
	"os"
	"time"

	"github.com/internet-equity/traceneck/internal/archive"
	"github.com/internet-equity/traceneck/internal/channel"
	"github.com/internet-equity/traceneck/internal/config"
	"github.com/internet-equity/traceneck/internal/meta"
	"github.com/internet-equity/traceneck/internal/network"
	"github.com/internet-equity/traceneck/internal/ping"
)

// flog: dedicated logger for failures -- which won't disable in quiet mode
var flog = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	// Define args
	config.Define()

	// Parse args
	err := config.Parse()

	// Ensure final teardown
	defer config.Teardown()

	// Handle parse error
	if err != nil {
		if errM := err.Error(); errM != "" {
			// log and exit(1)
			flog.Fatalln(errM)
		} else {
			// just exit(1)
			os.Exit(1)
		}
	}

	// Init metadata
	meta.Init()

	// Start background packet capture
	go network.CaptureProcess()

	// Start speedtest client
	go network.SpeedtestProcess()

	// Start pings to server if enabled
	if !config.NoPing {
		go ping.PingProcess()
	}

	// Wait until speedtest is complete
	<-channel.SpeedtestDone

	// Wait for relaxed state data
	time.Sleep(time.Duration(config.IdleTime) * time.Second)

	// Stop all processes
	close(channel.Stop)
	if !config.NoPing {
		<-channel.PingDone
	}
	<-channel.CaptureDone

	// Collect metadata
	meta.Collect()

	// Write metadata
	meta.Write()

	// Write archive
	if config.ShouldArchive() {
		archive.Write()
	}
}
