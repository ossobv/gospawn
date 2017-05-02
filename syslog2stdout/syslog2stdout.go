// Package syslog2stdout handles listening for syslog messages and
// writing them to stdout.
package syslog2stdout

import (
	"fmt"
	"net"
	"strconv"
)

type Listener interface {
	HandleAll()
	Close()
}

// Listen opens an UDP socket or Unix DGRAM socket depending on whether
// the supplied string looks like an integer or not.
func Listen(portOrFilename string) (Listener, error) {
	port, err := strconv.Atoi(portOrFilename)
	if err == nil {
		return listenUdp(port)
	}
	return listenUnixgram(portOrFilename)
}

func handleAll(conn net.PacketConn) {
	buf := make([]byte, 8192)
	for {
		n, addr, err := conn.ReadFrom(buf)
		fmt.Printf("[%s, %s, %s]\n", n, addr, err)
	}
}
