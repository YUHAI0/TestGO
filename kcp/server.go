package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
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

		for {
			println("accept...")
			s, err := listener.AcceptKCP()
			if err != nil {
				log.Fatal(err)
			}
			handleEcho(s, file)
		}
	} else {
		log.Fatal(err)
	}
}

// handleEcho send back everything it received
func handleEcho(conn *kcp.UDPSession, file string) {
	buf := make([]byte, bufsize)
	pfile, oErr := os.Create(file)
	if oErr != nil {
		println("create err: ", oErr.Error())
	}
	pindex := 0
	for {
		n, err := conn.Read(buf[:])


		_, err = pfile.Write(buf[:n])
		pindex += n

		fmt.Printf("receive %d bytes, total: %d \n", n, pindex)

		//_, err = io.CopyN(pfile, bytes.NewReader(buf[:n]), int64(n))

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
