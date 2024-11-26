package term

import (
	"fmt"
	"os"
	"regexp"
)

func Confirm(prompt string) {
	for {
		response := ""

		fmt.Printf("%s: continue? [yN] ", prompt)
		fmt.Scanln(&response)

		if confirmation, _ := regexp.MatchString(`^[yY]$`, response); confirmation {
			break
		}

		if rejection, _ := regexp.MatchString(`^[nN]?$`, response); rejection {
			os.Exit(1)
		}
	}
}

// IsTerm: whether stdout is attached to a character device (terminal)
func IsTerm() bool {
	fileInfo, _ := os.Stdout.Stat()
	fileMask := fileInfo.Mode() & os.ModeCharDevice
	return fileMask == os.ModeCharDevice
}
