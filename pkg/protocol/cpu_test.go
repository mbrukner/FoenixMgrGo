package protocol

import (
	"fmt"
	"testing"
)

type cpuCommandTestConnection struct {
	response []byte
	requests [][]byte
}

func (c *cpuCommandTestConnection) Open(string) error { return nil }
func (c *cpuCommandTestConnection) Close() error      { return nil }
func (c *cpuCommandTestConnection) IsOpen() bool      { return true }

func (c *cpuCommandTestConnection) Write(data []byte) (int, error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("unexpected request length %d", len(data))
	}
	c.requests = append(c.requests, append([]byte(nil), data...))
	c.response = append(c.response, ResponseSyncByte, 0, 0, 0)
	return len(data), nil
}

func (c *cpuCommandTestConnection) Read(n int) ([]byte, error) {
	if len(c.response) < n {
		return nil, fmt.Errorf("requested %d bytes with only %d available", n, len(c.response))
	}
	data := append([]byte(nil), c.response[:n]...)
	c.response = c.response[n:]
	return data, nil
}

func TestPreservingCPUCommandsDoNotUseResetCommands(t *testing.T) {
	conn := &cpuCommandTestConnection{}
	dp := NewDebugPort(conn, nil)

	if err := dp.StopCPU(); err != nil {
		t.Fatalf("StopCPU returned an error: %v", err)
	}
	if err := dp.StartCPU(); err != nil {
		t.Fatalf("StartCPU returned an error: %v", err)
	}

	if len(conn.requests) != 2 {
		t.Fatalf("CPU stop/start sent %d requests, want 2", len(conn.requests))
	}
	if got := conn.requests[0][1]; got != CMDStopCPU || got == CMDEnterDebug {
		t.Errorf("StopCPU command = 0x%02X, want non-resetting 0x%02X", got, CMDStopCPU)
	}
	if got := conn.requests[1][1]; got != CMDStartCPU || got == CMDExitDebug {
		t.Errorf("StartCPU command = 0x%02X, want non-resetting 0x%02X", got, CMDStartCPU)
	}
}
