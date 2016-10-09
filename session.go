package gomqtt

import (
	"net"
)

const (
	SessionStateNew     = 0
	SessionStateConnect = 1
)

type sessionType struct {
	conn      net.Conn
	state     int
	close     bool
	inMessage chan messageRaw // Income message
}

func newSession(conn net.Conn) *sessionType {
	return &sessionType{conn: conn,
		state:     SessionStateConnect,
		close:     false,
		inMessage: make(chan messageRaw, 100)}
}

func (s *sessionType) Normal() bool {
	return s.close == false
}

func (s *sessionType) Close() {
	s.close = true
}
