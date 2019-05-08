// Package syslog2stdout handles listening for syslog messages and
// writing them to stdout.
package syslog2stdout

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Syslogd handles incoming syslog packets, sending them on to stdout.
type Syslogd interface {
	// HandleAll handles all incoming packets until the socket is
	// closed.
	HandleAll()
	// Close closes the syslog socket.
	Close()

	// Description returns a description of the socket.
	Description() string
	// Addr2Prefix converts Syslogd implementation addresses to a prefix,
	// if necessary.
	Addr2Prefix(addr *net.Addr) string
}

// New opens up a UDP socket or Unix datagram socket depending on
// whether the supplied string looks like an integer or not.
func New(portOrFilename string) (Syslogd, error) {
	port, err := strconv.Atoi(portOrFilename)
	if err == nil {
		return newUDP(port)
	}
	return newUnixgram(portOrFilename)
}

// https://github.com/golang/go/issues/4373 -- do a string comparison to
// find out that the network connection is closed.
func isClosedErr(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "use of closed network connection") {
		return true
	}
	return false
}

func handleAll(syslogd Syslogd, conn net.PacketConn) {
	fmt.Fprintf(os.Stdout, "Spawned syslogd at %s\n",
		syslogd.Description())

	buf := make([]byte, 8192)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if isClosedErr(err) {
			fmt.Fprintf(os.Stderr, "DBG: %s: %s\n", syslogd.Description(),
				err.Error())
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: %s: %s\n", syslogd.Description(),
				err.Error())
			continue
		}

		// FIXME: the buf has to be parsed syslog-style
		fmt.Fprintf(os.Stdout, "%s%s\n", syslogd.Addr2Prefix(&addr),
			parseLog(buf[:n]))
	}
}
