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
	inMessage   chan *messageRaw // Incoming message
	outMessage  chan messageType // Outgoing message
	closeSignal chan int
}

func newSession(conn net.Conn, message messageConnect) *sessionType {
	return &sessionType{conn: conn,
		connectInfo: message,
		readBuff:    bytes.NewBuffer(make([]byte, 0, READ_BUFFER_CAP)),
		errorOccur:  false,
		inMessage:   make(chan *messageRaw, INCOME_CHAN_SIZE),
		outMessage:  make(chan messageType, OUTGOING_CHAN_SIZE),
		closeSignal: make(chan int, 10)}
}

func (session *sessionType) normal() bool {
	return session.errorOccur == false
}

func (session *sessionType) close() {
	session.errorOccur = true
	fmt.Printf("Close session %v\n", session.conn.RemoteAddr())
}

func (session *sessionType) start() {

	go session.serve_read()
	go session.serve_message()
	go session.serve_write()
	go session.serve_close()
}

func (session *sessionType) send_message(m messageType) {
	session.outMessage <- m
}

func (session *sessionType) serve_read() {
	var n int
	var err error

	conn := session.conn

	tmpBuf := make([]byte, 1024)
	for session.normal() {

		err = session.frame()
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
			fmt.Println("Read error ", err)

			continue
		}

		session.readBuff.Write(tmpBuf[:n])
	}

	fmt.Printf("End serve_read %v\n", session.conn.RemoteAddr())
	session.closeSignal <- 0
}

func (session *sessionType) frame() error {

	var n int
	var err error

	for {
		err = nil
		message := newMessageRaw()
		n, err = message.header.decode(session.readBuff.Bytes())
		if err != nil {
			_, ok := err.(ErrorDecodeMore)
			if ok {
				err = nil
			}
			break
		}

		session.readBuff.Next(n)
		if session.readBuff.Len() < message.header.Length {
			break
		}

		message.payload = make([]byte, message.header.Length)
		session.readBuff.Read(message.payload)

		session.inMessage <- message
	}
	return err
}

func (session *sessionType) serve_message() {

	for session.normal() {
		select {
		case message := <-session.inMessage:
			fmt.Printf("Received %v from %v\n", message.header.typeName(), session.connectInfo.clientId)

			if message.header.Type == MessageTypePub {
				session.handle_publish(message)
			}

		case <-time.After(time.Second * 3):

		}

	}

	fmt.Printf("End serve_message %v\n", session.conn.RemoteAddr())

}
func (session *sessionType) handle_publish(raw *messageRaw) {
	var err error

	req := newMessagePub()
	req.setHeader(raw.header)
	_, err = req.decodePayload(raw.payload)
	if err != nil {
		fmt.Println("Decode publish payload fail, client %v", session.connectInfo.clientId)
		return
	}
	fmt.Println(req)

	if req.getHeader().getQos() > QoS0 {
		resp := newMessagePubAck()
		resp.packetId = req.packetId

		session.send_message(resp)
	}
}

func (session *sessionType) serve_write() {

	for session.normal() {

		select {
		case message := <-session.outMessage:
			fmt.Printf("Send %v to %v\n", message.getHeader().typeName(), session.connectInfo.clientId)
			fmt.Println(message)

			data, _ := message.encode(nil)

			// FIXME: write error or timeout
			writeConnTotal(session.conn, data, 0)

		case <-time.After(time.Second * 3):
			fmt.Println("Timeout for serve_write")
		}

	}

	fmt.Printf("End serve_write %v\n", session.conn.RemoteAddr())
	session.closeSignal <- 1
}

func (session *sessionType) serve_close() {
	// Wait for serve_read, serve_write
	for i := 0; i < 2; i++ {
		<-session.closeSignal
	}

	fmt.Printf("Closing %v\n", session.conn.RemoteAddr())
	session.conn.Close()
}
