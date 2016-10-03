package gomqtt

import (
	"bytes"
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
	var n int
	var err error

	// Read fix header
	fmt.Println(conn.LocalAddr())
	fmt.Println(conn.RemoteAddr())
	defer conn.Close()

	for {

		rbuf := bytes.NewBuffer(make([]byte, 0, 4*1024))
		for {
			header := fixHeader{}

			// Read fix header
			tmpb := make([]byte, 128)
			for {
				n, err = conn.Read(tmpb)
				if err != nil {
					fmt.Println(err)
					return
				} else if n == 0 {
					fmt.Println("Read EOF from conn")
					return
				}
				fmt.Printf("Read %v bytes\n", n)

				// Append to read buffer
				rbuf.Write(tmpb[:n])

				n, err = header.decode(rbuf.Bytes())
				if err != nil {
					if _, ok := err.(ErrorDecodeMore); !ok {
						fmt.Println(err)
						return
					}

					// Read more
					continue
				}

				fmt.Printf("Decode success, n %v\n", n)

				// Read fix header complete
				rbuf.Truncate(n)
				break
			}

			fmt.Println("Read header: ")
			fmt.Println(header)
			break
		}

		break
	}
}
