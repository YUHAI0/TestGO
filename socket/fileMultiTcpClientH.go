package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"sync"
)

var fileMutex = sync.Mutex{}
func singleClientSendH(address string, file *os.File, start int, total int, wg* sync.WaitGroup) {
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

	//_, err = file.Seek(int64(start), 0)
	if err != nil {
		println("seek err:", err)
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

		fileMutex.Lock()
		nRead, err := io.ReadFull(reader, buffer)
		fileMutex.Unlock()

		total += nRead
		println("", address, "Total:", total, "Left: ", left)
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
	/*
	host := os.Args[2]

	p64, _ := strconv.ParseInt(os.Args[3], 10, 64)
	port := int(p64)

	n64, _ := strconv.ParseInt(os.Args[4], 10, 64)
	number := int(n64)
	*/

	hosts := os.Args[2:]
	number := len(hosts)

	pfile, e:= os.Open(file)
	if e != nil {
		println(e.Error())
		return
	}

	pp, _ := pfile.Stat()
	maxSize := pp.Size()
	segment := maxSize / int64(number)
	println("max: ", maxSize)

	runtime.GOMAXPROCS(2)

	var wg = sync.WaitGroup{}
	println(5)
	var n = 0
	for _, host := range hosts {
		addr := host
		fmt.Printf("host: %s\n", addr)

		total := segment
		if n == n + number - 1 {
			total = maxSize - int64(n) * segment
			println("# max: dd: totall", maxSize, int64(n) * segment, total)
		}
		nn := n
		wg.Add(1)
		go func() {
			singleClientSendH(addr, pfile, nn*int(segment), int(total), &wg)
		}()
		n++
	}

	println(5)
	wg.Wait()
}
