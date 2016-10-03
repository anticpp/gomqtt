package gomqtt

import (
	"fmt"
	"net"
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

		go serve_conn(conn)

	}

	return nil
}

func serve_conn(conn net.Conn) {
	/*
		var n int
		var err error
		var buf []byte
		var header fixHeader
	*/
	// Read fix header
	fmt.Println(conn.LocalAddr())
	fmt.Println(conn.RemoteAddr())
	defer conn.Close()

	for {
		/*
			// Read 1'st byte of fix header.
			buf = make([]byte, 1)
			n, err = conn.Read(buf)
			if err != nil {
				fmt.Println(err)
				return
			} else if n == 0 {
				fmt.Println("Read EOF from conn")
				return
			}
			var b = buf[0]
			header.Type = int32(0x0F & (b >> 4))
			header.Dup = int32(0x01 & (b >> 3))
			header.Qos = int32(0x03 & (b >> 1))
			header.Retain = int32(0x01 & b)

			fmt.Println(header)

			// Read remaining length.
			// We try to read the most 4 bytes.
			buf = make([]byte, 4)
			n, err = conn.Read(buf)
			if err != nil {
				fmt.Println(err)
				return
			} else if n == 0 {
				fmt.Println("Read EOF from conn")
				return
			}

			break*/
	}
}
