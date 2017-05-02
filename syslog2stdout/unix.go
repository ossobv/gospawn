package syslog2stdout

import (
	"net"
	"os"
)

type ListenerUnixgram struct {
	filename string
	conn net.PacketConn
}

func listenUnixgram(filename string) (Listener, error) {
	listenAddr := net.UnixAddr{Name: filename, Net: "unixgram"}
	conn, err := net.ListenUnixgram("unixgram", &listenAddr)
	if err != nil {
		return nil, err
	}
	l := &ListenerUnixgram{filename: filename, conn: conn}
	return l, nil
}

func (l *ListenerUnixgram) HandleAll() {
	handleAll(l.conn)
}

func (l *ListenerUnixgram) Close() {
	l.conn.Close()
	os.Remove(l.filename)
}
