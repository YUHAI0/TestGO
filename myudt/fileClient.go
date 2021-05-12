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


type MyUDPConn struct {
	localAddr string
	remoteAddr string
	conn *net.UDPConn
}

func Dial(addr string) (conn MyUDPConn) {
	println("#1")
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return
	}

	newConn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}

	return MyUDPConn{newConn.LocalAddr().String(), newConn.RemoteAddr().String(), newConn}
}


func (u* MyUDPConn) sendData(data []byte, seq int, seek int64) error {
	println("#2")
	return u.sendPacketWaitACK(proto.NewProto(data, seq, seek))
}

func (u* MyUDPConn) sendDataN(data []byte, seq int, n int, seek int64) error {
	println("#2")
	if n <= 0 {
		return nil
	}
	return u.sendPacketWaitACK(proto.NewProto(data[:n], seq, seek))
}

func (u* MyUDPConn) sendPacket(packet proto.Proto) {
	buffer, _ := proto.Serialize(packet)

	//n, _ := io.Copy(u.conn, bytes.NewReader(buffer))
	_, _ = io.Copy(u.conn, bytes.NewReader(buffer))

	//fmt.Printf("packet-written: bytes=%d\n", n)
}

var toberesendqueue  []proto.Proto
//var packetSignalMap map[string] chan string
var stopSignalMapMutex = sync.RWMutex{}
var stopSignalMap map[string] chan int = make(map[string] chan int)

var daemonChan chan string = make(chan string)

var gStopSignal chan int = make(chan int)

func resendPacketControllerDaemon() {
	println("#4")
	/*
	var chans []chan string

	for _, v := range packetSignalMap {
		chans = append(chans, v)
	}
	if len(chans) == 0 {
		return
	}

	cases := make([]reflect.SelectCase, len(chans))
	for i, ch := range chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	_, value, _:= reflect.Select(cases)
	// ok will be true if the channel has not been closed.
	//_ := chans[chosen]
	packetId := value.String()
	*/

	//
	for {
		println("Waiting daemon...")
		var packetId = <-daemonChan
		println("Get daemonchan packetid: ", packetId)

		// send stop signal
		fmt.Printf("Send stop signal of packetId: %s\n", packetId)
		stopSignalMapMutex.Lock()
		stopSignalMap[packetId] <- 1
		stopSignalMapMutex.Unlock()
	}
}

func (u *MyUDPConn) readFromRemote() *proto.ACK {
	println("#5")
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
	println("#6")

	stopSignalMapMutex.Lock()
	stopSignalMap[packetId] = stopSignal
	stopSignalMapMutex.Unlock()
}

func waitEchoRoutine(conn* MyUDPConn) {
	println("#7")
	go func() {
		ack := conn.readFromRemote()
		if ack == nil {
			return
		}
		fmt.Printf("Receive remote %s ack: pid %s\n", conn.conn.RemoteAddr().String(), ack.Id)
		daemonChan <- ack.Id
	}()
}


var stoppedCount int32 = 0
var sendCount int32 = 0
var resendGoCount int32 = 0

var seq = 0

//var mapStopSignal map[string] chan int = make(map[string] chan int)

func waitingDaemon(u *MyUDPConn) {
	println("Waiting echo...")
	for {
		//t1 := time.Now().Unix()
		ack := u.readFromRemote()
		if ack == nil {
			fmt.Printf("Noting receive from server.")
			return
		}
		//t2 := time.Now().Unix()
		//fmt.Printf("t1 %d, t2 %d\n", t1, t2)
		//fmt.Printf("Receive during: %d sec, remote %s\n\t ack: pid %s\n", t2-t1, u.conn.RemoteAddr().String(), ack.Id)
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

func (u* MyUDPConn) sendPacketWaitACK(packet proto.Proto) (err error) {
	println("#8, send packet")
	u.sendPacket(packet)
	atomic.AddInt32(&sendCount, 1)

	//jj, _ := json.Marshal(packet)

	stopSignal := make(chan int)
	registerStopSignal(packet.Id, stopSignal)

	const sendAckTimeout = 3000

	go func() {
		continueSignal := make(chan int)
		// 必须使用拷贝packet，否则会出现报文Data字段污染问题
		//copyPacket := proto.NewProto(packet.Data, packet.Seq, packet.Seek)
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
					//fmt.Printf("CONTINUED RESEND of packet: %s, seq: %d\n", copyPacket.Id, copyPacket.Seq)
					//j, _ := json.Marshal(copyPacket)
					//fmt.Printf("resend packet: %s\n", j)
					//fmt.Printf("resend data %s\n", string(copyPacket.Data))
					u.sendPacket(copyPacket)
			}
		}
	}()

	println("#send with ack done")
	return
}

func sendFile(host string, file string) (n int, err error) {
	conn := Dial(host)

	go func() {
		waitingDaemon(&conn)
	}()

	pfile, e := os.Open(file)
	if e != nil {
		fmt.Printf("open file %s error: %s\n", file, e.Error())
		return
	}
	var buffer = make([]byte, 4000)

	ncount := 0
	var windowCount = 0
	var window = 100
	var windowDelay = 20 * time.Millisecond
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
			time.Sleep(windowDelay)
			windowCount = 0
		}

		if nRead == 0 || err == io.EOF {
			break
		}
	}
	return
}


//var wg = sync.WaitGroup{}

func main() {
	file := os.Args[1]
	host := os.Args[2]

	_, err := sendFile(host, file)
	if err != nil {
		return
	}

	//println("waiting group")
	//wg.Add(1)
	//wg.Wait()
	<- gStopSignal
}
