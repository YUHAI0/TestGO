package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)



func fileServerStart(address string, file *os.File) (err error) {

	tcpaddr, _ := net.ResolveTCPAddr("tcp", address)
	pc, err := net.ListenTCP("tcp", tcpaddr)
	//pc, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		print("Listen err: ", err)
		return
	}
	fmt.Printf("%s\n", pc.Addr().String())

	print("Wait to accept...")
	conn, Aerr := pc.Accept()
	if Aerr != nil {
		print("Accept error: ", Aerr)
		return
	}

	reader := bufio.NewReader(conn)
	var maxBufferSize = 2000
	buffer := make([]byte, maxBufferSize)
	for {
		//n, errRead := conn.Read(buffer)
		//if errRead != nil {
		//	print("Read err: ", errRead)
		//	return
		//}

		Serr := conn.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
		if Serr != nil {
			print("Set err: ", Serr)
			return Serr
		}

		full, rerr := io.ReadFull(reader, buffer)
		if  rerr != nil {
			return rerr
		}
		print("read ", full, " bytes ")
		if full == 0 {
			break
		}

		_, err = file.Write(buffer)
		if err != nil {
			print("file w e:", err)
			return err
		}

		fmt.Printf("packet-received: bytes=%d from=%s\n",
			full, tcpaddr.String())

	}
	return
}

func main() {
	file := os.Args[1]
	pfile, err := os.Create(file)
	if err != nil {
		fmt.Printf("open file %s error: %s\n", file, err.Error())
	}
	err = fileServerStart("0.0.0.0:1900", pfile)
	if err != nil {
		return
	}
}
