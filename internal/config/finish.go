/*
 * finish: verify/clean/finish config data
 *
 * finishers defined in dedicated file
 *
 */
package config

import (
	"fmt"
	"log"

	"github.com/internet-equity/traceneck/internal/util/term"
)

// finish: invoke finishers and return any error
func finish() error {
	for _, finisher := range finishers {
		result := finisher()

		if result == nil {
			continue
		}

		if logM := result.Log(); logM != "" {
			log.Println(logM)
		}

		if notice := result.Notice(); notice != "" {
			fmt.Println(notice)
		}

		if result.Error() != "" {
			return result
		}

		if prompt := result.Prompt(); prompt != "" {
			confirmed := term.Confirm(prompt)

			if !confirmed {
				return result
			}
		}
	}

	return nil
}
