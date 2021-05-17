package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
)

func singleClientSend(address string, file *os.File, start int, total int, wg* sync.WaitGroup) {
	fmt.Printf("Send to %s, [%d, %d]\n", address, start, total)

	raddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		println("resolve: ", err.Error())
		return
	}

	reader := bufio.NewReader(file)

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		println("dial ", address, " err: ", err.Error())
		return
	}

	_, err = file.Seek(int64(start), 0)
	if err != nil {
		return
	}

	var left = total

	for {
		var buffer []byte
		const bufsize = 2000
		if left < bufsize {
			println("left: ", left)
			buffer = make([]byte, left)
		} else {
			buffer = make([]byte, bufsize)
		}

		nRead, err := io.ReadFull(reader, buffer)

		total += nRead
		println("", address, "Total:", total)
		if nRead == 0 {
			break
		}

		if err != nil {
			println("Read Err: ", err.Error())
			println("left: ", left, "nRead: ", nRead)

			return
		}
		fmt.Printf("read %d bytes\n", nRead)

		writer := bufio.NewWriter(conn)
		write, Werr := writer.Write(buffer)
		writer.Flush()
		//write, Werr := io.Copy(connu, pfile)

		if Werr != nil {
			print("Write err: ", Werr)
			return
		}
		fmt.Printf("packet-written: bytes=%d\n", write)
		left -= bufsize
		if left <= 0 {
			break
		}
	}

	//copyN, err := io.CopyN(conn, reader, int64(total))
	//if err != nil {
	//	return
	//}
	//
	//fmt.Printf("packet-written: bytes=%d\n", copyN)
	println("DONE")
	wg.Done()
}

func main() {
	file := os.Args[1]
	host := os.Args[2]

	p64, _ := strconv.ParseInt(os.Args[3], 10, 64)
	port := int(p64)

	n64, _ := strconv.ParseInt(os.Args[4], 10, 64)
	number := int(n64)

	pfile, e:= os.Open(file)
	if e != nil {
		println(e.Error())
		return
	}

	pp, _ := pfile.Stat()
	maxSize := pp.Size()
	segment := maxSize / int64(number)
	println("max: ", maxSize)

	runtime.GOMAXPROCS(1)

	var wg = sync.WaitGroup{}
	println(5)
	for p, n := port, 0; p < port + number; p, n = p+1, n+1 {
		addr := fmt.Sprintf("%s:%d", host, p)
		fmt.Printf("host: %s\n", addr)

		total := segment
		if p == port + number - 1 {
			total = maxSize - int64(n) * segment
			println("# max: dd: totall", maxSize, int64(n) * segment, total)
		}
		wg.Add(1)
		n := n
		go func() {
			singleClientSend(addr, pfile, n*int(segment), int(total), &wg)
		}()
	}

	println(5)
	wg.Wait()
}
