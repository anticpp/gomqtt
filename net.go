package gomqtt

import (
	"net"
	"time"
)

func readConnTotal(conn net.Conn, buf []byte, timeout int) error {

	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}

	var n int
	var err error
	var pos = 0

	for pos < len(buf) {
		n, err = conn.Read(buf[pos:])
		if err != nil {
			return err
		}
		pos += n
	}

	return nil

}

func writeConnTotal(conn net.Conn, buf []byte, timeout int) error {

	if timeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}

	var n int
	var err error
	var pos = 0

	for pos < len(buf) {
		n, err = conn.Write(buf[pos:])
		if err != nil {
			return err
		}

		pos += n
	}
	return nil
}
