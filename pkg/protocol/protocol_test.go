package protocol

import (
	"encoding/binary"
	"fmt"
	"testing"
)

type readTestConnection struct {
	response []byte
	requests [][]byte
}

func (c *readTestConnection) Open(string) error { return nil }
func (c *readTestConnection) Close() error      { return nil }
func (c *readTestConnection) IsOpen() bool      { return true }

func (c *readTestConnection) Write(data []byte) (int, error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("unexpected request length %d", len(data))
	}

	request := append([]byte(nil), data...)
	c.requests = append(c.requests, request)

	length := binary.BigEndian.Uint16(request[5:7])
	fill := byte(len(c.requests))
	c.response = append(c.response, ResponseSyncByte, 0, 0)
	for range int(length) {
		c.response = append(c.response, fill)
	}
	c.response = append(c.response, 0)

	return len(data), nil
}

func (c *readTestConnection) Read(n int) ([]byte, error) {
	if len(c.response) < n {
		return nil, fmt.Errorf("requested %d bytes with only %d available", n, len(c.response))
	}

	data := append([]byte(nil), c.response[:n]...)
	c.response = c.response[n:]
	return data, nil
}

func TestReadBlockSplitsLargeReads(t *testing.T) {
	conn := &readTestConnection{}
	dp := NewDebugPort(conn, nil)

	const (
		address = uint32(0x120000)
		length  = uint32(0x800)
	)

	data, err := dp.ReadBlock(address, length)
	if err != nil {
		t.Fatalf("ReadBlock returned an error: %v", err)
	}

	if len(data) != int(length) {
		t.Fatalf("ReadBlock returned %d bytes, want %d", len(data), length)
	}
	if len(conn.requests) != 2 {
		t.Fatalf("ReadBlock sent %d requests, want 2", len(conn.requests))
	}

	for i, request := range conn.requests {
		gotAddress := uint32(request[2])<<16 | uint32(request[3])<<8 | uint32(request[4])
		wantAddress := address + uint32(i)*maxReadTransactionSize
		if gotAddress != wantAddress {
			t.Errorf("request %d address = 0x%06X, want 0x%06X", i, gotAddress, wantAddress)
		}

		gotLength := binary.BigEndian.Uint16(request[5:7])
		if gotLength != uint16(maxReadTransactionSize) {
			t.Errorf("request %d length = %d, want %d", i, gotLength, maxReadTransactionSize)
		}
	}

	if data[0] != 1 || data[maxReadTransactionSize] != 2 {
		t.Errorf("chunk data was not assembled in order")
	}
}

func TestReadBlockHandlesFinalPartialChunk(t *testing.T) {
	conn := &readTestConnection{}
	dp := NewDebugPort(conn, nil)

	data, err := dp.ReadBlock(0x2000, maxReadTransactionSize+1)
	if err != nil {
		t.Fatalf("ReadBlock returned an error: %v", err)
	}

	if len(data) != int(maxReadTransactionSize+1) {
		t.Fatalf("ReadBlock returned %d bytes, want %d", len(data), maxReadTransactionSize+1)
	}
	if len(conn.requests) != 2 {
		t.Fatalf("ReadBlock sent %d requests, want 2", len(conn.requests))
	}

	lastLength := binary.BigEndian.Uint16(conn.requests[1][5:7])
	if lastLength != 1 {
		t.Errorf("final request length = %d, want 1", lastLength)
	}
}

func TestReadBlockRejectsRangeBeyondAddressSpace(t *testing.T) {
	conn := &readTestConnection{}
	dp := NewDebugPort(conn, nil)

	_, err := dp.ReadBlock(maxProtocolAddress, 2)
	if err == nil {
		t.Fatal("ReadBlock returned no error for a range beyond the address space")
	}
	if len(conn.requests) != 0 {
		t.Errorf("ReadBlock sent %d requests for an invalid range, want 0", len(conn.requests))
	}
}
