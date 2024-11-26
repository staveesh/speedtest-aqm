package config

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/gopacket/pcap"

	osUtil "github.com/internet-equity/traceneck/internal/util/os"
	"github.com/internet-equity/traceneck/internal/util/term"
)

// dedicated logger for failures -- which won't be disabled if we're in quiet mode
var flog = log.New(os.Stderr, "", log.LstdFlags)

func checkNakedOutPath() error {
	if Force || OutPath == "-" || osUtil.PathDirectoryLike(OutPath) || filepath.Ext(OutPath) != "" || !term.IsTerm() {
		return nil
	}
	core := errors.New("archive destination is ambiguous")
	return fmt.Errorf("%w without extension (e.g. \".tar\", \".tgz\", \".tar.gz\") "+
		"[or specify trailing slash to write outputs to directory]",
		core)
}

func checkTerminalOutput() error {
	// if we're going to write to stdout,
	// check that it's not the terminal
	if !Force && OutPath == "-" && term.IsTerm() {
		return errors.New("archive destination is character device (terminal)")
	}
	return nil
}

func checkInterface() error {
	iface, err := net.InterfaceByName(Interface)
	if err != nil {
		return errors.New("not found")
	}
	if iface.Flags&net.FlagLoopback != 0 {
		return errors.New("loopback interface")
	}
	if iface.Flags&net.FlagRunning == 0 {
		return errors.New("not running")
	}

	if handle, err := pcap.OpenLive(Interface, 0, false, 0); err == nil {
		handle.Close()
	} else {
		return errors.New("requires capture permission")
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return errors.New("addresses not found: " + err.Error())
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			InterfaceIP = append(InterfaceIP, ipnet.IP)
		}
	}

	return nil
}

func checkTool() error {
	var cmd *exec.Cmd
	if Tool == "ndt" {
		cmd = exec.Command("ndt7-client", "--help")
	} else if Tool == "ookla" {
		cmd = exec.Command("speedtest", "--version")
	} else {
		return errors.New("invalid tool")
	}

	if cmd.Run() != nil {
		return errors.New("not installed")
	}

	return nil
}

func checkPingType() error {
	if PingType != "icmp" && PingType != "udp" {
		return errors.New("invalid packet type")
	}

	return nil
}

func checkDirectHop() error {
	if DirectHop < 0 || DirectHop > MaxTTL {
		return fmt.Errorf("not in range [0, %d]", MaxTTL)
	}

	return nil
}

// OutPath: if not directory-like, not stdout, ensure can open it for writing
func checkOutPath() error {
	if OutPath != "-" && !osUtil.PathDirectoryLike(OutPath) {
		outDir := filepath.Dir(OutPath)

		if err := osUtil.DirAvail(outDir); err != nil {
			return err
		}

		outFile, err := os.Create(OutPath)
		if err != nil {
			return errors.New("could not open for writing: " + err.Error())
		}
		defer outFile.Close()
	}

	return nil
}

// WorkDir: establish dir path and ensure writeable
func checkWorkDir() error {
	if osUtil.PathDirectoryLike(OutPath) {
		// working directory can be output directory
		WorkDir = OutPath
	} else {
		// write to temporary directory before writing archive to path
		var err error
		WorkDir, err = os.MkdirTemp("", "traceneck-")
		if err != nil {
			return errors.New("could not create temporary directory")
		}
	}

	if err := osUtil.DirAvail(WorkDir); err != nil {
		return err
	}

	if err := osUtil.DirWriteable(WorkDir); err != nil {
		return errors.New("requires write access")
	}

	return nil
}

func checkTshark() error {
	if TShark && exec.Command("tshark", "--version").Run() != nil {
		return errors.New("not installed")
	}

	return nil
}

func verifyFlags() {
	if err := checkTerminalOutput(); err != nil {
		term.Confirm(err.Error())
	}

	if err := checkNakedOutPath(); err != nil {
		if core := errors.Unwrap(err); core != nil {
			fmt.Println(err.Error())
			err = core
		}
		term.Confirm(err.Error())
	}

	if err := checkInterface(); err != nil {
		// log and exit(1)
		flog.Fatalln("[config] interface:", Interface, err)
	} else {
		log.Println("[config] interface:", Interface)
	}

	if err := checkTool(); err != nil {
		// log and exit(1)
		flog.Fatalln("[config] tool:", Tool, err)
	} else {
		log.Println("[config] tool:", Tool)
	}

	if err := checkPingType(); err != nil {
		// log and exit(1)
		flog.Fatalln("[config] ping type:", PingType, err)
	} else {
		log.Println("[config] ping type:", PingType)
	}

	log.Println("[config] max ttl:", MaxTTL)

	if err := checkDirectHop(); err != nil {
		// log and exit(1)
		flog.Fatalln("[config] direct hop:", DirectHop, err)
	} else {
		log.Println("[config] direct hop:", DirectHop)
	}

	if err := checkOutPath(); err != nil {
		// log and exit(1)
		flog.Fatalln("[config] output path:", err)
	} else {
		log.Println("[config] output path:", OutPath)
	}

	if err := checkWorkDir(); err != nil {
		// log and exit(1)
		flog.Fatalln("[config] working dir:", err)
	} else {
		log.Println("[config] working dir:", WorkDir)
	}

	if TShark {
		if err := checkTshark(); err != nil {
			// log and exit(1)
			flog.Fatalln("[config] tshark:", TShark, err)
		} else {
			log.Println("[config] tshark:", TShark)
		}
	}

	log.Printf("[config] idle time: %ds\n", IdleTime)
}
