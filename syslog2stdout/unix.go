package syslog2stdout

import (
	"fmt"
	"net"
	"syscall"
)

type syslogdUnixgram struct {
	filename string
	conn net.PacketConn
}

func newUnixgram(filename string) (Syslogd, error) {
	listenAddr := net.UnixAddr{Name: filename, Net: "unixgram"}
	conn, err := net.ListenUnixgram("unixgram", &listenAddr)
	if err != nil {
		return nil, err
	}
	s := &syslogdUnixgram{filename: filename, conn: conn}
	return s, nil
}

func (s *syslogdUnixgram) HandleAll() {
	handleAll(s, s.conn)
}

func (s *syslogdUnixgram) Close() {
	s.conn.Close()
	syscall.Unlink(s.filename)
}

func (s *syslogdUnixgram) Description() string {
	return fmt.Sprintf("UNIX(%s)", s.filename)
}

func (s *syslogdUnixgram) Addr2Prefix(addr *net.Addr) string {
	return ""
}
