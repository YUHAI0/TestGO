package main

import (
	"fmt"
	"github.com/oxtoacart/go-udt/udt"
	"io"
	"net"
	"os"
)

func FileServerStart(address string, file *os.File) (err error) {

	addr, _ := net.ResolveUDPAddr("udp", address)

	pc, err := udt.ListenUDT("udp", addr)
	if err != nil {
		return
	}

	//fmt.Printf("%s\n", pc.Addr().String())

	var maxBufferSize = 4000
	buffer := make([]byte, maxBufferSize)

	print("Before")
	rw, aerr := pc.Accept()
	print("After")
	if aerr != nil {
		print("A error: ", aerr)
	}

	for {
		//n, _:= rw.Read(buffer)
		n, rerr := io.ReadFull(rw, buffer)
		if n == 0 {
			break
		}
		if rerr != nil {
			print("r error: ", rerr.Error())
			return rerr
		}

		_, err := file.Write(buffer)
		if err != nil {
			print("write error: ", err)
			return err
		}

		fmt.Printf("packet-received: bytes=%d from=%s\n",
			n, addr.String())
	}
	return
}

func main() {
	file := os.Args[1]
	pfile, err := os.Create(file)
	if err != nil {
		fmt.Printf("open file %s error: %s\n", file, err.Error())
	}
	err = FileServerStart("0.0.0.0:1900", pfile)
	if err != nil {
		print(err.Error())
		return
	}
}
