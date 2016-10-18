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

		go serve_connect(conn)
	}

	return nil
}

func serve_connect(conn net.Conn) {

	var connectReq = newMessageConnect()
	var n int
	var headerSize int
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

		headerSize, err = header.decode(readBuf.Bytes())
		if err != nil {
			_, ok := err.(ErrorDecodeMore)
			if !ok {
				errorOccur = true
			}
			continue
		}

		// Success
		break
	}

	if errorOccur {
		conn.Close()
		return
	}

	fmt.Printf("Header decode len %v\n", headerSize)
	readBuf.Next(headerSize)

	connectReq.setHeader(header)

	// Read remaining payload
	payload := make([]byte, header.Length)
	n, _ = readBuf.Read(payload)
	fmt.Printf("Now payload size %v, Remaining %v\n", n, header.Length-n)
	if header.Length-n > 0 {

		err = readConnTotal(conn, payload[n:], 5)
		if err != nil {
			fmt.Println("Read remaining payload fail")
			conn.Close()
			return
		}
	}

	n, err = connectReq.decodePayload(payload)
	if err != nil {
		fmt.Println("Decode connect error")
		conn.Close()
		return
	}
	fmt.Printf("Payload decode len %v\n", n)

	// Connect success
	fmt.Println("Connect Success:")
	fmt.Println(connectReq)

	// Response

	connectAck := newMessageConnectAck()
	var respBuff []byte
	respBuff, err = connectAck.encode(nil)
	if err != nil {
		fmt.Println("Encode connectAck error")
		conn.Close()
		return
	}
	fmt.Printf("Resp connectAck message len %v\n", len(respBuff))
	err = writeConnTotal(conn, respBuff, 5)
	if err != nil {
		fmt.Println("write connectAck error")
		conn.Close()
		return
	}

	session := newSession(conn, *connectReq)
	fmt.Printf("Remaining buffer %v, keep in session if some.\n", readBuf.Len())
	if readBuf.Len() > 0 {
		session.readBuff.Write(readBuf.Bytes())
	}

	go serve_read(session)
}

func serve_read(session *sessionType) {
	var n int
	var err error

	conn := session.conn

	tmpBuf := make([]byte, 1024)
	for session.normal() {

		err = process_input(session)
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

		fmt.Printf("Read %v bytes\n", n)
		session.readBuff.Write(tmpBuf[:n])
	}
}

func process_input(session *sessionType) error {

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
		fmt.Printf("Header size %v\n", n)
		session.readBuff.Next(n)

		fmt.Println("Header: ", message.header)
		fmt.Printf("Remaining %v\n", session.readBuff.Len())

		if session.readBuff.Len() < message.header.Length {
			return nil
		}

		message.payload = make([]byte, message.header.Length)
		session.readBuff.Read(message.payload)
		fmt.Printf("Payload: %v\n", len(message.payload))

		session.inMessage <- message
	}
}

func serve_message(session *sessionType) {

	for {
		raw := <-session.inMessage

		fmt.Println("serve message: ", raw)

	}

}
