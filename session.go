package gomqtt

import (
	"net"
)

const (
	SessionStateConnect       = 0
	SessionStateConnectFinish = 1
)

type sessionType struct {
	conn  net.Conn
	state int
	close bool
	cmsg  chan messageRaw // Income message
}

func newSession(conn net.Conn) *sessionType {
	return &sessionType{conn: conn,
		state: SessionStateConnect,
		close: false,
		cmsg:  make(chan messageRaw, 100)}
}

func (s *sessionType) Normal() bool {
	return s.close == false
}

func (s *sessionType) Close() {
	s.close = true
}
