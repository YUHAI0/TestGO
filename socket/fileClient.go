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
	raddr, err := net.ResolveUDPAddr("udp", address)
	print(raddr)
	if err != nil {
		return
	}

	//conn, err := net.DialUDP("udp", nil, raddr)

	if err != nil {
		return
	}

	//defer conn.Close()

	pfile, e := os.Open(file)
	if e != nil {
		fmt.Printf("open file %s error: %s\n", file, e.Error())
	}
	reader := bufio.NewReader(pfile)

	var buffer = make([]byte, 2000)

	connu, err := net.DialUDP("udp", nil, raddr)

	for {
		nRead, err := io.ReadFull(reader, buffer)
		fmt.Printf("read %d bytes\n", nRead)
		if nRead == 0 {
			break
		}

		if err != nil {
			print("read full err:", err)
			return
		}

		//n, err := io.CopyBuffer(conn, reader, buffer)
		//if err != nil {
		//	fmt.Printf("Err: %s\n", err.Error())
		//	return
		//}
		//write, err := conn.Write(buffer)
		//breader := bytes.NewReader(buffer)
		//print(host, " ", err)

		//write, err := bufio.NewReader(connu).Read(buffer)
		writer := bufio.NewWriter(connu)
		write, err := writer.Write(buffer)
		writer.Flush()
		if write == 0 {
			break
		}

		if err != nil {
			print("Write err: ", err)
			return
		}
		fmt.Printf("packet-written: bytes=%d\n", write)
	}
}
