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

	var cmd *exec.Cmd
	var logParser func(*io.PipeReader)

	switch config.Tool {
	case "ookla":
		cmdArgs := []string{"--accept-license", "-f", "json", "-p", "yes"}
		if config.Server != "" {
			cmdArgs = append(cmdArgs, "--host", config.Server)
		}
		cmd = exec.Command("speedtest", cmdArgs...)
		logParser = logParserOokla
	case "ookla-http":
		var cmdArgs []string
		if config.Server == "" {
			log.Println("[ookla-http] No server specified, falling back to regular ookla.")
			cmdArgs = []string{"--accept-license", "-f", "json", "-p", "yes"}
			cmd = exec.Command("speedtest", cmdArgs...)
			logParser = logParserOokla
		} else {
			cmdArgs = []string{"--json", "--server-ip", config.Server}
			cmd = exec.Command("tools/ookla-http/speedtest.py", cmdArgs...)
			logParser = logParserOoklaHttp
		}
	case "iperf":
		cmdArgs := []string{"-c", config.Server, "-J"}
		if config.Server == "" {
			log.Println("[iperf] No server specified, using default iperf3 public servers may be required.")
			cmdArgs = []string{"-J"}
		}
		cmd = exec.Command("iperf3", cmdArgs...)
		logParser = logParserIperf
	default:
		cmdArgs := []string{"-format", "json"}
		if config.Server != "" {
			cmdArgs = append(cmdArgs, "-no-verify", "-server", config.Server)
		}
		cmd = exec.Command("ndt7-client", cmdArgs...)
		logParser = logParserNdt7
	}

	cmd.Stdout = logOut

	meta.MMeta.SpeedtestStartTime = timeUtil.UnixNow()
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

	meta.MMeta.SpeedtestEndTime = timeUtil.UnixNow()
	log.Println("[speedtest] complete")

	<-logParserDone
	log.Println("[speedtest] [log parser] complete")
}
