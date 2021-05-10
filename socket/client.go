package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// client wraps the whole functionality of a UDP client that sends
// a message and waits for a response coming back from the server
// that it initially targetted.
func client(address string, reader io.Reader) (err error) {
	// Resolve the UDP address so that we can make use of DialUDP
	// with an actual IP and port instead of a name (in case a
	// hostname is specified).
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}

	var maxBufferSize = 10000

	// Although we're not in a connection-oriented transport,
	// the act of `dialing` is analogous to the act of performing
	// a `connect(2)` syscall for a socket of type SOCK_DGRAM:
	// - it forces the underlying socket to only read and write
	//   to and from a specific remote address.
	conn, err := net.DialUDP("udp", nil, raddr)
	fmt.Printf("%s\n", conn.LocalAddr().String())
	writen, _ := conn.Write([]byte("hello"))
	fmt.Printf("Writed %d\n", writen)

	if err != nil {
		return
	}

	// Closes the underlying file descriptor associated with the,
	// socket so that it no longer refers to any file.
	defer conn.Close()

	doneChan := make(chan error, 1)


	go func() {
		fmt.Printf("#1\n")
		// It is possible that this action blocks, although this
		// should only occur in very resource-intensive situations:
		// - when you've filled up the socket buffer and the OS
		//   can't dequeue the queue fast enough.
		n, err := io.Copy(conn, reader)
		if err != nil {
			doneChan <- err
			return
		}
		fmt.Printf("#2\n")

		fmt.Printf("packet-written: bytes=%d\n", n)

		fmt.Printf("#3\n")
		buffer := make([]byte, maxBufferSize)

		// Set a deadline for the ReadOperation so that we don't
		// wait forever for a server that might not respond on
		// a resonable amount of time.
		fmt.Printf("#4\n")
		var timeout = 3000 * time.Millisecond
		deadline := time.Now().Add(timeout)
		err = conn.SetReadDeadline(deadline)
		if err != nil {
			doneChan <- err
			return
		}

		fmt.Printf("#5\n")
		nRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			doneChan <- err
			return
		}

		fmt.Printf("#6\n")
		fmt.Printf("packet-received: bytes=%d from=%s\n",
			nRead, addr.String())

		fmt.Printf("%d\n", string(buffer))

		doneChan <- nil
	}()

	//select {
	//case <-ctx.Done():
	//	fmt.Println("cancelled")
	//	err = ctx.Err()
	//case err = <-doneChan:
	//}

	return
}

func Write() {
	var addr = "localhost:1988"
	// Perform the address resolution and also
	// specialize the socket to only be able
	// to read and write to and from such
	// resolved address.
	conn, err := net.Dial("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Call the `Write()` method of the implementor
	// of the `io.Writer` interface.
	n, err := fmt.Fprintf(conn, "something")
	fmt.Println(n)
}

// maxBufferSize specifies the size of the buffers that
// are used to temporarily hold data from the UDP packets
// that we receive.
const maxBufferSize = 1024

// server wraps all the UDP echo server functionality.
// ps.: the server is capable of answering to a single
// client at a time.
func server(address string) (err error) {

	// Given that waiting for packets to arrive is blocking by nature and we want
	// to be able of canceling such action if desired, we do that in a separate
	// go routine.
	go func() {
		for {
			// ListenPacket provides us a wrapper around ListenUDP so that
			// we don't need to call `net.ResolveUDPAddr` and then subsequentially
			// perform a `ListenUDP` with the UDP address.
			//
			// The returned value (PacketConn) is pretty much the same as the one
			// from ListenUDP (UDPConn) - the only difference is that `Packet*`
			// methods and interfaces are more broad, also covering `ip`.
			pc, errL := net.ListenPacket("udp", address)
			if errL != nil {
				return
			}
			fmt.Printf("%s\n", pc.LocalAddr().String())

			// `Close`ing the packet "connection" means cleaning the data structures
			// allocated for holding information about the listening socket.
			//defer pc.Close()

			doneChan := make(chan error, 1)
			buffer := make([]byte, maxBufferSize)

			fmt.Printf("111\n")
			// By reading from the connection into the buffer, we block until there's
			// new content in the socket that we're listening for new packets.
			//
			// Whenever new packets arrive, `buffer` gets filled and we can continue
			// the execution.
			//
			// note.: `buffer` is not being reset between runs.
			//	  It's expected that only `n` reads are read from it whenever
			//	  inspecting its contents.
			err := pc.SetDeadline(time.Now().Add(2000 * time.Millisecond))
			if err != nil {
				fmt.Printf("###%s\n", err.Error())
				return 
			}
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}

			fmt.Printf("2\n")
			fmt.Printf("packet-received: bytes=%d from=%s\n",
				n, addr.String())

			// Setting a deadline for the `write` operation allows us to not block
			// for longer than a specific timeout.
			//
			// In the case of a write operation, that'd mean waiting for the send
			// queue to be freed enough so that we are able to proceed.
			var timeout = 500 * time.Millisecond
			deadline := time.Now().Add(timeout)
			err = pc.SetWriteDeadline(deadline)
			if err != nil {
				doneChan <- err
				return
			}

			fmt.Printf("3\n")
			// Write the packet's contents back to the client.
			n, err = pc.WriteTo(buffer[:n], addr)
			if err != nil {
				doneChan <- err
				return
			}

			fmt.Printf("packet-written: bytes=%d to=%s\n", n, addr.String())
		}
	}()

	return
}

func main() {

	var err = server("0.0.0.0:19881")
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		return
	}

	err = client("localhost:19881", bufio.NewReader(os.Stdin))
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		return
	}

	Write()

	var wg = sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
