# Claude Development Notes for go-textcleaner

## Socket Interface Testing & Session Persistence

### Overview
The project includes a Unix domain socket server interface for the headless TextCleaner core, allowing external clients to connect and send commands. The socket server can run in two modes:
1. **Headless server mode**: Socket server only, no GUI (persistent background service)
2. **GUI client mode**: GUI that connects to an existing socket server and loads its session state

### Session Persistence Feature
You can now create and maintain persistent sessions:
- Start a headless socket server that maintains state
- Connect multiple GUI clients to it
- Each GUI client automatically loads the current session state from the socket server
- Changes made in one client are visible to others in real-time

### Architecture
- **Socket Server**: `textcleaner_socket.go` - Implements the socket listener and message protocol
- **Protocol**: Length-prefixed JSON (4-byte big-endian length + JSON payload)
- **Multi-client support**: Multiple clients (GUIs, test clients, scripts) can connect simultaneously to the same socket server
- **Thread-safe core**: `TextCleanerCore` uses `sync.RWMutex` for thread-safe concurrent access from socket clients
- **Main Integration**: `main.go` with `--socket` flag for GUI clients, `--headless` for server-only mode
- **GUI Callbacks**: Socket commands trigger registered GUI refresh callbacks via `glib.IdleAdd()` for real-time synchronization across multiple connected GUIs

### Command-Line Usage

```bash
./go-textcleaner [options]

Options:
  -socket string
        Listen on Unix socket at this path for session persistence (e.g., /tmp/textcleaner.sock)
  -headless
        Run in headless mode (server only, no GUI). Requires --socket flag.
  -log-json
        Log raw JSON commands in headless mode
  -log-commands
        Log formatted commands in headless mode (with truncated arguments and responses)

Examples:
  ./go-textcleaner                                      # Start GUI only
  ./go-textcleaner --headless --socket /tmp/text.sock  # Start headless server
  ./go-textcleaner --socket /tmp/text.sock             # Connect GUI to running server
  ./go-textcleaner --headless --socket /tmp/text.sock --log-commands  # Headless with formatted logging
  ./go-textcleaner --headless --socket /tmp/text.sock --log-json      # Headless with JSON logging
  ./go-textcleaner --headless --socket /tmp/text.sock --log-json --log-commands  # Both logging modes
```

### Running Tests

#### 1. Build the project
```bash
go build -o go-textcleaner
```

#### 2. Run unit tests
```bash
# Run all tests
go test -v -timeout 15s

# Run only socket tests
go test -v -timeout 15s -run Socket

# Run specific test
go test -v -run TestSocketServerStart
```

### Session Persistence Workflow

This is the recommended approach for long-running sessions with persistent state.

#### Terminal 1: Start headless socket server
```bash
# Start the server with no GUI (persistent background service)
go run . --headless --socket /tmp/textcleaner.sock
# Or with built binary:
./go-textcleaner --headless --socket /tmp/textcleaner.sock
# Output: TextCleaner headless server listening on /tmp/textcleaner.sock
#         Press Ctrl+C to stop
```

The server now:
- Maintains the pipeline and data persistently
- Accepts multiple concurrent socket connections from clients
- Gracefully shuts down on Ctrl+C

#### Terminal 2: Start GUI client (loads existing session)
```bash
# Connect GUI to the running server, loads its state automatically
go run . --socket /tmp/textcleaner.sock
```

On startup:
- Connects to the socket server
- Loads the current pipeline
- Loads the current input text
- Restores the selected node
- Displays: "Connected to socket server at /tmp/textcleaner.sock, loading session..."
- Then starts the GUI with the loaded state

#### Terminal 3: Start another GUI (optional)
```bash
# Connect another GUI to the same server (loads the same session)
./go-textcleaner --socket /tmp/textcleaner.sock
```

This starts a second GUI connected to the same server. Both GUIs:
- Load the same pipeline and input text
- See changes in real-time when either GUI modifies data
- Are fully synchronized

#### Terminal 4: Connect test client
```bash
# Connect test client to the same server
go run test_socket_client.go --socket /tmp/textcleaner.sock
```

The test client:
- Connects to the shared socket server
- Can execute commands that modify the pipeline
- Changes are visible in all connected GUIs (Terminals 2 and 3) in real-time

#### Real-time Synchronization
- **GUI changes**: Instantly available to socket clients via commands
- **Socket changes**: Instantly visible in all connected GUIs via refresh callbacks
- **Multiple GUIs**: You can start multiple GUIs at any time - they all load the current state
- **Persistent state**: Data persists in the server even after all clients disconnect
- **Concurrent clients**: All client types (GUIs, test clients, custom scripts) can connect simultaneously

