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
	//var maxBufferSize = 100000000

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
	for {
		//n, errRead := conn.Read(buffer)
		//if errRead != nil {
		//	print("Read err: ", errRead)
		//	return
		//}

		Serr := conn.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
		if Serr != nil {
			return Serr
		}

		/*
		buffer := make([]byte, maxBufferSize)
		n, Rerr := reader.Read(buffer)
		if Rerr != nil {
			print("Read err: ", Rerr)
			return
		}

		_, err = file.Write(buffer)
		if err != nil {
			return err
		}

		fmt.Printf("packet-received: bytes=%d from=%s\n",
			n, tcpaddr.String())
		 */
		Copy, Cerr := io.Copy(file, reader)
		if Cerr != nil {
			print("Err: ", Cerr)
			return Cerr
		}
		if Copy == 0 {
			return
		}
		print("copy: ", Copy)
	}
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
