package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

// We do not accept frames with more than maxDataLength bytes of data.
// If such a frame is received, associated connection is closed.
const maxDataLength = 256

func main() {
	// Listen on TCP port 50000 on all interfaces.
	l, err := net.Listen("tcp", ":50000")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Waiting for connections on port 50000...")
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("New connection.")
		// Handle the connection in a new goroutine.
		go handleConn(conn)
	}
}

// handleConn() automaton states.
const (
	wait_length_LSB = iota
	wait_length_MSB
	wait_data
)

func handleConn(c net.Conn) {
	// Automaton state.
	var currentState int
	// Length of data in received frame.
	var dataLength int
	// Data in received frame.
	var recMessage []byte
	var errFrame error
	// Ensure read (and write) operations are blocking, with no timeout.
	c.SetDeadline(time.Time{})
	// Create receive buffer.
	recBuf := make([]byte, 4)
	// Initialize automaton state.
	currentState = wait_length_LSB
	// Close connection when exiting.
	defer c.Close()
	for {
		// Wait for data.
		recLen, errRead := c.Read(recBuf)
		if errRead != nil {
			if errRead.Error() == "EOF" {
				fmt.Println("Connection closed by remote side.")
				break
			} else {
				log.Fatal(errRead)
			}
		}
		fmt.Printf("Received %d bytes.\n", recLen)
		// Process received data.
		currentState, dataLength, recMessage, errFrame = processRecData(recBuf[:recLen], currentState, dataLength, recMessage)
		if errFrame != nil {
			fmt.Println(errFrame.Error() + " - closing connection.")
			break
		}
	}
}

// processRecData() implements a (simple) finite state automaton, to decode
// received frames.
func processRecData(data []byte, currentState int, dataLength int, recMessage []byte) (cs int, dl int, rm []byte, err error) {
	// Process every byte of received buffer.
	rm = recMessage
	for _, b := range data {
		fmt.Printf("%02X ", b)
		switch currentState {
		case wait_length_LSB:
			dataLength = int(b)
			currentState = wait_length_MSB
		case wait_length_MSB:
			dataLength = dataLength + int(b)*256
			if dataLength >= maxDataLength {
				cs = currentState
				dl = dataLength
				rm = nil
				err = errors.New("Error: data length too large")
				return cs, dl, rm, err
			}
			rm = nil
			currentState = wait_data
		case wait_data:
			// Data length 0 is supported.
			if len(rm) >= dataLength {
				fmt.Printf("Message received: ", rm)
				// TODO: call message processing here.
				currentState = wait_length_LSB
			} else {
				rm = append(rm, b)
			}
		default:
			log.Fatal(errors.New("Error: unknown state"))
		}
	}
	// At this stage, no error.
	cs = currentState
	dl = dataLength
	return cs, dl, rm, nil
}
