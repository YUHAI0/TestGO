package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)


func main() {
	file := os.Args[1]
	host := os.Args[2]

	address := host
	raddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return
	}

	if err != nil {
		return
	}

	pfile, e := os.Open(file)
	if e != nil {
		fmt.Printf("open file %s error: %s\n", file, e.Error())
	}
	reader := bufio.NewReader(pfile)

	var buffer = make([]byte, 2000)

	connu, err := net.DialTCP("tcp", nil, raddr)

	for {
		nRead, err := io.ReadFull(reader, buffer)
		if err != nil {
			return
		}
		fmt.Printf("read %d bytes\n", nRead)

		//writer := bufio.NewWriter(connu)
		//write, err := writer.Write(buffer)
		write, err := io.CopyBuffer(connu, reader, buffer)

		if err != nil {
			print("Write err: ", err)
			return
		}
		fmt.Printf("packet-written: bytes=%d\n", write)
	}
}
