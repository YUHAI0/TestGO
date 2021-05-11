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

	//reader := bufio.NewReader(pfile)
	var buffer = make([]byte, 4000)

	for {
		//nRead, err := io.ReadFull(reader, buffer)
		nRead, err  := pfile.Read(buffer[:])
		fmt.Printf("read %d bytes\n", nRead)

		if err != nil && err != io.EOF {
			print("read full err:", err.Error())
		}

		//writer := bufio.NewWriter(conn)
		//write, err := writer.Write(buffer)
		n, _ := io.CopyN(conn, bytes.NewReader(buffer), int64(nRead))
		//writer.Flush()

		fmt.Printf("packet-written: bytes=%d\n", n)

		if nRead == 0 || err == io.EOF {
			break
		}
	}
}
