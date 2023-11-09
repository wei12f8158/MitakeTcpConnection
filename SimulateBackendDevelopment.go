package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
)

type TCMDHeader struct {
	HeadSpec [4]byte // File head identification CMDHEADER_HEADSPEC
	Secno    int32   // Serial number, Server will return to Client
	CMDLen   int32   // Data length
	CMDType  int32   // Command code
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Client connected:", conn.RemoteAddr().String())

	// Create a buffer to hold the incoming data
	buf := make([]byte, 4096)

	// Define Secno outside of the loop
	Secno := int32(0)

	for {
		// Read data from the connection into the buffer
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading data:", err)
			return
		}

		// The buffer may be larger than the amount of received data, trim the unused bytes
		data := buf[:n]

		// Print out the received data as a string
		fmt.Println("Received data:", string(data))

		// your command string
		var command []byte
		if strings.Contains(string(data), "W7003") {
			command, err = os.ReadFile("W7003.txt")
		} else {
			command, err = os.ReadFile("W7005.txt")
		}

		if err != nil {
			fmt.Println("Unable to read command file:", err)
			return
		}

		// increment the Secno
		Secno++

		// create the header with command length
		header := TCMDHeader{
			HeadSpec: [4]byte{'%', '9', '(', '@'}, // null terminated string
			Secno:    Secno,                       // incrementing serial number for each response
			CMDLen:   int32(len(command)),         // length of the command that follows the header
			CMDType:  110000,                      // put appropriate command code here
		}

		// create a byte slice with enough space for the header and the command
		response := make([]byte, binary.Size(header)+len(command))

		// write the header to the byte slice
		copy(response[:4], header.HeadSpec[:])
		binary.LittleEndian.PutUint32(response[4:], uint32(header.Secno))
		binary.LittleEndian.PutUint32(response[8:], uint32(header.CMDLen))
		binary.LittleEndian.PutUint32(response[12:], uint32(header.CMDType))

		// write the command to the byte slice
		copy(response[binary.Size(header):], command)

		// send the data
		_, err = conn.Write(response)
		if err != nil {
			fmt.Println("Error sending response:", err)
			return
		}
	}
}

func main() {

	// Start the TCP server on a specific address and port
	listener, err := net.Listen("tcp", "127.0.0.1:1234")

	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	fmt.Println("Server started, listening on", listener.Addr().String())

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle the connection in a separate goroutine
		go handleConnection(conn)
	}
}
