package syslog2stdout

import (
	"fmt"
	"net"
)

type syslogdUDP struct {
	name string
	conn net.PacketConn
}

func newUDP(port int) (Syslogd, error) {
	var listenAddr = net.UDPAddr{IP: nil, Port: port, Zone: ""}
	conn, err := net.ListenUDP("udp", &listenAddr)
	if err != nil {
		return nil, err
	}

	s := &syslogdUDP{name: fmt.Sprintf("UDP(%d)", port), conn: conn}
	return s, nil
}

func (s *syslogdUDP) HandleAll() {
	handleAll(s, s.conn)
}

func (s *syslogdUDP) Close() {
	s.conn.Close()
}

func (s *syslogdUDP) Description() string {
	return s.name
}

func (s *syslogdUDP) Addr2Prefix(addr *net.Addr) string {
	return fmt.Sprintf("%s: ", (*addr).String())
}
