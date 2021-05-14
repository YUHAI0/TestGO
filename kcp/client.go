package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"log"
	"os"
	"time"
)

const bufsize = 10000

func client(addr string, file string) {
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	// wait for server to become ready
	time.Sleep(time.Second)
	pfile, _ := os.Open(file)

	println(addr)
	// dial to the echo server
	count := 0
	if sess, err := kcp.DialWithOptions(addr, block, 10, 3); err == nil {
		for {
			buf := make([]byte, bufsize)

			nRead, _ := pfile.Read(buf[:])
			count += nRead
			if nRead == 0 {
				println("DONE: ", count)
				break
			}

			println("read ", nRead)

			nCopy, err := io.Copy(sess, bytes.NewReader(buf[:nRead]))

			if err == nil {
				fmt.Printf("write %d bytes\n", nCopy)
			} else {
				log.Fatal(err)
			}
		}
	} else {
		log.Fatal(err)
	}
}

func main() {
	addr := os.Args[1]
	file := os.Args[2]
	client(addr, file)

	sig := make(chan int)

	<-sig
}