### Command Logging in Headless Mode

The headless server supports two types of command logging to help you monitor and debug socket interactions:

#### JSON Logging (`--log-json`)
Logs the raw JSON commands exactly as received from clients:

```bash
./go-textcleaner --headless --socket /tmp/text.sock --log-json
```

Example output:
```
TextCleaner headless server listening on /tmp/textcleaner.sock
JSON command logging: enabled
Press Ctrl+C to stop
[JSON] {"action":"create_node","params":{"type":"operation","name":"Uppercase","operation":"Uppercase","arg1":"","arg2":"","condition":""}}
[JSON] {"action":"set_input_text","params":{"text":"hello world"}}
[JSON] {"action":"get_output_text","params":{}}
```

#### Formatted Command Logging (`--log-commands`)
Logs human-readable formatted commands with truncated arguments and responses:

```bash
./go-textcleaner --headless --socket /tmp/text.sock --log-commands
```

Example output:
```
TextCleaner headless server listening on /tmp/textcleaner.sock
Formatted command logging: enabled
Press Ctrl+C to stop
[CMD] create_node(operation: Uppercase) => OK (node_id=node_0)
[CMD] set_input_text(hello world) => OK
[CMD] get_output_text() => OK (output=HELLO WORLD)
[CMD] export_pipeline() => OK (pipeline: 245 bytes)
```

**Truncation behavior:**
- Node IDs: truncated to 20 characters
- Node names: truncated to 30 characters
- Text arguments: truncated to 50 characters
- Error messages: truncated to 100 characters
- Long responses show byte count instead of full content (e.g., pipelines)

#### Using Both Logging Modes
You can enable both logging modes simultaneously for maximum visibility:

```bash
./go-textcleaner --headless --socket /tmp/text.sock --log-json --log-commands
```

Example output:
```
TextCleaner headless server listening on /tmp/textcleaner.sock
JSON command logging: enabled
Formatted command logging: enabled
Press Ctrl+C to stop
[JSON] {"action":"create_node","params":{"type":"operation","name":"Uppercase","operation":"Uppercase","arg1":"","arg2":"","condition":""}}
[CMD] create_node(operation: Uppercase) => OK (node_id=node_0)
```

**Use cases:**
- `--log-json`: Debugging protocol issues, command replay, or integration testing
- `--log-commands`: Monitoring server activity, user behavior analysis, or audit trails
- Both flags: Comprehensive logging for troubleshooting complex issues

### Running in Different Modes

#### GUI Only (No Socket)
```bash
go run .
# or
./go-textcleaner
```
Starts the GUI with a fresh pipeline. No socket server, changes not persistent across sessions.

#### Headless Server Only
```bash
go run . --headless --socket /tmp/textcleaner.sock
# or
./go-textcleaner --headless --socket /tmp/textcleaner.sock
```
Starts a persistent socket server with no GUI. Perfect for long-running background services.

#### GUI Client (Connects to Server)
```bash
go run . --socket /tmp/textcleaner.sock
# or
./go-textcleaner --socket /tmp/textcleaner.sock
```
Connects GUI to a running socket server and loads its session state. Multiple GUIs can connect to the same server. **Requires the server to be running first.**

#### Available Commands in Test Client

**Session State Commands (useful for session persistence):**

**1. Get input text:**
```json
{"action":"get_input_text","params":{}}
```
Returns the current input text being processed.

**2. Get output text:**
```json
{"action":"get_output_text","params":{}}
```
Returns the result of processing the input through the pipeline.

**3. Get selected node ID:**
```json
{"action":"get_selected_node_id","params":{}}
```
Returns the ID of the currently selected node (or empty string if none selected).

**Pipeline Management Commands:**

**4. List all root nodes:**
```json
{"action":"list_nodes","params":{}}
```

**5. Create a node:**
```json
{"action":"create_node","params":{"type":"operation","name":"Uppercase","operation":"Uppercase","arg1":"","arg2":"","condition":""}}
```

**6. Update a node:**
```json
{"action":"update_node","params":{"node_id":"node_0","type":"operation","name":"Uppercase","operation":"Uppercase","arg1":"","arg2":"","condition":""}}
```

**7. Delete a node:**
```json
{"action":"delete_node","params":{"node_id":"node_0"}}
```

**8. Add child node:**
```json
{"action":"add_child_node","params":{"parent_id":"node_0","type":"operation","name":"Lowercase","operation":"Lowercase","arg1":"","arg2":"","condition":""}}
```

**9. Select a node:**
```json
{"action":"select_node","params":{"node_id":"node_0"}}
```

