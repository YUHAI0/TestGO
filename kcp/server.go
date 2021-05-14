package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
	"io"

	//"net"

	"log"
	"os"
)

func main() {
	file := os.Args[1]
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	host := "0.0.0.0:1001"
	println(host)
	//if listener, err := kcp.Listen(host); err == nil {
	if listener, err := kcp.ListenWithOptions(host, block, 10, 3); err == nil {
		// spin-up the client
		println("accept...")
		s, err := listener.AcceptKCP()
		if err != nil {
			log.Fatal(err)
		}

		for {
			handleEcho(s, file)
		}
	} else {
		log.Fatal(err)
	}
}

// handleEcho send back everything it received
func handleEcho(conn *kcp.UDPSession, file string) {
	buf := make([]byte, 4096)
	pfile, oErr := os.Create(file)
	if oErr != nil {
		println("create err: ", oErr.Error())
	}
	pindex := 0
	for {
		println("r1")
		n, err := conn.Read(buf[:])
		println("r2")
		fmt.Printf("receive %d bytes\n", n)

		//_, err = pfile.WriteAt(buf, int64(pindex))
		pindex += n

		_, err = io.CopyN(pfile, bytes.NewReader(buf[:n]), int64(n))

		if err != nil {
			println("copy err: ", err.Error())
			return
		}

		if err != nil {
			log.Println(err)
			return
		}
	}
}
