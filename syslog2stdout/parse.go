// Package syslog2stdout handles listening for syslog messages and
// writing them to stdout.
package syslog2stdout

import (
	"bytes"
	"fmt"
)

var priorities = map[int]string{
	0: "emerg", // LOG_EMERG
	1: "alert", // LOG_ALERT
	2: "crit",  // LOG_CRIT
	3: "err",   // LOG_ERR
	4: "warn",  // LOG_WARNING
	5: "note",  // LOG_NOTICE
	6: "info",  // LOG_INFO
	7: "debug", // LOG_DEBUG
}

var facilities = map[int]string{
	(0 << 3):  "kern",     // LOG_KERN
	(1 << 3):  "user",     // LOG_USER
	(2 << 3):  "mail",     // LOG_MAIL
	(3 << 3):  "daemon",   // LOG_DAEMON
	(4 << 3):  "auth",     // LOG_AUTH
	(5 << 3):  "syslog",   // LOG_SYSLOG
	(6 << 3):  "lpr",      // LOG_LPR
	(7 << 3):  "news",     // LOG_NEWS
	(8 << 3):  "uucp",     // LOG_UUCP
	(9 << 3):  "cron",     // LOG_CRON
	(10 << 3): "authpriv", // LOG_AUTHPRIV
	(11 << 3): "ftp",      // LOG_FTP
	(16 << 3): "local0",   // LOG_LOCAL0
	(17 << 3): "local1",   // LOG_LOCAL1
	(18 << 3): "local2",   // LOG_LOCAL2
	(19 << 3): "local3",   // LOG_LOCAL3
	(20 << 3): "local4",   // LOG_LOCAL4
	(21 << 3): "local5",   // LOG_LOCAL5
	(22 << 3): "local6",   // LOG_LOCAL6
	(23 << 3): "local7",   // LOG_LOCAL7
}

func parseLog(msg []byte) string {
	// <NUM>
	if len(msg) < 5 || msg[0] != '<' {
		value := string(bytes.TrimSpace(msg))
		return fmt.Sprintf("unknown.0: %s", value)
	}
	level := 0
	pos := 1
	for ; pos < 5; pos++ {
		if msg[pos] >= '0' && msg[pos] <= '9' {
			level *= 10
			level += int(msg[pos] - '0')
		} else {
			break
		}
	}
	if msg[pos] != '>' {
		value := string(bytes.TrimSpace(msg))
		return fmt.Sprintf("unknown.0: %s", value)
	}
	var strLevel string
	if level < 192 {
		strLevel = fmt.Sprintf("%s.%s",
			facilities[level & ^0x7], priorities[level&0x7])
	} else {
		strLevel = fmt.Sprintf("unknown.%d", level)
	}
	// May  8 10:02:21
	if msg[pos+4] != ' ' || msg[pos+7] != ' ' || msg[pos+16] != ' ' {
		value := string(bytes.TrimSpace(msg[pos+1:]))
		return fmt.Sprintf("%s: %s", strLevel, value)
	}
	// Convert the rest (XXX: what happens with invalid utf8/low ascii?)
	value := string(bytes.TrimSpace(msg[pos+17:]))
	return fmt.Sprintf("%s: %s", strLevel, value)
}