**10. Get a specific node:**
```json
{"action":"get_node","params":{"node_id":"node_0"}}
```

**Text Processing Commands:**

**11. Set input text:**
```json
{"action":"set_input_text","params":{"text":"hello world"}}
```

**12. Get pipeline structure:**
```json
{"action":"get_pipeline","params":{}}
```

**Import/Export Commands:**

**13. Export pipeline:**
```json
{"action":"export_pipeline","params":{}}
```

**14. Import pipeline:**
```json
{"action":"import_pipeline","params":{"json":"[...]"}}
```

**Type `help` in the test client** to see all available commands with examples.

### Testing Example Workflow - Session Persistence

**Terminal 1: Start persistent headless server**
```bash
go run . --headless --socket /tmp/textcleaner.sock
# Output: TextCleaner headless server listening on /tmp/textcleaner.sock
```

**Terminal 2: Connect GUI (loads session)**
```bash
go run . --socket /tmp/textcleaner.sock
# Output: Connected to socket server at /tmp/textcleaner.sock, loading session...
#         Session loaded successfully
#         GUI opens with empty pipeline
```

**Terminal 3: Send commands via test client**
```bash
go run test_socket_client.go --socket /tmp/textcleaner.sock
# Interactive prompt appears
```

In test client prompt:
```
> {"action":"create_node","params":{"type":"operation","name":"Uppercase","operation":"Uppercase","arg1":"","arg2":"","condition":""}}
Response: { "success": true, "result": { "node_id": "node_0" } }

> {"action":"set_input_text","params":{"text":"hello world"}}
Response: { "success": true }

> {"action":"get_output_text","params":{}}
Response: { "success": true, "result": { "output": "HELLO WORLD" } }
```

**What happens:**
1. GUI in Terminal 2 sees all changes in real-time (pipeline tree updates, output text changes)
2. If you close the GUI and restart it (Terminal 2), it will reload the same session state
3. The server maintains state until you shut it down

### Protocol Details

**Message Format:**
```
[4 bytes: big-endian length] [JSON payload]
```

**Example (hex):**
```
00 00 00 30  (length = 48 bytes)
7b 22 61 63 74 69 6f 6e 22 3a ... (JSON data)
```

**Response Format:**
```json
{
  "success": true,
  "result": { ... },
  "error": ""  // Only present if success is false
}
```

### Key Implementation Files

**Core modifications:**
- `textcleaner_core.go` - Added `sync.RWMutex` for thread-safe concurrent access
  - All public methods now use proper locking
  - Write operations use `Lock()`, read-only operations use `RLock()`

**Socket implementation:**
- `textcleaner_socket.go` - Complete socket server and client implementation
  - **SocketServer**: Listens for connections, executes commands, triggers callbacks
    - `UpdateCallback` type for GUI refresh notifications
    - `SetUpdateCallback()` method to register GUI refresh handler
    - Calls callback after each command execution
  - **SocketClient**: Allows GUI to connect and query socket server
    - `NewSocketClient()` - Connect to running server
    - `Execute()` - Send command and get response
    - `Close()` - Disconnect gracefully
- `textcleaner_socket_test.go` - 7 test functions (all passing)
- `test_socket_client.go` - Interactive test client for manual testing

**Command extensions:**
- `textcleaner_commands.go` - New commands for session persistence
  - `get_input_text` - Returns current input text
  - `get_selected_node_id` - Returns currently selected node ID
  - Both commands are used by GUI on startup to load session state

**GUI integration:**
- `main.go` - Supports connection to existing socket servers
  - `loadStateFromSocket()` - GUI connects to server and loads:
    - Current pipeline via export_pipeline
    - Current input text via get_input_text
    - Current selected node via get_selected_node_id
  - `refreshUIFromCore()` method updates all UI elements when socket commands modify core
  - Uses `glib.IdleAdd()` to queue GUI updates from socket thread to main GTK thread
  - GUI does not start its own socket server—it only connects to existing ones
  - Clear error message if server is not running: guides user to start headless server first

### Thread Safety & Concurrency

Multiple clients (GUIs, test clients, etc.) connecting to the same socket server are handled safely:

**Locking Strategy:**
- `TextCleanerCore` has `sync.RWMutex` protecting all state
- All public methods acquire locks before accessing/modifying fields
- Write operations (Create, Update, Delete, SetInputText, Import) use `Lock()`
- Read operations (Get, Export) use `RLock()` to allow concurrent reads

**Goroutine Model:**
- **Main thread**: GTK event loop (handles user interactions and UI updates)
- **Socket thread**: Background goroutine accepting socket connections
- **Socket handler**: Separate goroutine for each connected client (limited to one)

