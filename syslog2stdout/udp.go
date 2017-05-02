package syslog2stdout

import (
	"net"
)

type ListenerUdp struct {
	conn net.PacketConn
}

func listenUdp(port int) (Listener, error) {
	var listenAddr = net.UDPAddr{IP: nil, Port: port, Zone: ""}
	conn, err := net.ListenUDP("udp", &listenAddr)
	if err != nil {
		return nil, err
	}
	l := &ListenerUdp{conn: conn}
	return l, nil
}

func (l *ListenerUdp) HandleAll() {
	handleAll(l.conn)
}

func (l *ListenerUdp) Close() {
	l.conn.Close()
}
