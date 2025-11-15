// +build ignore

package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	socketPath := flag.String("socket", "/tmp/textcleaner.sock", "Path to the socket file")
	flag.Parse()

	client, err := NewSocketClient(*socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Printf("Connected to TextCleaner socket at %s\n", *socketPath)
	fmt.Println("Type 'help' for available commands, 'quit' to exit")
	fmt.Println()

	// Read commands from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if line == "quit" {
			fmt.Println("Disconnecting...")
			break
		}

		if line == "help" {
			printHelp()
			continue
		}

		// Try to parse as JSON
		var cmd map[string]interface{}
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			fmt.Printf("Invalid JSON: %v\n", err)
			fmt.Println("Type 'help' for examples")
			continue
		}

		// Send command and get response
		response, err := client.Execute(line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Pretty print response
		var resp interface{}
		if err := json.Unmarshal([]byte(response), &resp); err != nil {
			fmt.Printf("Response: %s\n", response)
		} else {
			data, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Printf("Response: %s\n", string(data))
		}
		fmt.Println()
	}
}

// SocketClient manages connection to the socket server
type SocketClient struct {
	conn net.Conn
}

// NewSocketClient creates a new socket client
func NewSocketClient(socketPath string) (*SocketClient, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to socket: %w", err)
	}

	return &SocketClient{conn: conn}, nil
}

// Execute sends a command and returns the response
func (sc *SocketClient) Execute(cmdJSON string) (string, error) {
	// Send command
	if err := sc.sendMessage([]byte(cmdJSON)); err != nil {
		return "", err
	}

	// Receive response
	response, err := sc.receiveMessage()
	if err != nil {
		return "", err
	}

	return string(response), nil
}

// sendMessage sends a length-prefixed message
func (sc *SocketClient) sendMessage(data []byte) error {
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	if _, err := sc.conn.Write(lengthBuf); err != nil {
		return err
	}

	if _, err := sc.conn.Write(data); err != nil {
		return err
	}

	return nil
}

// receiveMessage receives a length-prefixed message
func (sc *SocketClient) receiveMessage() ([]byte, error) {
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(sc.conn, lengthBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	data := make([]byte, length)

	if _, err := io.ReadFull(sc.conn, data); err != nil {
		return nil, err
	}

	return data, nil
}

// Close closes the connection
func (sc *SocketClient) Close() error {
	return sc.conn.Close()
}

// printHelp prints available commands
func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println()
	fmt.Println("SESSION STATE COMMANDS (useful for session persistence):")
	fmt.Println("1. Get input text:")
	fmt.Println(`   {"action":"get_input_text","params":{}}`)
	fmt.Println()
	fmt.Println("2. Get output text:")
	fmt.Println(`   {"action":"get_output_text","params":{}}`)
	fmt.Println()
	fmt.Println("3. Get selected node ID:")
	fmt.Println(`   {"action":"get_selected_node_id","params":{}}`)
	fmt.Println()
	fmt.Println("PIPELINE MANAGEMENT COMMANDS:")
	fmt.Println("4. List all root nodes:")
	fmt.Println(`   {"action":"list_nodes","params":{}}`)
	fmt.Println()
	fmt.Println("5. Create a node:")
	fmt.Println(`   {"action":"create_node","params":{"type":"operation","name":"Uppercase","operation":"Uppercase","arg1":"","arg2":"","condition":""}}`)
	fmt.Println()
	fmt.Println("6. Update a node:")
	fmt.Println(`   {"action":"update_node","params":{"node_id":"node_0","type":"operation","name":"Uppercase","operation":"Uppercase","arg1":"","arg2":"","condition":""}}`)
	fmt.Println()
	fmt.Println("7. Delete a node:")
	fmt.Println(`   {"action":"delete_node","params":{"node_id":"node_0"}}`)
	fmt.Println()
	fmt.Println("8. Add child node:")
	fmt.Println(`   {"action":"add_child_node","params":{"parent_id":"node_0","type":"operation","name":"Lowercase","operation":"Lowercase","arg1":"","arg2":"","condition":""}}`)
	fmt.Println()
	fmt.Println("9. Select a node:")
	fmt.Println(`   {"action":"select_node","params":{"node_id":"node_0"}}`)
	fmt.Println()
	fmt.Println("10. Get a specific node:")
	fmt.Println(`   {"action":"get_node","params":{"node_id":"node_0"}}`)
	fmt.Println()
	fmt.Println("TEXT PROCESSING COMMANDS:")
	fmt.Println("11. Set input text:")
	fmt.Println(`   {"action":"set_input_text","params":{"text":"hello world"}}`)
	fmt.Println()
	fmt.Println("12. Get pipeline:")
	fmt.Println(`   {"action":"get_pipeline","params":{}}`)
	fmt.Println()
	fmt.Println("IMPORT/EXPORT COMMANDS:")
	fmt.Println("13. Export pipeline:")
	fmt.Println(`   {"action":"export_pipeline","params":{}}`)
	fmt.Println()
	fmt.Println("14. Import pipeline:")
	fmt.Println(`   {"action":"import_pipeline","params":{"json":"[...]"}}`)
	fmt.Println()
	fmt.Println("Type 'help' again to see this message, or 'quit' to exit")
	fmt.Println()
}
