package gomqtt

import (
	"bytes"
	"net"
)

const (
	READ_BUFFER_CAP  = 4 * 1024
	INCOME_CHAN_SIZE = 100
)

type sessionType struct {
	conn        net.Conn
	connectInfo messageConnect
	readBuff    *bytes.Buffer
	errorOccur  bool
	inMessage   chan messageRaw // Incoming message
}

func newSession(conn net.Conn, message messageConnect) *sessionType {
	return &sessionType{conn: conn,
		connectInfo: message,
		readBuff:    bytes.NewBuffer(make([]byte, 0, READ_BUFFER_CAP)),
		errorOccur:  false,
		inMessage:   make(chan messageRaw, INCOME_CHAN_SIZE)}
}

func (s *sessionType) normal() bool {
	return s.errorOccur == false
}

func (s *sessionType) close() {
	s.errorOccur = true
}
