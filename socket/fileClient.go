package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
)


func main() {
	file := os.Args[1]
	host := os.Args[2]

	address := host
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, raddr)

	if err != nil {
		return
	}

	defer conn.Close()

	pfile, e := os.Open(file)
	if e != nil || pfile == nil {
		fmt.Printf("open file %s error: %s\n", file, e.Error())
		return
	}

	var buffer = make([]byte, 4000)

	for {
		nRead, err  := pfile.Read(buffer[:])
		fmt.Printf("read %d bytes\n", nRead)

		if err != nil && err != io.EOF {
			print("read full err:", err.Error())
		}

		n, _ := io.CopyN(conn, bytes.NewReader(buffer), int64(nRead))

		fmt.Printf("packet-written: bytes=%d\n", n)

		var echo []byte = make([]byte, 100)
		conn.ReadFrom(echo)
		fmt.Printf("echo %s", string(echo))

		if nRead == 0 || err == io.EOF {
			break
		}
	}
}
