package connection

import (
	"fmt"
	"io"
	"net"

	"go.bug.st/serial"
)

const (
	// Protocol command constants (from protocol package)
	cmdReadMem  = 0x00
	cmdWriteMem = 0x01
)

// Bridge represents a TCP-to-serial relay server
type Bridge struct {
	tcpHost    string
	tcpPort    int
	serialPort string
	baudRate   int
	timeout    int
}

// NewBridge creates a new TCP bridge
func NewBridge(tcpHost string, tcpPort int, serialPort string, baudRate int, timeout int) *Bridge {
	return &Bridge{
		tcpHost:    tcpHost,
		tcpPort:    tcpPort,
		serialPort: serialPort,
		baudRate:   baudRate,
		timeout:    timeout,
	}
}

// Listen starts the TCP server and relays messages to the serial port
func (b *Bridge) Listen() error {
	addr := fmt.Sprintf("%s:%d", b.tcpHost, b.tcpPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Listening for connections to %s on port %d\n", b.tcpHost, b.tcpPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		fmt.Printf("Received connection from %s\n", conn.RemoteAddr().String())

		// Handle connection in a goroutine to support multiple clients
		go b.handleConnection(conn)
	}
}

// handleConnection processes a single TCP connection
func (b *Bridge) handleConnection(tcpConn net.Conn) {
	defer tcpConn.Close()

	for {
		// Read 7-byte request header
		header := make([]byte, 7)
		if _, err := io.ReadFull(tcpConn, header); err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading header: %v\n", err)
			} else {
				fmt.Printf("Connection from %s closed\n", tcpConn.RemoteAddr().String())
			}
			return
		}

		command := header[1]

		// Extract data length from header bytes 5-6 (big-endian)
		dataLength := uint16(header[5])<<8 | uint16(header[6])

		// Read data payload if this is a write command
		var data []byte
		if command == cmdWriteMem {
			data = make([]byte, dataLength)
			if _, err := io.ReadFull(tcpConn, data); err != nil {
				fmt.Printf("Error reading data: %v\n", err)
				return
			}
		}

		// Read LRC byte
		lrcByte := make([]byte, 1)
		if _, err := io.ReadFull(tcpConn, lrcByte); err != nil {
			fmt.Printf("Error reading LRC: %v\n", err)
			return
		}

		// Construct complete request to send to serial port
		request := make([]byte, 0, len(header)+len(data)+1)
		request = append(request, header...)
		if data != nil {
			request = append(request, data...)
		}
		request = append(request, lrcByte...)

		// Open serial port for this transaction
		mode := &serial.Mode{
			BaudRate: b.baudRate,
		}
		serialConn, err := serial.Open(b.serialPort, mode)
		if err != nil {
			fmt.Printf("Error opening serial port: %v\n", err)
			return
		}

		// Send request to serial port
		numWritten, err := serialConn.Write(request)
		if err != nil {
			serialConn.Close()
			fmt.Printf("Error writing to serial port: %v\n", err)
			return
		}
		if numWritten != len(request) {
			serialConn.Close()
			fmt.Printf("Serial write error: wrote %d bytes, expected %d\n", numWritten, len(request))
			return
		}

		// Read response from serial port
		// 1. Sync byte
		responseSyncByte := make([]byte, 1)
		if _, err := io.ReadFull(serialConn, responseSyncByte); err != nil {
			serialConn.Close()
			fmt.Printf("Error reading response sync: %v\n", err)
			return
		}

		// 2. Two status bytes
		responseStatusBytes := make([]byte, 2)
		if _, err := io.ReadFull(serialConn, responseStatusBytes); err != nil {
			serialConn.Close()
			fmt.Printf("Error reading status bytes: %v\n", err)
			return
		}

		// 3. Data payload (only for read commands)
		var responseData []byte
		if command == cmdReadMem && dataLength > 0 {
			responseData = make([]byte, dataLength)
			if _, err := io.ReadFull(serialConn, responseData); err != nil {
				serialConn.Close()
				fmt.Printf("Error reading response data: %v\n", err)
				return
			}
		}

		// 4. Response LRC byte
		responseLrcByte := make([]byte, 1)
		if _, err := io.ReadFull(serialConn, responseLrcByte); err != nil {
			serialConn.Close()
			fmt.Printf("Error reading response LRC: %v\n", err)
			return
		}

		serialConn.Close()

		// Construct complete response
		response := make([]byte, 0, 1+2+len(responseData)+1)
		response = append(response, responseSyncByte...)
		response = append(response, responseStatusBytes...)
		if responseData != nil {
			response = append(response, responseData...)
		}
		response = append(response, responseLrcByte...)

		// Send response back to TCP client
		if _, err := tcpConn.Write(response); err != nil {
			fmt.Printf("Error writing response to TCP: %v\n", err)
			return
		}
	}
}
