package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
)

func FileServerStart(address string, file *os.File) (err error) {
	var maxBufferSize = 4000

	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", pc.LocalAddr().String())

	buffer := make([]byte, maxBufferSize)

	for {
		n, addr, _:= pc.ReadFrom(buffer)

		copyN, err := io.CopyN(file, bytes.NewReader(buffer), int64(n))
		if err != nil {
			return err
		}

		fmt.Printf("packet-received: bytes=%d , writed=%d, from=%s\n" ,n, copyN, addr.String())

		pc.WriteTo([]byte("DONE"), addr)
	}
}

func main() {
	file := os.Args[1]
	pfile, err := os.Create(file)
	if err != nil {
		fmt.Printf("open file %s error: %s\n", file, err.Error())
	}
	err = FileServerStart("0.0.0.0:1900", pfile)
	if err != nil {
		return
	}
}
