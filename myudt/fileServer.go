package main

import (
	"encoding/json"
	"fmt"
	"github.com/oxtoacart/go-udt/myudt/proto"
	"net"
	"os"
	"sync"
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

	var fileMutex = sync.RWMutex{}
	for {
		n, addr, _:= pc.ReadFrom(buffer)
		data, _:= proto.Deserialize(buffer)
		if !proto.Validate(data) {
			j, _ := json.Marshal(data)
			fmt.Printf("Packet not valid: %s\n", j)
			continue
		}

		go func() {
			fileMutex.Lock()
			//copyN, err := io.Copy(file, bytes.NewReader(data.Data))
			copyN, err := file.WriteAt(data.Data, data.Seek)

			writeTotal += int64(copyN)
			fmt.Printf("seq: %d, Total %d,\tWrite %d bytes, \n", data.Seq, writeTotal, copyN)

			if err != nil {
				println("Err: ", err.Error())
			}

			fileMutex.Unlock()
		}()

		fmt.Printf("packet-%d-received: version: %s, id: %s, seek: %d, length: %d, bytes=%d from=%s\n",
			data.Seq,
			data.Version, data.Id, data.Seek, data.Length, n, addr.String())

		ack := proto.NewACK(data.Id)
		var ad, _ = proto.ACKSer(ack)
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
