package gomqtt

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type Context struct {
	listenAddr string
	loop       bool
	ln         net.Listener
}

func New() *Context {
	return &Context{loop: true}
}

func (c *Context) Stop() {
	c.loop = false
}

func (c *Context) Listen(addr string) error {
	c.listenAddr = addr

	var err error
	c.ln, err = net.Listen("tcp", c.listenAddr)
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) Loop() error {

	for c.loop {

		conn, err := c.ln.Accept()
		if err != nil {
			continue
		}

		fmt.Println(conn.LocalAddr())
		fmt.Println(conn.RemoteAddr())

		//session := newSession(conn)
		//go serve_read(session)
		//go serve_write(session)
		//go serve_message(session)
		//go serve_state(session)
		go serve_connect(conn)
	}

	return nil
}

func serve_connect(conn net.Conn) {
	var message = new(messageConnect)
	var n int
	var err error
	var errorOccur bool = false

	readBuf := bytes.NewBuffer(make([]byte, 0, 4*1024))
	tmpBuf := make([]byte, 1024)

	// Read header
	var header fixHeader
	for !errorOccur {

		conn.SetReadDeadline(time.Now().Add(time.Duration(5) * time.Second))
		n, err = conn.Read(tmpBuf)
		if err != nil {
			errorOccur = true
			continue
		}
		readBuf.Write(tmpBuf[:n])

		n, err = header.decode(readBuf.Bytes())
		if err != nil {
			_, ok := err.(ErrorDecodeMore)
			if !ok {
				errorOccur = true
			}
			continue
		}

		// Read complete
		readBuf.Next(n)
		break

	}
	message.setHeader(header)

	// Read payload
	for !errorOccur && readBuf.Len() < header.Length {

		conn.SetReadDeadline(time.Now().Add(time.Duration(5) * time.Second))
		n, err = conn.Read(tmpBuf)
		if err != nil {
			errorOccur = true
			continue
		}
		readBuf.Write(tmpBuf[:n])

	}

	if errorOccur {
		conn.Close()
		return
	}

	payload := make([]byte, header.Length)
	readBuf.Read(payload)

	n, err = message.decode(payload)
	if err != nil {
		conn.Close()
		return
	}

	// Connect success
	fmt.Println("Connect Success:")
	fmt.Println(message)
	fmt.Printf("Remaining buffer %v", readBuf.Len())
}

func serve_read(session *sessionType) {
	var n int
	var err error

	conn := session.conn

	readBuf := bytes.NewBuffer(make([]byte, 0, 4*1024))
	tmpBuf := make([]byte, 1024)
	for session.Normal() {

		conn.SetReadDeadline(time.Now().Add(time.Duration(5) * time.Second))
		n, err = conn.Read(tmpBuf)
		if err != nil {
			nerr, ok := err.(*net.OpError)
			if !(ok && nerr.Timeout()) {
				session.Close()
			}
			fmt.Println(err)

			continue
		}

		fmt.Printf("Read %v bytes\n", n)
		readBuf.Write(tmpBuf[:n])

		err = process_input(session, readBuf)
		if err != nil {
			fmt.Println(err)
			session.Close()
			continue
		}

	}
}

func process_input(session *sessionType, readBuf *bytes.Buffer) error {

	var n int
	var err error
	for {
		message := messageRaw{}
		n, err = message.header.decode(readBuf.Bytes())
		if err != nil {
			_, ok := err.(ErrorDecodeMore)
			if ok {
				return nil
			}
			return err
		}
		fmt.Printf("Header size %v\n", n)
		readBuf.Next(n)

		fmt.Println("Header: ")
		fmt.Println(message.header)
		fmt.Printf("Remaining %v\n", readBuf.Len())

		if readBuf.Len() < message.header.Length {
			return nil
		}

		message.payload = make([]byte, message.header.Length)
		readBuf.Read(message.payload)
		fmt.Printf("Payload: %v\n", len(message.payload))

		session.inMessage <- message
	}
}

func serve_message(session *sessionType) {

	for {
		raw := <-session.inMessage

		var message messageType
		if raw.header.Type == MessageTypeConnect {
			message = &messageConnect{}
		} else {
			fmt.Println("Unsupport message %v", raw.header.Type)
			session.Close()
			break
		}
		message.setHeader(raw.header)
		message.decode(raw.payload)
		fmt.Println("serve message: ", message)
		//fmt.Println(raw)

	}

}

func serve_state(session *sessionType) {

}
