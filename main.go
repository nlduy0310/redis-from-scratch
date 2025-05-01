package main

import (
	"fmt"
	"io"
	"net"

	"github.com/nlduy0310/redis-from-scratch/server"
	"github.com/nlduy0310/redis-from-scratch/utils"
)

func main() {

	l, err := net.Listen(server.LISTEN_NETWORK, ":"+server.LISTEN_PORT)
	utils.PanicIf(err, "error setting up listener")

	conn, err := l.Accept()
	utils.PanicIf(err, "error accepting connection")
	defer conn.Close()

	resp := server.NewResp(conn)

	for {
		msg, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client closed connection. Shutting down...")
				return
			}
			utils.PanicIf(err, "error while reading message from client")
		}
		fmt.Printf("Received messaged from client:\n%s\n", msg.String())
		conn.Write([]byte("+OK\r\n"))
	}
}
