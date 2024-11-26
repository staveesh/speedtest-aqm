package os

import (
	"errors"
	"os"
)

// dirAvail ensures directory dir exists
func DirAvail(dir string) error {
	if fd, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.New("could not create directory: " + err.Error())
		}
	} else if !fd.IsDir() {
		return errors.New("not a directory: " + dir)
	}
	return nil
}

// dirWriteable checks whether directory dir is writeable
//
// A temporary file is created under the directory path and
// removed immediately following this check.
func DirWriteable(dir string) error {
	file, err := os.CreateTemp(dir, ".traceneck-")
	if err == nil {
		defer file.Close()
		defer os.Remove(file.Name())
	}
	return err
}

// pathDirectoryLike: whether path attempts to specify a directory
//
// A directory is specified via a trailing path separator slash.
func PathDirectoryLike(path string) bool {
	pathLen := len(path)
	return pathLen > 0 && os.IsPathSeparator(path[pathLen-1])
}
