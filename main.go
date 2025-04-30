package main

import (
	"io"
	"net"
)

func main() {
	l, err := net.Listen(LISTEN_NETWORK, ":"+LISTEN_PORT)
	panicIf(err, "error setting up listener")

	conn, err := l.Accept()
	panicIf(err, "error accepting connection")
	defer conn.Close()

	for {
		msgBuf := make([]byte, MAX_MSG_SIZE_BYTES)

		_, err := conn.Read(msgBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			panicIf(err, "error reading client message")
		}
		conn.Write([]byte("+OK\r\n"))
	}

}
