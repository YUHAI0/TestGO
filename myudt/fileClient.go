package main

import (
	"bytes"
	"fmt"
	"github.com/oxtoacart/go-udt/myudt/proto"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

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
		stopSignalMap[packetId] <- 1
	}
}

func (u *MyUDPConn) readFromRemote() proto.ACK {
	println("#5")
	var buf = make([]byte, 2000)
	_, _, err := u.conn.ReadFrom(buf)
	if err != nil {
		return proto.ACK{}
	}

	ack, _ := proto.ACKDes(buf)
	println("receive ack: ", ack.Id)
	return ack
}

func registerStopSignal(packetId string, stopSignal chan int) {
	println("#6")
	stopSignalMap[packetId] = stopSignal
}

func waitEchoRoutine(conn* MyUDPConn) {
	println("#7")
	go func() {
		ack := conn.readFromRemote()
		fmt.Printf("Receive remote %s ack: pid %s\n", conn.conn.RemoteAddr().String(), ack.Id)
		daemonChan <- ack.Id
	}()
}

const sendAckTimeout = 1000
func (u* MyUDPConn) sendPacketWaitACK(packet proto.Proto) (err error) {
	println("#8")
	packetId := u.sendPacket(packet)

	toberesendqueue = append(toberesendqueue, packet)

	stopSignal := make(chan int)
	resendSignal := make(chan int)

	registerStopSignal(packetId, stopSignal)

	go func() {
		fmt.Printf("Wait until timeout")
		time.Sleep(sendAckTimeout * time.Millisecond)
		resendSignal <- 1
	}()

	waitEchoRoutine(u)

	select {
		case <- stopSignal:
			println("stopped resend")
			return
		case <- resendSignal:
			println("resending...")
			serr := u.sendPacketWaitACK(packet)
			if serr != nil {
				return serr
			}
	}

	print("#9")
	return
}

func sendFile(host string, file string) (n int, err error) {
	conn := Dial(host)

	pfile, e := os.Open(file)
	if e != nil || pfile == nil {
		fmt.Printf("open file %s error: %s\n", file, e.Error())
		return
	}
	var buffer = make([]byte, 4000)

	ncount := 0
	for {
		print("111")
		nRead, err  := pfile.Read(buffer[:])
		fmt.Printf("read %d bytes\n", nRead)

		if err != nil && err != io.EOF {
			print("read full err:", err.Error())
		}

		conn.sendData(buffer)
		println("##########115")
		ncount += nRead

		fmt.Printf("packet-written: bytes=%d\n", n)

		if nRead == 0 || err == io.EOF {
			print("113")
			break
		}
		print("112")
	}
	return
}


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

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
