package main

import (
	"bufio"
	"fmt"
	"github.com/oxtoacart/go-udt/udt"
	"io"
	"net"
	"os"
)

func main() {
	file := os.Args[1]
	host := os.Args[2]

	address := host
	raddr, err := net.ResolveUDPAddr("udp", address)
	print(raddr)
	if err != nil {
		return
	}

	if err != nil {
		return
	}

e:

	pfile, e := os.Open(file)
	if e != nil {
		fmt.Printf("open file %s error: %s\n", file, e.Error())
	}
	reader := bufio.NewReader(pfile)

	var buffer = make([]byte, 2000)

	connu, err := udt.DialUDT("udp", nil, raddr)

	for {
		nRead, err := io.ReadFull(reader, buffer)
		fmt.Printf("read %d bytes\n", nRead)
		if nRead == 0 {
			goto e
			break
		}

		if err != nil {
			print("read full err:", err)
			return
		}

		writer := bufio.NewWriter(connu)
		write, err := writer.Write(buffer)
		writer.Flush()
		if write == 0 {
			goto e
			break
		}

		if err != nil {
			print("Write err: ", err)
			return
		}
		fmt.Printf("packet-written: bytes=%d\n", write)
	}
}