**Cross-thread Communication:**
- Socket handler executes commands on the shared core (protected by mutex)
- After command execution, callback is invoked via `glib.IdleAdd()`
- GUI refresh happens on the main GTK thread (thread-safe)
- No direct GUI modifications from socket threads

**Data Safety:**
- Core state is always protected by mutex - no dirty reads
- GUI reflects accurate core state after each update
- Socket commands see consistent core state

### Cleanup

The socket file (`/tmp/textcleaner.sock`) is automatically removed:
- On server shutdown
- On server startup (removes stale sockets)
- On graceful shutdown (SIGINT/SIGTERM)

Manual cleanup if needed:
```bash
rm /tmp/textcleaner.sock
```

### Error Handling & Troubleshooting

**Issue: --headless used without --socket**
```
Error: --headless requires --socket to specify socket path
```
**Solution:** Use both flags together: `./go-textcleaner --headless --socket /tmp/textcleaner.sock`

**Issue: GUI fails with "no socket server running"**
```
Error: Failed to load session from socket: no socket server running at /tmp/textcleaner.sock
start one first with: ./go-textcleaner --headless --socket /tmp/textcleaner.sock
```
**Solution:** Start the headless server first in a separate terminal:
```bash
# Terminal 1: Start the server
./go-textcleaner --headless --socket /tmp/textcleaner.sock

# Terminal 2: Start GUI(s) after server is running
./go-textcleaner --socket /tmp/textcleaner.sock
```

**Issue: Socket file already exists from previous crash**
```bash
# Socket file is automatically cleaned up, but if needed:
rm /tmp/textcleaner.sock
```

**Multiple GUIs can connect to the same server**
You can start multiple GUIs connecting to the same socket server. They will all load the same session state and see changes in real-time:
```bash
# Terminal 1: Start headless server
./go-textcleaner --headless --socket /tmp/textcleaner.sock

# Terminal 2: Start first GUI
./go-textcleaner --socket /tmp/textcleaner.sock

# Terminal 3: Start second GUI (or more)
./go-textcleaner --socket /tmp/textcleaner.sock

# Terminal 4: Send commands via test client
go run test_socket_client.go --socket /tmp/textcleaner.sock
```
Both GUIs in Terminals 2 and 3 will see the changes in real-time.

**Issue: GUI doesn't see changes from socket commands**
- Ensure socket callback is registered (should happen automatically)
- Check console for "Error reading from client" messages
- Restart both server and GUI

### Testing the Session Persistence

**Recommended workflow for testing:**

1. **Build**
   ```bash
   go build -o go-textcleaner
   ```

2. **Terminal 1 - Start persistent headless server**
   ```bash
   ./go-textcleaner --headless --socket /tmp/textcleaner.sock
   # Output: TextCleaner headless server listening on /tmp/textcleaner.sock
   #         Press Ctrl+C to stop
   ```

3. **Terminal 2 - Connect GUI (loads session)**
   ```bash
   ./go-textcleaner --socket /tmp/textcleaner.sock
   # Output: Connected to socket server at /tmp/textcleaner.sock, loading session...
   #         Session loaded successfully
   # GUI window opens
   ```

4. **Terminal 3 - Modify via test client**
   ```bash
   go run test_socket_client.go --socket /tmp/textcleaner.sock

   # In client prompt:
   > {"action":"create_node","params":{"type":"operation","name":"Uppercase","operation":"Uppercase","arg1":"","arg2":"","condition":""}}
   Response: { "success": true, "result": { "node_id": "node_0" } }

   > {"action":"set_input_text","params":{"text":"test"}}
   Response: { "success": true }
   ```

5. **Observe in Terminal 2 GUI**
   - Pipeline tree updates instantly
   - Output text updates instantly
   - All changes from test client reflected in real-time

6. **Close GUI and restart it (Terminal 2)**
   ```bash
   # Close the GUI window
   # Then restart:
   ./go-textcleaner --socket /tmp/textcleaner.sock
   # Output: Connected to socket server... Session loaded successfully
   # The same pipeline and input/output are restored!
   ```

### Future Enhancements

- [ ] Add TCP socket support (for remote connections)
- [ ] Add authentication/authorization
- [ ] Add rate limiting
- [ ] Add session management
- [ ] Add HTTP/REST API wrapper around socket
- [ ] Add WebSocket support for browser clients
- [x] Add command logging/auditing for socket operations (implemented with `--log-json` and `--log-commands` flags)
- [ ] Add undo/redo across both GUI and socket interfaces
