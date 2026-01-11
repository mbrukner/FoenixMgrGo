package connection

import (
	"fmt"
	"time"

	"github.com/daschewie/foenixmgr/pkg/config"
	"go.bug.st/serial"
)

// SerialConnection implements Connection interface for serial port communication
type SerialConnection struct {
	port   serial.Port
	config *config.Config
}

// NewSerialConnection creates a new serial connection with the given configuration
func NewSerialConnection(cfg *config.Config) *SerialConnection {
	return &SerialConnection{
		config: cfg,
	}
}

// Open establishes a serial connection to the specified port
func (s *SerialConnection) Open(portName string) error {
	if s.config == nil {
		// Load default config if not provided
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		s.config = cfg
	}

	mode := &serial.Mode{
		BaudRate: s.config.DataRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	// Attempt to open the port
	port, err := serial.Open(portName, mode)
	if err != nil {
		// Try to close and reopen (matching Python behavior)
		if port != nil {
			port.Close()
		}
		port, err = serial.Open(portName, mode)
		if err != nil {
			return fmt.Errorf("failed to open serial port %s: %w", portName, err)
		}
	}

	// Set timeout
	timeout := time.Duration(s.config.Timeout) * time.Second
	if err := port.SetReadTimeout(timeout); err != nil {
		port.Close()
		return fmt.Errorf("failed to set read timeout: %w", err)
	}

	s.port = port
	return nil
}

// Close closes the serial connection
func (s *SerialConnection) Close() error {
	if s.port == nil {
		return nil
	}
	return s.port.Close()
}

// IsOpen returns true if the connection is currently open
func (s *SerialConnection) IsOpen() bool {
	return s.port != nil
}

// Read reads exactly n bytes from the serial port
func (s *SerialConnection) Read(n int) ([]byte, error) {
	if s.port == nil {
		return nil, fmt.Errorf("serial port not open")
	}

	buf := make([]byte, n)
	totalRead := 0

	for totalRead < n {
		bytesRead, err := s.port.Read(buf[totalRead:])
		if err != nil {
			return nil, fmt.Errorf("serial read error: %w", err)
		}
		if bytesRead == 0 {
			return nil, fmt.Errorf("serial read timeout (expected %d bytes, got %d)", n, totalRead)
		}
		totalRead += bytesRead
	}

	return buf, nil
}

// Write writes all data to the serial port
func (s *SerialConnection) Write(data []byte) (int, error) {
	if s.port == nil {
		return 0, fmt.Errorf("serial port not open")
	}

	totalWritten := 0
	for totalWritten < len(data) {
		n, err := s.port.Write(data[totalWritten:])
		if err != nil {
			return totalWritten, fmt.Errorf("serial write error: %w", err)
		}
		totalWritten += n
	}

	return totalWritten, nil
}

// SetConfig updates the configuration for this connection
func (s *SerialConnection) SetConfig(cfg *config.Config) {
	s.config = cfg
}
