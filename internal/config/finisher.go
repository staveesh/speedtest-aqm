/*
 * finisher: interfaces for use by finisher functions
 *
 * any finisher function is expected to return a struct abiding by the ConfigFinish interface
 *
 */
package config

import "fmt"

// ConfigFinish: interface used by all finishers
type ConfigFinish interface {
	Log() string
	Notice() string
	Prompt() string
	Error() string
}

// ConfigEval
//
// Standard finishers evalute user-provided configuration,
// returning a standard log message for valid values
// or a value-dependent error message for invalid values
//
// The user is not prompted for confirmation; errors are returned immediately.
type ConfigEval struct {
	Label  string // config label
	Value  string // config value
	ErrorM string // config error message (if any)
}

func (err ConfigEval) Log() string {
	if err.ErrorM == "" {
		return fmt.Sprintf("[config] %s: %s", err.Label, err.Value)
	}
	return ""
}

func (err ConfigEval) Error() string {
	if err.ErrorM == "" {
		return ""
	}
	return fmt.Sprintf("[config] %s: %s: %s", err.Label, err.Value, err.ErrorM)
}

func (err ConfigEval) Notice() string {
	return ""
}

func (err ConfigEval) Prompt() string {
	return ""
}

// Confirmation
//
// Alternatively, finishers may prompt the user for confirmation of their configuration.
// This ConfigFinish is returned as an error only if the user rejects the confirmation.
//
// A notice is optionally printed (rather than logged), whereby verbose information may be printed
// once, above the confirmation dialog.
//
// The error message for a Confirmation is an empty string, as no additional information need be
// communicated to the user.
type Confirmation struct {
	Label   string // confirmation label
	NoticeT string // template of above to be printed once
}

// Notice to print
//
// interpolate Label into NoticeT to reduce repetition
func (err Confirmation) Notice() string {
	if err.NoticeT == "" {
		return ""
	}
	return fmt.Sprintf(err.NoticeT, err.Label)
}

// Label is treated as the Prompt
func (err Confirmation) Prompt() string {
	return err.Label
}

func (err Confirmation) Log() string {
	return ""
}

// No need for Error
func (err Confirmation) Error() string {
	return ""
}
