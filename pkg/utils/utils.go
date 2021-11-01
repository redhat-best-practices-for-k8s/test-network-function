package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ArgListToMap takes a list of strings of the form "key=value" and translate it into a map
// of the form {key: value}
func ArgListToMap(lst []string) map[string]string {
	retval := make(map[string]string)
	for _, arg := range lst {
		splitArgs := strings.Split(arg, "=")
		if len(splitArgs) == 1 {
			retval[splitArgs[0]] = ""
		} else {
			retval[splitArgs[0]] = splitArgs[1]
		}
	}
	return retval
}

// FilterArray takes a list and a predicate and returns a list of all elements for whom the predicate returns true
func FilterArray(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func CheckFileExists(path, name string) {
	fullPath, _ := filepath.Abs(path)
	if _, err := os.Stat(path); err == nil {
		log.Infof("Path to %s file found and valid: %s ", name, fullPath)
	} else if errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Path to %s file not found: %s , Exiting", name, fullPath)
	} else {
		log.Infof("Path to %s file not valid: %s , err=%s, exiting", name, fullPath, err)
	}
}
