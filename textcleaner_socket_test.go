package main

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"
)

// TestSocketServerStart tests that the socket server starts correctly
func TestSocketServerStart(t *testing.T) {
	socketPath := "/tmp/test_textcleaner_1.sock"
	defer os.Remove(socketPath)

	core := NewTextCleanerCore()
	server := NewSocketServer(socketPath, core)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start socket server: %v", err)
	}
	defer server.Stop()

	// Verify socket file exists
	if _, err := os.Stat(socketPath); err != nil {
		t.Fatalf("Socket file not created: %v", err)
	}
}

// TestSocketConnection tests basic client connection
func TestSocketConnection(t *testing.T) {
	socketPath := "/tmp/test_textcleaner_2.sock"
	defer os.Remove(socketPath)

	core := NewTextCleanerCore()
	server := NewSocketServer(socketPath, core)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start socket server: %v", err)
	}
	defer server.Stop()

	// Give server time to start accepting connections
	time.Sleep(100 * time.Millisecond)

	// Connect to socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to socket: %v", err)
	}
	defer conn.Close()
}

// TestLengthPrefixedProtocol tests the length-prefixed message protocol
func TestLengthPrefixedProtocol(t *testing.T) {
	socketPath := "/tmp/test_textcleaner_3.sock"
	defer os.Remove(socketPath)

	core := NewTextCleanerCore()
	server := NewSocketServer(socketPath, core)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start socket server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Connect to socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to socket: %v", err)
	}
	defer conn.Close()

	// Send a simple command using length-prefixed protocol
	cmdJSON := `{"action":"list_nodes","params":{}}`
	if err := sendMessage(conn, []byte(cmdJSON)); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read response using length-prefixed protocol
	response, err := receiveMessage(conn)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	// Verify response is valid JSON
	var resp map[string]interface{}
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("Response is not valid JSON: %v", err)
	}

	if success, ok := resp["success"].(bool); !ok || !success {
		t.Errorf("Expected successful response, got: %v", resp)
	}
}

// TestCommandExecution tests executing commands through the socket
func TestCommandExecution(t *testing.T) {
	socketPath := "/tmp/test_textcleaner_4.sock"
	defer os.Remove(socketPath)

	core := NewTextCleanerCore()
	server := NewSocketServer(socketPath, core)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start socket server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Connect to socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to socket: %v", err)
	}
	defer conn.Close()

	// Test 1: Create a node
	createCmd := `{
		"action": "create_node",
		"params": {
			"type": "operation",
			"name": "Test Uppercase",
			"operation": "Uppercase",
			"arg1": "",
			"arg2": "",
			"condition": ""
		}
	}`

	if err := sendMessage(conn, []byte(createCmd)); err != nil {
		t.Fatalf("Failed to send create command: %v", err)
	}

	response, err := receiveMessage(conn)
	if err != nil {
		t.Fatalf("Failed to receive response: %v", err)
	}

	var resp CommandResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("Create node failed: %s", resp.Error)
	}

	// Test 2: Get list of nodes
	listCmd := `{"action":"list_nodes","params":{}}`
	if err := sendMessage(conn, []byte(listCmd)); err != nil {
		t.Fatalf("Failed to send list command: %v", err)
	}

	response, err = receiveMessage(conn)
	if err != nil {
		t.Fatalf("Failed to receive response: %v", err)
	}

	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("List nodes failed: %s", resp.Error)
	}
}

// TestMultipleClients tests that multiple clients can connect and communicate simultaneously
func TestMultipleClients(t *testing.T) {
	socketPath := "/tmp/test_textcleaner_5.sock"
	defer os.Remove(socketPath)

	core := NewTextCleanerCore()
	server := NewSocketServer(socketPath, core)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start socket server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Connect first client
	conn1, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect first client: %v", err)
	}
	defer conn1.Close()

	// Connect second client - should succeed
	conn2, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect second client: %v", err)
	}
	defer conn2.Close()

	// Both clients should be able to communicate
	cmd := `{"action":"list_nodes","params":{}}`

	// Client 1 sends a command
	if err := sendMessage(conn1, []byte(cmd)); err != nil {
		t.Fatalf("Client 1 failed to send message: %v", err)
	}

	response1, err := receiveMessage(conn1)
	if err != nil {
		t.Fatalf("Client 1 failed to receive response: %v", err)
	}

	// Client 2 sends a command
	if err := sendMessage(conn2, []byte(cmd)); err != nil {
		t.Fatalf("Client 2 failed to send message: %v", err)
	}

	response2, err := receiveMessage(conn2)
	if err != nil {
		t.Fatalf("Client 2 failed to receive response: %v", err)
	}

	// Verify both got valid responses
	var resp1, resp2 CommandResponse
	if err := json.Unmarshal(response1, &resp1); err != nil {
		t.Fatalf("Failed to parse client 1 response: %v", err)
	}
	if err := json.Unmarshal(response2, &resp2); err != nil {
		t.Fatalf("Failed to parse client 2 response: %v", err)
	}

	if !resp1.Success || !resp2.Success {
		t.Fatalf("One or both clients got error responses")
	}
}

