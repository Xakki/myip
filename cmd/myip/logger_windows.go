//go:build windows

package main

import (
	"errors"
	"io"
)

func getSyslogWriter() (io.Writer, error) {
	return nil, errors.New("syslog is not supported on windows")
}

func isSyslogSupported() bool {
	return false
}

func getSystemWriter() (io.Writer, error) {
	return nil, errors.New("system logger (syslog) is not supported on windows")
}
