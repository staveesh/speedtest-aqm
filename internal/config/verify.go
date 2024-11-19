package config

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/google/gopacket/pcap"
)

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

func checkDataDir() error {
	if fd, err := os.Stat(OutDir); err != nil {
		if err := os.MkdirAll(OutDir, 0755); err != nil {
			return errors.New("could not create directory: " + err.Error())
		}
	} else if !fd.IsDir() {
		return errors.New("not a directory")
	}

	if tmpFile, err := os.CreateTemp(OutDir, ".write"); err != nil {
		return errors.New("requires write access")
	} else {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
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
	if err := checkInterface(); err != nil {
		log.Println("[config] interface:", Interface, err)
		os.Exit(1)
	} else {
		log.Println("[config] interface:", Interface)
	}

	if err := checkTool(); err != nil {
		log.Println("[config] tool:", Tool, err)
		os.Exit(1)
	} else {
		log.Println("[config] tool:", Tool)
	}

	if err := checkPingType(); err != nil {
		log.Println("[config] ping type:", PingType, err)
		os.Exit(1)
	} else {
		log.Println("[config] ping type:", PingType)
	}

	log.Println("[config] max ttl:", MaxTTL)

	if err := checkDirectHop(); err != nil {
		log.Println("[config] direct hop:", DirectHop, err)
		os.Exit(1)
	} else {
		log.Println("[config] direct hop:", DirectHop)
	}

	if err := checkDataDir(); err != nil {
		log.Println("[config] output dir:", OutDir, err)
		os.Exit(1)
	} else {
		log.Println("[config] output dir:", OutDir)
	}

	if TShark {
		if err := checkTshark(); err != nil {
			log.Println("[config] tshark:", TShark, err)
			os.Exit(1)
		} else {
			log.Println("[config] tshark:", TShark)
		}
	}

	log.Printf("[config] idle time: %ds\n", IdleTime)
}