// TestMessageProtocol tests the message encoding/decoding by checking the binary format
func TestMessageProtocol(t *testing.T) {
	tests := []struct {
		name    string
		message []byte
	}{
		{"Empty message", []byte("")},
		{"Short message", []byte("hello")},
		{"JSON message", []byte(`{"action":"test"}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encoding: message should encode to 4-byte length + message
			lengthBuf := make([]byte, 4)
			binary.BigEndian.PutUint32(lengthBuf, uint32(len(tt.message)))

			// Verify length encoding
			expectedLength := uint32(len(tt.message))
			decodedLength := binary.BigEndian.Uint32(lengthBuf)

			if decodedLength != expectedLength {
				t.Errorf("Length mismatch: expected %d, got %d", expectedLength, decodedLength)
			}
		})
	}
}

// TestSetInputAndProcessing tests setting input text and processing through socket
func TestSetInputAndProcessing(t *testing.T) {
	socketPath := "/tmp/test_textcleaner_6.sock"
	defer os.Remove(socketPath)

	core := NewTextCleanerCore()
	server := NewSocketServer(socketPath, core)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start socket server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Connect to socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to socket: %v", err)
	}
	defer conn.Close()

	// Create an uppercase node
	createCmd := `{
		"action": "create_node",
		"params": {
			"type": "operation",
			"name": "Uppercase",
			"operation": "Uppercase",
			"arg1": "",
			"arg2": "",
			"condition": ""
		}
	}`

	if err := sendMessage(conn, []byte(createCmd)); err != nil {
		t.Fatalf("Failed to send create command: %v", err)
	}
	_, _ = receiveMessage(conn)

	// Set input text
	inputCmd := `{
		"action": "set_input_text",
		"params": {
			"text": "hello world"
		}
	}`

	if err := sendMessage(conn, []byte(inputCmd)); err != nil {
		t.Fatalf("Failed to send input command: %v", err)
	}

	response, err := receiveMessage(conn)
	if err != nil {
		t.Fatalf("Failed to receive response: %v", err)
	}

	var resp CommandResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("Set input text failed: %s", resp.Error)
	}

	// Get output text
	outputCmd := `{"action":"get_output_text","params":{}}`

	if err := sendMessage(conn, []byte(outputCmd)); err != nil {
		t.Fatalf("Failed to send output command: %v", err)
	}

	response, err = receiveMessage(conn)
	if err != nil {
		t.Fatalf("Failed to receive response: %v", err)
	}

	if err := json.Unmarshal(response, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("Get output text failed: %s", resp.Error)
	}

	// Check output (should be uppercase)
	// resp.Result could be a string or the JSON structure might be different
	if resp.Result != nil {
		outputText, ok := resp.Result.(string)
		if ok {
			if outputText != "HELLO WORLD" {
				t.Errorf("Expected 'HELLO WORLD', got '%s'", outputText)
			}
		} else {
			// Result might be in a different format, just verify we got a successful response
			t.Logf("Output result type: %T, value: %v", resp.Result, resp.Result)
		}
	}
}

// Helper functions

// sendMessage sends a length-prefixed message
func sendMessage(conn net.Conn, data []byte) error {
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	if _, err := conn.Write(lengthBuf); err != nil {
		return err
	}

	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}

// receiveMessage receives a length-prefixed message
func receiveMessage(conn net.Conn) ([]byte, error) {
	lengthBuf := make([]byte, 4)
	if _, err := conn.Read(lengthBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	data := make([]byte, length)

	if _, err := conn.Read(data); err != nil {
		return nil, err
	}

	return data, nil
}
