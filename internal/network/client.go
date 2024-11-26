package network

import (
	"io"
	"log"
	"os/exec"

	"github.com/internet-equity/traceneck/internal/channel"
	"github.com/internet-equity/traceneck/internal/config"
	"github.com/internet-equity/traceneck/internal/meta"
	timeUtil "github.com/internet-equity/traceneck/internal/util/time"
)

var logParserDone = make(channel.Type)

// Wrapper functions for synchronization
func SpeedtestProcess() {
	defer close(channel.SpeedtestDone)

	// Create a pipe to capture stdout
	logIn, logOut := io.Pipe()
	defer logOut.Close()

	cmd := exec.Command("ndt7-client", "-format", "json")
	logParser := logParserNdt7
	if config.Tool == "ookla" {
		cmd = exec.Command(
			"speedtest",
			"--accept-license",
			"-f", "json",
			"-p", "yes",
		)
		logParser = logParserOokla
	}
	cmd.Stdout = logOut

	meta.MMeta.SpeedtestStartTime = timeUtil.GetTime()
	if err := cmd.Start(); err != nil {
		log.Println("[speedtest] client error:", err)
		return
	}
	log.Println("[speedtest] started")

	go logParser(logIn)
	log.Println("[speedtest] [log parser] started")

	if err := cmd.Wait(); err != nil {
		log.Println("[speedtest] client error:", err)
		return
	}
	meta.MMeta.SpeedtestEndTime = timeUtil.GetTime()
	log.Println("[speedtest] complete")

	<-logParserDone
	log.Println("[speedtest] [log parser] complete")
}
