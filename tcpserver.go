/*
 *  Copyright (C) 2015 Pascal Bodin
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

// We do not accept frames with more than maxDataLength bytes of data.
// If such a frame is received, associated connection is closed.
const maxDataLength = 256
// Port on which we will wait for TCP connection requests.
const inPort = 50000

func main() {
	// Listen on TCP port on all interfaces.
	fmt.Println("Port: " + strconv.Itoa(inPort))
	l, err := net.Listen("tcp", ":" + strconv.Itoa(inPort))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Waiting for connections...")
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
			fmt.Println("wait_length_LSB")
			dataLength = int(b)
			currentState = wait_length_MSB
		case wait_length_MSB:
			fmt.Println("wait_length_MSB")
			dataLength = dataLength + int(b)*256
			if dataLength >= maxDataLength {
				cs = currentState
				dl = dataLength
				rm = nil
				// Connection will be closed by caller.
				err = errors.New("Error: data length too large")
				return cs, dl, rm, err
			}
			if dataLength == 0 {
				// Empty frame: no data.
				fmt.Println("Empty frame received.")
				currentState = wait_length_LSB
			} else {
				rm = nil
				currentState = wait_data
			}
		case wait_data:
			fmt.Println("wait_data")
			rm = append(rm, b)
			if len(rm) >= dataLength {
				fmt.Println("Message received.")
				// TODO: call message processing here.
				currentState = wait_length_LSB
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
