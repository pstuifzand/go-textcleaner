package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// UpdateCallback is called when the core state changes via socket command
type UpdateCallback func()

// SocketClient allows GUI to connect to and query a running socket server
type SocketClient struct {
	conn net.Conn
}

// NewSocketClient connects to a running socket server
func NewSocketClient(socketPath string) (*SocketClient, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to socket server at %s: %w", socketPath, err)
	}

	return &SocketClient{conn: conn}, nil
}

// Close closes the connection to the socket server
func (sc *SocketClient) Close() error {
	if sc.conn != nil {
		return sc.conn.Close()
	}
	return nil
}

// Execute sends a command and returns the response
func (sc *SocketClient) Execute(cmdJSON string) (map[string]interface{}, error) {
	// Send command
	if err := sc.sendMessage([]byte(cmdJSON)); err != nil {
		return nil, err
	}

	// Receive response
	data, err := sc.receiveMessage()
	if err != nil {
		return nil, err
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response, nil
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

// SocketServer manages the Unix domain socket interface for TextCleanerCore
type SocketServer struct {
	socketPath string
	core       *TextCleanerCore
	listener   net.Listener
	mu         sync.Mutex
	done       chan struct{}
	stopped    chan struct{} // Closed when server has fully shut down
	callbacks  []UpdateCallback // Callbacks called after each command execution to update UIs
}

// NewSocketServer creates a new socket server instance
func NewSocketServer(socketPath string, core *TextCleanerCore) *SocketServer {
	return &SocketServer{
		socketPath: socketPath,
		core:       core,
		done:       make(chan struct{}),
		stopped:    make(chan struct{}),
		callbacks:  make([]UpdateCallback, 0),
	}
}

// SetUpdateCallback adds a callback to be called after each socket command
// This is used to notify all connected GUIs to refresh when socket commands modify the core
func (ss *SocketServer) SetUpdateCallback(callback UpdateCallback) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.callbacks = append(ss.callbacks, callback)
}

// Start begins listening on the Unix domain socket
func (ss *SocketServer) Start() error {
	// Remove existing socket file if it exists
	if err := os.Remove(ss.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Create Unix domain socket listener
	listener, err := net.Listen("unix", ss.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket %s: %w", ss.socketPath, err)
	}

	ss.listener = listener

	// Set up signal handling for graceful shutdown
	go ss.handleSignals()

	// Accept connections in a goroutine
	go ss.acceptConnections()

	return nil
}

// acceptConnections accepts incoming connections (multiple clients supported)
func (ss *SocketServer) acceptConnections() {
	for {
		select {
		case <-ss.done:
			return
		default:
		}

		conn, err := ss.listener.Accept()
		if err != nil {
			// Check if we're shutting down
			select {
			case <-ss.done:
				return
			default:
				fmt.Fprintf(os.Stderr, "Error accepting connection: %v\n", err)
				continue
			}
		}

		// Handle each client in a separate goroutine (allow multiple concurrent clients)
		go ss.handleClient(conn)
	}
}

// handleClient handles communication with a connected client
func (ss *SocketServer) handleClient(conn net.Conn) {
	defer conn.Close()

	reader := &lengthPrefixedReader{conn: conn}
	writer := &lengthPrefixedWriter{conn: conn}

	for {
		// Read JSON command
		data, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				// Client disconnected normally
				return
			}
			fmt.Fprintf(os.Stderr, "Error reading from client: %v\n", err)
			return
		}

		// Execute command through the core
		response := ss.core.ExecuteCommand(string(data))

		// Send JSON response
		if err := writer.Write([]byte(response)); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to client: %v\n", err)
			return
		}

		// Trigger all registered update callbacks (e.g., to refresh all GUIs)
		ss.mu.Lock()
		callbacks := append([]UpdateCallback{}, ss.callbacks...)
		ss.mu.Unlock()
		for _, callback := range callbacks {
			callback()
		}
	}
}

// handleSignals sets up graceful shutdown on signals
func (ss *SocketServer) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	ss.Stop()
}

// Stop gracefully shuts down the socket server
func (ss *SocketServer) Stop() error {
	close(ss.done)

	if ss.listener != nil {
		ss.listener.Close()
	}

	// Remove socket file
	os.Remove(ss.socketPath)

	// Signal that the server has stopped
	close(ss.stopped)

	return nil
}

// Wait blocks until the server is fully shut down
func (ss *SocketServer) Wait() {
	<-ss.stopped
}

// ============================================================================
// Length-Prefixed Protocol Implementation
// ============================================================================

// lengthPrefixedReader reads length-prefixed messages (4-byte big-endian length + data)
type lengthPrefixedReader struct {
	conn net.Conn
}

// Read reads a single length-prefixed message
func (r *lengthPrefixedReader) Read() ([]byte, error) {
	// Read 4-byte length prefix
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(r.conn, lengthBuf); err != nil {
		return nil, err
	}

	// Decode length
	length := binary.BigEndian.Uint32(lengthBuf)

	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(r.conn, data); err != nil {
		return nil, err
	}

	return data, nil
}

// lengthPrefixedWriter writes length-prefixed messages (4-byte big-endian length + data)
type lengthPrefixedWriter struct {
	conn net.Conn
}

// Write writes a single length-prefixed message
func (w *lengthPrefixedWriter) Write(data []byte) error {
	// Create length prefix
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	// Write length + data
	if _, err := w.conn.Write(lengthBuf); err != nil {
		return err
	}

	if _, err := w.conn.Write(data); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// Response Types (for convenience)
// ============================================================================

// CommandResponse represents a response from the server
type CommandResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse creates a successful response
func SuccessResponse(result interface{}) string {
	resp := CommandResponse{
		Success: true,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

// ErrorResponse creates an error response
func ErrorResponse(err string) string {
	resp := CommandResponse{
		Success: false,
		Error:   err,
	}
	data, _ := json.Marshal(resp)
	return string(data)
}
