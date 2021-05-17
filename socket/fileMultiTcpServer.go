package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
)

func singleServer(addr string, file *os.File, wg* sync.WaitGroup) {
	tcpaddr, _ := net.ResolveTCPAddr("tcp", addr)

	pc, err := net.ListenTCP("tcp", tcpaddr)

	if err != nil {
		print("Listen err: ", err)
		wg.Done()
		return
	}

	fmt.Printf("%s Wait to accept...\n", pc.Addr().String())
	conn, Aerr := pc.Accept()

	if Aerr != nil {
		print("Accept error: ", Aerr)
		wg.Done()
	}

	total := 0

	reader := bufio.NewReader(conn)
	var maxBufferSize = 2000
	buffer := make([]byte, maxBufferSize)

	for {
		full, rerr := io.ReadFull(reader, buffer)
		if  rerr != nil {
			println("read err: ", rerr.Error())
			goto end
		}
		total += full
		println("read ", full, " bytes ", " total: ", total, " bytes")
		if full == 0 {
			goto end
		}

		_, err = file.Write(buffer)
		if err != nil {
			print("file w e:", err)
			goto end
		}

		fmt.Printf("packet-received: bytes=%d from=%s\n",
			full, tcpaddr.String())
	}

	end:
		wg.Done()
}

func fileMultiServerStart(host string, portStart int, number int, file *os.File, wg* sync.WaitGroup) (err error) {

	for port:=portStart; port<portStart+number; port++ {
		addr := fmt.Sprintf("%s:%d", host, port)
		wg.Add(1)
		go singleServer(addr, file, wg)
	}
	return
}

func main() {
	println("args: file number")
	file := os.Args[1]
	number := os.Args[2]
	pfile, err := os.Create(file)
	if err != nil {
		fmt.Printf("open file %s error: %s\n", file, err.Error())
	}
	parseInt, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		println("p i e: ", err.Error())
		return
	}
	var wg = sync.WaitGroup{}
	err = fileMultiServerStart("0.0.0.0", 1900, int(parseInt), pfile, &wg)
	if err != nil {
		println("s start err ", err.Error())
		return
	}
	wg.Wait()
}
