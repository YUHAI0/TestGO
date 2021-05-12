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

var mainSendCount = 0

/*
func oldsendFile(host string, file string) (n int, err error) {
	raddr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}

	defer conn.Close()

	pfile, e := os.Open(file)
	if e != nil || pfile == nil {
		fmt.Printf("open file %s error: %s\n", file, e.Error())
		return
	}

	var buffer = make([]byte, 4000)

	ncount := 0
	for {
		nRead, err  := pfile.Read(buffer[:])
		fmt.Printf("read %d bytes\n", nRead)

		if err != nil && err != io.EOF {
			print("read full err:", err.Error())
		}

		n, _ := io.CopyN(conn, bytes.NewReader(buffer), int64(nRead))
		ncount += int(n)

		fmt.Printf("packet-written: bytes=%d\n", n)

		if nRead == 0 || err == io.EOF {
			break
		}
	}

	return ncount, nil
}
*/

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


func (u* MyUDPConn) sendData(data []byte) error {
	println("#2")
	return u.sendPacketWaitACK(proto.NewProto(data))
}

func (u* MyUDPConn) sendPacket(packet proto.Proto) (packetId string) {
	fmt.Printf("#3, %s, %d, %s\n", packet.Version, packet.Length, packet.Id)
	buffer, _ := proto.Serialize(packet)

	n, _ := io.Copy(u.conn, bytes.NewReader(buffer))

	fmt.Printf("packet-written: bytes=%d\n", n)
	return packet.Id
}

var toberesendqueue  []proto.Proto
//var packetSignalMap map[string] chan string
var stopSignalMapMutex = sync.RWMutex{}
var stopSignalMap map[string] chan int = make(map[string] chan int)

var daemonChan chan string = make(chan string)

func myinit() {
	go resendPacketControllerDaemon()
}

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

const sendAckTimeout = 30000
//var resendMap map[string] proto.Proto = make(map[string] proto.Proto)
//var gConn *MyUDPConn = nil

var stoppedCount int32 = 0

var sendCount int32 = 0

var resendGoCount int32 = 0

func (u* MyUDPConn) sendPacketWaitACK(packet proto.Proto) (err error) {
	println("#8, send packet")
	packetId := u.sendPacket(packet)
	atomic.AddInt32(&sendCount, 1)

	// register tobe resend when timeout
	//resendMap[packetId] = packet

	toberesendqueue = append(toberesendqueue, packet)

	stopSignal := make(chan int)
	//resendSignal := make(chan int)

	registerStopSignal(packetId, stopSignal)

	//go func() {
	//	fmt.Printf("Wait until timeout\n")
	//	time.Sleep(sendAckTimeout * time.Millisecond)
	//	resendSignal <- 1
	//}()

	//waitEchoRoutine(u)

	println("Waiting echo...")
	go func() {
		t1 := time.Now().Unix()
		ack := u.readFromRemote()
		if ack == nil {
			return
		}
		t2 := time.Now().Unix()
		fmt.Printf("t1 %d, t2 %d\n", t1, t2)
		fmt.Printf("Receive during: %d sec, remote %s ack: pid %s\n", t2-t1,u.conn.RemoteAddr().String(), ack.Id)
		//daemonChan <- ack.Id
		stopSignal <- 1
	}()

	go func() {
		atomic.AddInt32(&resendGoCount, 1)
		continueSignal := make(chan int)
		for {
			go func() {
				time.Sleep(sendAckTimeout * time.Millisecond)
				continueSignal <- 1
			}()

			select {
				case <- stopSignal:
					atomic.AddInt32(&stoppedCount, 1)
					println("STOPPED count: ", stoppedCount)
					println("send count: ", sendCount)
					println("resend go count: ", resendGoCount)
					println("main send count", mainSendCount)
					return
				case <- continueSignal:
					println("CONTINUED")
			}
			u.sendPacket(packet)
		}
	}()

	//select { //	case <- stopSignal:
	//		println("stopped resend")
	//		return
	//	case <- resendSignal:
	//		println("resending...")
	//		serr := u.sendPacketWaitACK(packet)
	//		if serr != nil {
	//			return serr
	//		}
	//}

	println("#send with ack done")
	return
}

//var gConn = nil
func sendFile(host string, file string) (n int, err error) {
	conn := Dial(host)

	pfile, e := os.Open(file)
	if e != nil || pfile == nil {
		fmt.Printf("open file %s error: %s\n", file, e.Error())
		return }
	var buffer = make([]byte, 4000)

	ncount := 0
	var windowCount = 0
	var window = 100
	var windowDelay = 20 * time.Millisecond
	for {
		nRead, err  := pfile.Read(buffer[:])
		fmt.Printf("read %d bytes\n", nRead)

		if err != nil && err != io.EOF {
			print("read full err:", err.Error())
		}

		conn.sendData(buffer)
		mainSendCount += 1
		fmt.Printf("sendCount: %d\n", mainSendCount)
		ncount += nRead
		windowCount += 1
		if windowCount >= window {
			time.Sleep(windowDelay)
			windowCount = 0
		}

		fmt.Printf("packet-written: bytes=%d\n", n)

		if nRead == 0 || err == io.EOF {
			break
		}
	}
	return
}


var wg = sync.WaitGroup{}
func main() {
	file := os.Args[1]
	host := os.Args[2]

	go func() {
		resendPacketControllerDaemon()
	}()

	_, err := sendFile(host, file)
	if err != nil {
		return
	}

	println("waiting go")
	go func() {
		for {
			//for packetId,packet := range resendMap {
			//
			//}
		}
	}()

	println("waiting group")
	wg.Add(1)
	wg.Wait()
}
