/*
 * finishers: slice of ConfigFinish-returning closures
 *
 * verify/clean/finish config data here
 *
 */
package config

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/google/gopacket/pcap"

	osUtil "github.com/internet-equity/traceneck/internal/util/os"
	"github.com/internet-equity/traceneck/internal/util/term"
)

// finishers: slice of configuration-checking closures
//
// closures return a ConfigFinish, such as a ConfigEval or Confirmation
//
// ConfigEval will either log a valid value or interrupt the process with an error, depending on
// whether an error message (ErrorM) is supplied.
//
// Confirmations will initiate a [yN] prompt, and only interrupt the process if the user rejects
// the prompt.
var finishers = [...]func() ConfigFinish{
	// Confirmation-returning closures
	//
	// OutPath: checkNakedOutPath: confirm ambiguous value missing both trailing slash and file extension
	func() ConfigFinish {
		if Force || OutPath == "-" || osUtil.PathDirectoryLike(OutPath) || filepath.Ext(OutPath) != "" || !term.IsTerm() {
			return nil
		}
		return Confirmation{
			Label: "archive destination is ambiguous",
			NoticeT: "%s without extension (e.g. \".tar\", \".tgz\", \".tar.gz\") " +
				"[or specify trailing slash to write outputs to directory]",
		}
	},

	// OutPath: checkTerminalOutput: if we're going to write to stdout, check that it's not the terminal
	func() ConfigFinish {
		if Force || OutPath != "-" || !term.IsTerm() {
			return nil
		}
		return Confirmation{Label: "archive destination is character device (terminal)"}
	},

	// ConfigEval-returning closures
	//
	// Interface: checkInterface: check Interface and set InterfaceIP
	func() ConfigFinish {
		iface, err := net.InterfaceByName(Interface)
		if err != nil {
			return ConfigEval{
				Label:  "interface",
				Value:  Interface,
				ErrorM: "not found",
			}
		}
		if iface.Flags&net.FlagLoopback != 0 {
			return ConfigEval{
				Label:  "interface",
				Value:  Interface,
				ErrorM: "loopback interface",
			}
		}
		if iface.Flags&net.FlagRunning == 0 {
			return ConfigEval{
				Label:  "interface",
				Value:  Interface,
				ErrorM: "not running",
			}
		}

		if handle, err := pcap.OpenLive(Interface, 0, false, 0); err == nil {
			handle.Close()
		} else {
			return ConfigEval{
				Label:  "interface",
				Value:  Interface,
				ErrorM: "requires capture permission",
			}
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return ConfigEval{
				Label:  "interface",
				Value:  Interface,
				ErrorM: "addresses not found: " + err.Error(),
			}
		}

		// all good: collect InterfaceIP
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				InterfaceIP = append(InterfaceIP, ipnet.IP)
			}
		}

		return ConfigEval{Label: "interface", Value: Interface}
	},

	// Tool: checkTool
	func() ConfigFinish {
		var cmd *exec.Cmd
		if Tool == "ndt" {
			cmd = exec.Command("ndt7-client", "--help")
		} else if Tool == "ookla" {
			cmd = exec.Command("speedtest", "--version")
		} else if Tool == "ookla-http" {
			cmd = exec.Command("tools/ookla-http/speedtest.py", "--version")
		} else {
			return ConfigEval{
				Label:  "tool",
				Value:  Tool,
				ErrorM: "invalid tool",
			}
		}

		if cmd.Run() != nil {
			return ConfigEval{
				Label:  "tool",
				Value:  Tool,
				ErrorM: "not installed",
			}
		}

		return ConfigEval{Label: "tool", Value: Tool}
	},

	// PingType: checkPingType
	func() ConfigFinish {
		if PingType != "icmp" && PingType != "udp" {
			return ConfigEval{
				Label:  "ping type",
				Value:  PingType,
				ErrorM: "invalid packet type",
			}
		}

		return ConfigEval{Label: "ping type", Value: PingType}
	},

	// MaxTTL: log only
	func() ConfigFinish {
		return ConfigEval{Label: "max ttl", Value: strconv.Itoa(MaxTTL)}
	},

	// DirectHop: checkDirectHop
	func() ConfigFinish {
		if DirectHop < 0 || DirectHop > MaxTTL {
			return ConfigEval{
				Label:  "direct hop",
				Value:  strconv.Itoa(DirectHop),
				ErrorM: fmt.Sprintf("not in range [0, %d]", MaxTTL),
			}
		}

		return ConfigEval{Label: "direct hop", Value: strconv.Itoa(DirectHop)}
	},

	// OutPath: checkOutPath: if not directory-like, nor stdout, ensure can open it for writing
	func() ConfigFinish {
		if OutPath != "-" && !osUtil.PathDirectoryLike(OutPath) {
			outDir := filepath.Dir(OutPath)

			if err := osUtil.DirAvail(outDir); err != nil {
				return ConfigEval{
					Label:  "output path",
					Value:  OutPath,
					ErrorM: err.Error(),
				}
			}

			outFile, err := os.Create(OutPath)
			if err != nil {
				return ConfigEval{
					Label:  "output path",
					Value:  OutPath,
					ErrorM: "could not open for writing: " + err.Error(),
				}
			}
			defer outFile.Close()
		}

		return ConfigEval{Label: "output path", Value: OutPath}
	},

	// WorkDir: checkWorkDir: establish dir path and ensure writeable
	func() ConfigFinish {
		if osUtil.PathDirectoryLike(OutPath) {
			// working directory can be output directory
			WorkDir = OutPath
		} else {
			// write to temporary directory before writing archive to path
			var err error
			TempWorkDir, err = os.MkdirTemp("", "traceneck-")
			if err != nil {
				return ConfigEval{
					Label:  "working dir",
					Value:  WorkDir,
					ErrorM: "could not create temporary directory",
				}
			}
			WorkDir = TempWorkDir
		}

		if err := osUtil.DirAvail(WorkDir); err != nil {
			return ConfigEval{
				Label:  "working dir",
				Value:  WorkDir,
				ErrorM: err.Error(),
			}
		}

		if err := osUtil.DirWriteable(WorkDir); err != nil {
			return ConfigEval{
				Label:  "working dir",
				Value:  WorkDir,
				ErrorM: "requires write access",
			}
		}

		return ConfigEval{Label: "working dir", Value: WorkDir}
	},

	// TShark: checkTshark
	func() ConfigFinish {
		if TShark && exec.Command("tshark", "--version").Run() != nil {
			return ConfigEval{
				Label:  "tshark",
				Value:  strconv.FormatBool(TShark),
				ErrorM: "not installed",
			}
		}

		return ConfigEval{Label: "tshark", Value: strconv.FormatBool(TShark)}
	},

	// IdleTime: log only
	func() ConfigFinish {
		return ConfigEval{Label: "idle time", Value: strconv.Itoa(IdleTime)}
	},
}
