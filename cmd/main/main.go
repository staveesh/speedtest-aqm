package main

import (
	"time"

	"github.com/internet-equity/traceneck/internal/archive"
	"github.com/internet-equity/traceneck/internal/channel"
	"github.com/internet-equity/traceneck/internal/config"
	"github.com/internet-equity/traceneck/internal/meta"
	"github.com/internet-equity/traceneck/internal/network"
	"github.com/internet-equity/traceneck/internal/ping"
)

func main() {
	// Define args
	config.Define()

	// Parse args
	config.Parse()

	// Init metadata
	meta.Init()

	// Start background packet capture
	go network.CaptureProcess()

	// Start speedtest client
	go network.SpeedtestProcess()

	// Start pings to server
	go ping.PingProcess()

	// Wait until speedtest is complete
	<-channel.SpeedtestDone

	// Wait for relaxed state data
	time.Sleep(time.Duration(config.IdleTime) * time.Second)

	// Stop all processes
	close(channel.Stop)
	<-channel.PingDone
	<-channel.CaptureDone

	// Collect metadata
	meta.Collect()

	// Write metadata
	meta.Write()

	// Create tar
	if config.Archive {
		archive.CreateArchive()
	}
}
