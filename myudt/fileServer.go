package main

import (
	"bytes"
	"fmt"
	"github.com/oxtoacart/go-udt/myudt/proto"
	"io"
	"net"
	"os"
)


func FileServerStart(address string, file *os.File) (err error) {
	var maxBufferSize = 5000

	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", pc.LocalAddr().String())

	buffer := make([]byte, maxBufferSize)

	var ackCount = 0
	var writeTotal int64 = 0
	for {
		n, addr, _:= pc.ReadFrom(buffer)
		data, _:= proto.Deserialize(buffer)

		//_, err := file.Write(data.Data)

		go func() {
			copyN, err := io.Copy(file, bytes.NewReader(data.Data))
			writeTotal += copyN
			fmt.Printf("Total %d,\tWrite %d bytes, \n", writeTotal, copyN)

			if err != nil {
				println("Err: ", err.Error())
			}
		}()

		fmt.Printf("packet-received: version: %s, id: %s, length: %d, bytes=%d from=%s\n",
			data.Version, data.Id, data.Length, n, addr.String())

		ack := proto.NewACK(data.Id)

		var ad, _ = proto.ACKSer(ack)

		//println("Waiting sleep...")
		//time.Sleep(10 * time.Millisecond)
		//println("Waiting done")
		wn, werr := pc.WriteTo(ad, addr)
		ackCount += 1
		println("ACK count: ", ackCount)

		fmt.Printf("Write to %d bytes\n", wn)

		if werr != nil {
			println("Write to err: ", werr)
			return werr
		}
	}
}

func main() {
	file := os.Args[1]
	pfile, err := os.Create(file)
	if err != nil {
		fmt.Printf("open file %s error: %s\n", file, err.Error())
	}
	err = FileServerStart("0.0.0.0:1900", pfile)
	if err != nil {
		println(err.Error())
		return
	}
}
