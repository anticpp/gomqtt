package gomqtt

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

const (
	READ_BUFFER_CAP    = 4 * 1024
	INCOME_CHAN_SIZE   = 100
	OUTGOING_CHAN_SIZE = 100
)

type sessionType struct {
	conn        net.Conn
	connectInfo messageConnect
	readBuff    *bytes.Buffer
	errorOccur  bool
	inMessage   chan *messageRaw  // Incoming message
	outMessage  chan *messageType // Outgoing message
}

func newSession(conn net.Conn, message messageConnect) *sessionType {
	return &sessionType{conn: conn,
		connectInfo: message,
		readBuff:    bytes.NewBuffer(make([]byte, 0, READ_BUFFER_CAP)),
		errorOccur:  false,
		inMessage:   make(chan *messageRaw, INCOME_CHAN_SIZE),
		outMessage:  make(chan *messageType, OUTGOING_CHAN_SIZE)}
}

func (session *sessionType) normal() bool {
	return session.errorOccur == false
}

func (session *sessionType) close() {
	session.errorOccur = true
}

func (session *sessionType) start() {

	go session.serve_read()
	go session.serve_incoming()
	//go session.serve_outgoing()
}

func (session *sessionType) serve_read() {
	var n int
	var err error

	conn := session.conn

	tmpBuf := make([]byte, 1024)
	for session.normal() {

		err = session.frame_input()
		if err != nil {
			fmt.Println(err)
			session.close()
			continue
		}

		conn.SetReadDeadline(time.Now().Add(time.Duration(5) * time.Second))
		n, err = conn.Read(tmpBuf)
		if err != nil {
			nerr, ok := err.(*net.OpError)
			if !(ok && nerr.Timeout()) {
				session.close()
			}
			fmt.Println(err)

			continue
		}

		//fmt.Printf("Read %v bytes\n", n)
		session.readBuff.Write(tmpBuf[:n])
	}
}

func (session *sessionType) frame_input() error {

	var n int
	var err error

	for {
		message := newMessageRaw()
		n, err = message.header.decode(session.readBuff.Bytes())
		if err != nil {
			_, ok := err.(ErrorDecodeMore)
			if ok {
				return nil
			}
			return err
		}
		//fmt.Printf("Header size %v\n", n)
		session.readBuff.Next(n)

		//fmt.Println("Header: ", message.header)
		//fmt.Printf("Remaining %v\n", session.readBuff.Len())

		if session.readBuff.Len() < message.header.Length {
			return nil
		}

		message.payload = make([]byte, message.header.Length)
		session.readBuff.Read(message.payload)
		//fmt.Printf("DEBUG, Length: %v, Payload: %v\n", message.header.Length, len(message.payload))

		session.inMessage <- message
	}
}

func (session *sessionType) serve_incoming() {

	for {
		message := <-session.inMessage
		fmt.Printf("Received %v from %v\n", message.header.TypeName(), session.connectInfo.clientId)

		if message.header.Type == MessageTypePub {
			session.handle_publish(message)
		}

	}

}
func (session *sessionType) handle_publish(raw *messageRaw) {
	var err error

	m := newMessagePub()
	m.setHeader(raw.header)
	_, err = m.decodePayload(raw.payload)
	if err != nil {
		fmt.Println("Decode publish payload fail, client %v", session.connectInfo.clientId)
		return
	}
	fmt.Println(m)
}

func (s *sessionType) serve_outgoing() {

}
