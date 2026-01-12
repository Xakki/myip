//go:build !windows

package main

import (
	"io"
	"log/syslog"
)

func getSyslogWriter() (io.Writer, error) {
	return syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "myip")
}

func isSyslogSupported() bool {
	return true
}

func getSystemWriter() (io.Writer, error) {
	return syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "myip")
}
