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

	connu, err := net.DialTCP("tcp", nil, raddr)

	total := 0
	for {
		var buffer = make([]byte, 2000)
		nRead, err := io.ReadFull(reader, buffer)
		total += nRead
		print("Total:", total)
		if nRead == 0 {
			break
		}

		if err != nil {
			print("Read Err: ", err)
			return
		}
		fmt.Printf("read %d bytes\n", nRead)

		writer := bufio.NewWriter(connu)
		write, Werr := writer.Write(buffer)
		writer.Flush()
		//write, Werr := io.Copy(connu, pfile)

		if Werr != nil {
			print("Write err: ", Werr)
			return
		}
		fmt.Printf("packet-written: bytes=%d\n", write)
	}
	print("DONE")

}
