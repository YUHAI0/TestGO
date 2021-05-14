package main

import (
	"bytes"
	"fmt"
	"github.com/oxtoacart/go-udt/myudt/proto"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type FUDPConn struct {
	localAddr string
	remoteAddr string
	conn *net.UDPConn
}

func fdial(addr string) (conn FUDPConn) {
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return
	}

	newConn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}

	return FUDPConn{newConn.LocalAddr().String(), newConn.RemoteAddr().String(), newConn}
}

func (u*FUDPConn) sendData(data []byte, seq int, seek int64) error {
	return u.sendPacketWaitACK(proto.NewProto(data, seq, seek))
}

func (u*FUDPConn) sendDataN(data []byte, seq int, n int, seek int64) error {
	if n <= 0 {
		return nil
	}
	return u.sendPacketWaitACK(proto.NewProto(data[:n], seq, seek))
}

func (u*FUDPConn) sendPacket(packet proto.Proto) {
	buffer, _ := proto.Serialize(packet)

	_, _ = io.Copy(u.conn, bytes.NewReader(buffer))
}

var stopSignalMapMutex = sync.RWMutex{}
var stopSignalMap map[string] chan int = make(map[string] chan int)
var fgStopSignal chan int = make(chan int)

func (u *FUDPConn) readFromRemote() *proto.ACK {
	var buf = make([]byte, 2000)
	n, _, err := u.conn.ReadFrom(buf)
	if err != nil {
		println("readfrom Err: ", err.Error())
		return nil
	}
	if n == 0 {
		return nil
	}

	ack, _ := proto.ACKDes(buf)
	println("receive ack: ", ack.Id)
	return &ack
}

func registerStopSignal(packetId string, stopSignal chan int) {
	stopSignalMapMutex.Lock()
	stopSignalMap[packetId] = stopSignal
	stopSignalMapMutex.Unlock()
}

var stoppedCount int32 = 0
var sendCount int32 = 0
var resendGoCount int32 = 0
var seq = 0


func waitingDaemon(u *FUDPConn) {
	println("Waiting echo...")
	for {
		ack := u.readFromRemote()
		if ack == nil {
			fmt.Printf("Noting receive from server.")
			return
		}

		stopSignalMapMutex.Lock()
		if _, ok := stopSignalMap[ack.Id]; ok {
			stopSignal := stopSignalMap[ack.Id]
			stopSignal <- 1
		} else {
			fmt.Printf("[WARN] receive an unregister signal ackId %s\n", ack.Id)
			fmt.Printf("len map: %d\n", len(stopSignalMap))
		}
		stopSignalMapMutex.Unlock()
	}
}

func (u*FUDPConn) sendPacketWaitACK(packet proto.Proto) (err error) {
	u.sendPacket(packet)
	atomic.AddInt32(&sendCount, 1)

	stopSignal := make(chan int)
	registerStopSignal(packet.Id, stopSignal)

	const sendAckTimeout = 3000

	go func() {
		continueSignal := make(chan int)
		// 必须使用拷贝packet，否则会出现报文Data字段污染问题
		copyPacket := packet
		stopSignalCp := stopSignal

		for {
			go func() {
				time.Sleep(sendAckTimeout * time.Millisecond)
				continueSignal <- 1
			}()

			select {
			case <- stopSignalCp:
				atomic.AddInt32(&stoppedCount, 1)
				println("STOPPED count: ", stoppedCount)
				println("send count: ", sendCount)
				println("resend go count: ", resendGoCount)
				println("main seq: ", seq)
				return
			case <- continueSignal:
				atomic.AddInt32(&resendGoCount, 1)
				fmt.Printf("CONTINUED RESEND of packet: %s, seq: %d\n", copyPacket.Id, copyPacket.Seq)
				u.sendPacket(copyPacket)
			}
		}
	}()
	println("#send with ack done")
	return
}

func sendFile(host string, file string) (n int, err error) {
	conn := fdial(host)

	//go func() {
	//	waitingDaemon(&conn)
	//}()

	pfile, e := os.Open(file)
	if e != nil {
		fmt.Printf("open file %s error: %s\n", file, e.Error())
		return
	}
	var buffer = make([]byte, 4000)

	ncount := 0
	var windowCount = 0
	var window = 3000
	var windowDelay = 3000 * time.Millisecond
	var seek int64 = 0
	for {
		nRead, err  := pfile.Read(buffer[:])

		if err != nil && err != io.EOF {
			println("read full err:", err.Error())
		}

		conn.sendDataN(buffer, seq, nRead, seek)

		seq += 1
		seek += int64(nRead)

		fmt.Printf("send seq: %d, seek: %d\n", seq, seek)

		ncount += nRead
		windowCount += 1
		if windowCount >= window {
			fmt.Printf("Waiting window delay %d ms\n", windowDelay)
			time.Sleep(windowDelay)
			windowCount = 0
		}

		if nRead == 0 || err == io.EOF {
			break
		}
	}
	return
}

func main() {
	file := os.Args[1]
	host := os.Args[2]

	_, err := sendFile(host, file)
	if err != nil {
		return
	}
	<-fgStopSignal
}

