package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)


func main() {

	address := "localhost:1900"
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, raddr)

	if err != nil {
		return
	}

	defer conn.Close()

	//doneChan := make(chan error, 1)

	//reader := bufio.NewReader(os.Stdin)
	//os.Args[0]
	//file := "socket/client.go"
	file := os.Args[1]
	pfile, e := os.Open(file)
	if e != nil {
		fmt.Printf("open file %s error: %s\n", file, e.Error())
	}
	reader := bufio.NewReader(pfile)

	n, err := io.Copy(conn, reader)
	if err != nil {
		return
	}

	fmt.Printf("packet-written: bytes=%d\n", n)
}
