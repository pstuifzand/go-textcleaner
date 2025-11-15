package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	appTitle  = "TextCleaner"
	appWidth  = 1200
	appHeight = 700
)

type TextCleaner struct {
	commands         TextCleanerCommands // Interface for all operations (socket wrapper)
	headlessProc     *os.Process         // Child headless process (if started by this GUI)
	window           *gtk.Window
	inputView        *gtk.TextView
	outputView       *gtk.TextView
	inputBuffer      *gtk.TextBuffer
	outputBuffer     *gtk.TextBuffer
	copyButton       *gtk.Button
	pipelineTree     *gtk.TreeView
	treeStore        *gtk.TreeStore
	paletteTree      *gtk.TreeView // Operations palette tree
	selectedNode     *gtk.TreePath
	nodeTypeCombo    *gtk.ComboBoxText
	operationCombo   *gtk.ComboBoxText
	argument1        *gtk.Entry
	argument2        *gtk.Entry
	conditionEntry   *gtk.Entry
	nodeNameEntry    *gtk.Entry
	createNodeButton *gtk.Button
	editNodeButton   *gtk.Button
	deleteNodeButton *gtk.Button
	indentButton     *gtk.Button
	unindentButton   *gtk.Button
	moveUpButton     *gtk.Button
	moveDownButton   *gtk.Button
	addChildButton   *gtk.Button
	editingMode      bool // True when actively editing a node (after double-click)
}

func main() {
	// Parse command-line flags
	socketPath := flag.String("socket", "", "Listen on Unix socket at this path (e.g., /tmp/textcleaner.sock)")
	headless := flag.Bool("headless", false, "Run in headless mode (server only, no GUI)")
	repl := flag.Bool("repl", false, "Run REPL mode (requires --socket)")
	logJSON := flag.Bool("log-json", false, "Log raw JSON commands in headless mode")
	logCommands := flag.Bool("log-commands", false, "Log formatted commands in headless mode")
	flag.Parse()

	// Create the headless core
	core := NewTextCleanerCore()

	// If headless mode with socket, start server and exit
	if *headless {
		if *socketPath == "" {
			log.Fatalf("Error: --headless requires --socket to specify socket path\n")
		}
		runHeadlessServer(*socketPath, core, *logJSON, *logCommands)
		return
	}

	// If REPL mode, start REPL and exit
	if *repl {
		if *socketPath == "" {
			log.Fatalf("Error: --repl requires --socket to specify socket path\n")
		}
		runREPLMode(*socketPath)
		return
	}

	// Otherwise, run GUI mode
	// Use default socket path if not specified
	if *socketPath == "" {
		*socketPath = generateRandomSocketPath()
	}

	// Initialize GTK
	gtk.Init(nil)

	// Try to connect to existing socket server
	socketClient, err := NewSocketClient(*socketPath)
	var headlessProc *os.Process

	if err != nil {
		// No existing server, start headless server as child process
		fmt.Printf("Starting headless socket server at %s...\n", *socketPath)
		headlessProc, err = startHeadlessChildProcess(*socketPath)
		if err != nil {
			log.Fatalf("Error: Failed to start headless socket server: %v\n", err)
		}

		// Wait for server to start and listen for connections
		socketClient, err = waitForSocketServer(*socketPath, 5*time.Second)
		if err != nil {
			headlessProc.Kill()
			log.Fatalf("Error: Failed to connect to socket server: %v\n", err)
		}
	}

	// Successfully connected to socket server
	commands := NewSocketClientCommands(socketClient)
	if err := loadStateFromSocket(core, socketClient); err != nil {
		if headlessProc != nil {
			headlessProc.Kill()
		}
		log.Fatalf("Failed to load session from socket: %v\n", err)
	}
	fmt.Printf("Connected to socket server at %s, loading session...\n", *socketPath)

	// Create the GUI application
	app := &TextCleaner{
		commands:     commands,
		headlessProc: headlessProc,
	}
	app.BuildUI()

	// Populate the pipeline tree from the loaded session
	app.refreshPipelineTree()

	// Populate the input buffer with the loaded input text
	inputText := commands.GetInputText()
	app.inputBuffer.SetText(inputText)

	// Update output display with processed text
	app.updateTextDisplay()

	// Restore the selected node from the session (if any)
	selectedNodeID := commands.GetSelectedNodeID()
	if selectedNodeID != "" {
		commands.SelectNode(selectedNodeID)
		node := commands.GetNode(selectedNodeID)
		if node != nil {
			app.loadNodeToUI(node)
		}
	}

	fmt.Println("Session loaded successfully")

	// Run the GUI (blocks until window is closed)
	gtk.Main()

	// Clean up child process if we started it
	if headlessProc != nil {
		headlessProc.Kill()
	}
}

// runHeadlessServer starts a socket server without GUI
func runHeadlessServer(socketPath string, core *TextCleanerCore, logJSON bool, logCommands bool) {
	server := NewSocketServer(socketPath, core)

	// Enable logging if requested
	server.SetLogJSON(logJSON)
	server.SetLogCommands(logCommands)

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start socket server: %v\n", err)
	}

	fmt.Printf("TextCleaner headless server listening on %s\n", socketPath)
	if logJSON {
		fmt.Println("JSON command logging: enabled")
	}
	if logCommands {
		fmt.Println("Formatted command logging: enabled")
	}
	fmt.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal (handled by the server itself)
	server.Wait()
	fmt.Println("Server stopped")
}

// runREPLMode starts a REPL session connected to a socket server
func runREPLMode(socketPath string) {
	session, err := NewREPLSession(socketPath)
	if err != nil {
		log.Fatalf("Error: Failed to connect to socket server: %v\n", err)
	}

	if err := session.Run(); err != nil {
		log.Fatalf("Error: REPL error: %v\n", err)
	}
}

// loadStateFromSocket loads the current state from a socket server via an existing client
func loadStateFromSocket(core *TextCleanerCore, client *SocketClient) error {

	// Load the current pipeline
	pipelineResp, err := client.Execute(`{"action":"export_pipeline","params":{}}`)
	if err != nil {
		return fmt.Errorf("failed to get pipeline: %w", err)
	}

	if success, ok := pipelineResp["success"].(bool); ok && success {
		if result, ok := pipelineResp["result"].(map[string]interface{}); ok {
			if pipeline, ok := result["pipeline"]; ok {
				// Convert the pipeline object back to JSON string for import
				pipelineJSON, err := json.Marshal(pipeline)
				if err != nil {
					return fmt.Errorf("failed to marshal pipeline: %w", err)
				}
				// Import the pipeline into our core
				if err := core.ImportPipeline(string(pipelineJSON)); err != nil {
					return fmt.Errorf("failed to import pipeline: %w", err)
				}
			}
		}
	}

	// Load the current input text
	inputResp, err := client.Execute(`{"action":"get_input_text","params":{}}`)
	if err != nil {
		return fmt.Errorf("failed to get input text: %w", err)
	}

	if success, ok := inputResp["success"].(bool); ok && success {
		if result, ok := inputResp["result"].(map[string]interface{}); ok {
			if text, ok := result["text"].(string); ok {
				core.SetInputText(text)
			}
		}
	}

	// Load the current selected node
	selectedResp, err := client.Execute(`{"action":"get_selected_node_id","params":{}}`)
	if err == nil {
		if success, ok := selectedResp["success"].(bool); ok && success {
			if result, ok := selectedResp["result"].(map[string]interface{}); ok {
				if nodeID, ok := result["node_id"].(string); ok && nodeID != "" {
					core.SelectNode(nodeID)
				}
			}
		}
	}

	fmt.Println("Session loaded successfully")
	return nil
}

func (tc *TextCleaner) BuildUI() {
	// Create main window
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	tc.window = win
	tc.window.SetTitle(appTitle)
	tc.window.SetDefaultSize(appWidth, appHeight)
	tc.window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Create main vertical box
	mainBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	mainBox.SetMarginTop(5)
	mainBox.SetMarginBottom(5)
	mainBox.SetMarginStart(5)
	mainBox.SetMarginEnd(5)

	// Create toolbar
	toolbar, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)

	// Spacer
	spacer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	toolbar.PackStart(spacer, true, true, 0)

	// Copy button
	copyButton, _ := gtk.ButtonNewWithLabel("Copy to Clipboard")
	tc.copyButton = copyButton
	toolbar.PackStart(copyButton, false, false, 0)

	mainBox.PackStart(toolbar, false, false, 0)

	// Create main horizontal paned (pipeline panel | text panes)
	mainPaned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	mainPaned.SetPosition(450) // Pipeline panel width

	// Create pipeline panel (left side)
	pipelinePanel := tc.createPipelinePanel()
	mainPaned.Add1(pipelinePanel)

	// Create horizontal paned for input/output (right side)
	textPaned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	textPaned.SetPosition(375) // Input pane width (half of remaining space after 450px pipeline)

	// Create input pane
	inputFrame := tc.createTextPane("Input", true)
	textPaned.Add1(inputFrame)

	// Create output pane
	outputFrame := tc.createTextPane("Output", false)
	textPaned.Add2(outputFrame)

	mainPaned.Add2(textPaned)

	mainBox.PackStart(mainPaned, true, true, 0)

	tc.window.Add(mainBox)
	tc.window.ShowAll()

	// Wire up event handlers
	tc.setupEventHandlers()
}

func (tc *TextCleaner) createNodeControls() *gtk.Box {
	controlsBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	controlsBox.SetMarginTop(10)
	controlsBox.SetMarginBottom(10)
	controlsBox.SetMarginStart(10)
	controlsBox.SetMarginEnd(10)

	// Node Type selector
	typeRow, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	typeLabel, _ := gtk.LabelNew("Node Type:")
	typeLabel.SetXAlign(0)
	typeLabel.SetWidthChars(12)
	typeRow.PackStart(typeLabel, false, false, 0)

	nodeTypeCombo, _ := gtk.ComboBoxTextNew()
	tc.nodeTypeCombo = nodeTypeCombo
	nodeTypeCombo.AppendText("Operation")
	nodeTypeCombo.AppendText("If (Conditional)")
	nodeTypeCombo.AppendText("ForEachLine")
	nodeTypeCombo.AppendText("Group")
	nodeTypeCombo.SetActive(0)
	typeRow.PackStart(nodeTypeCombo, true, true, 0)
	controlsBox.PackStart(typeRow, false, false, 0)

	// Operation selector (hidden for non-operation nodes)
	opRow, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	opLabel, _ := gtk.LabelNew("Operation:")
	opLabel.SetXAlign(0)
	opLabel.SetWidthChars(12)
	opRow.PackStart(opLabel, false, false, 0)

	operationCombo, _ := gtk.ComboBoxTextNew()
	tc.operationCombo = operationCombo
	operations := GetOperations()
	for _, op := range operations {
		operationCombo.AppendText(op.Name)
	}
	operationCombo.SetActive(0)
	opRow.PackStart(operationCombo, true, true, 0)
	controlsBox.PackStart(opRow, false, false, 0)

	// Node Name
	nameRow, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	nameLabel, _ := gtk.LabelNew("Name:")
	nameLabel.SetXAlign(0)
	nameLabel.SetWidthChars(12)
	nameRow.PackStart(nameLabel, false, false, 0)

	nodeNameEntry, _ := gtk.EntryNew()
	tc.nodeNameEntry = nodeNameEntry
	nodeNameEntry.SetPlaceholderText("Optional display name")
	nameRow.PackStart(nodeNameEntry, true, true, 0)
	controlsBox.PackStart(nameRow, false, false, 0)

	// Condition (for If nodes)
	condRow, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	condLabel, _ := gtk.LabelNew("Condition:")
	condLabel.SetXAlign(0)
	condLabel.SetWidthChars(12)
	condRow.PackStart(condLabel, false, false, 0)

	conditionEntry, _ := gtk.EntryNew()
	tc.conditionEntry = conditionEntry
	conditionEntry.SetPlaceholderText("Regex pattern")
	condRow.PackStart(conditionEntry, true, true, 0)
	controlsBox.PackStart(condRow, false, false, 0)

	// Argument 1
	arg1Row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	arg1Label, _ := gtk.LabelNew("Arg1:")
	arg1Label.SetXAlign(0)
	arg1Label.SetWidthChars(12)
	arg1Row.PackStart(arg1Label, false, false, 0)

	arg1Entry, _ := gtk.EntryNew()
	tc.argument1 = arg1Entry
	arg1Row.PackStart(arg1Entry, true, true, 0)
	controlsBox.PackStart(arg1Row, false, false, 0)

	// Argument 2
	arg2Row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	arg2Label, _ := gtk.LabelNew("Arg2:")
	arg2Label.SetXAlign(0)
	arg2Label.SetWidthChars(12)
	arg2Row.PackStart(arg2Label, false, false, 0)

	arg2Entry, _ := gtk.EntryNew()
	tc.argument2 = arg2Entry
	arg2Row.PackStart(arg2Entry, true, true, 0)
	controlsBox.PackStart(arg2Row, false, false, 0)

	// Buttons row
	buttonRow, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)

	createNodeButton, _ := gtk.ButtonNewWithLabel("Create Node")
	tc.createNodeButton = createNodeButton
	buttonRow.PackStart(createNodeButton, true, true, 0)

	editNodeButton, _ := gtk.ButtonNewWithLabel("Update Node")
	tc.editNodeButton = editNodeButton
	editNodeButton.SetSensitive(false)
	buttonRow.PackStart(editNodeButton, true, true, 0)

	controlsBox.PackStart(buttonRow, false, false, 0)

	// Tree operations row 1 (Add Child, Indent, Unindent, Delete)
	treeOpsRow1, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)

	addChildButton, _ := gtk.ButtonNewWithLabel("Add Child")
	tc.addChildButton = addChildButton
	addChildButton.SetSensitive(false)
	treeOpsRow1.PackStart(addChildButton, true, true, 0)

	indentButton, _ := gtk.ButtonNewWithLabel("Indent")
	tc.indentButton = indentButton
	indentButton.SetSensitive(false)
	treeOpsRow1.PackStart(indentButton, true, true, 0)

	unindentButton, _ := gtk.ButtonNewWithLabel("Unindent")
	tc.unindentButton = unindentButton
	unindentButton.SetSensitive(false)
	treeOpsRow1.PackStart(unindentButton, true, true, 0)

	deleteNodeButton, _ := gtk.ButtonNewWithLabel("Delete")
	tc.deleteNodeButton = deleteNodeButton
	deleteNodeButton.SetSensitive(false)
	treeOpsRow1.PackStart(deleteNodeButton, true, true, 0)

	controlsBox.PackStart(treeOpsRow1, false, false, 0)

	// Tree operations row 2 (Move Up, Move Down)
	treeOpsRow2, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)

	moveUpButton, _ := gtk.ButtonNewWithLabel("Move Up")
	tc.moveUpButton = moveUpButton
	moveUpButton.SetSensitive(false)
	treeOpsRow2.PackStart(moveUpButton, true, true, 0)

	moveDownButton, _ := gtk.ButtonNewWithLabel("Move Down")
	tc.moveDownButton = moveDownButton
	moveDownButton.SetSensitive(false)
	treeOpsRow2.PackStart(moveDownButton, true, true, 0)

	// Spacer
	spacer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	treeOpsRow2.PackStart(spacer, true, true, 0)

	controlsBox.PackStart(treeOpsRow2, false, false, 0)

	return controlsBox
}

func (tc *TextCleaner) createOperationsPalette() *gtk.Box {
	paletteBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)

	// Title label
	titleLabel, _ := gtk.LabelNew("Operations Palette")
	titleLabel.SetMarkup("<b>Operations</b>")
	paletteBox.PackStart(titleLabel, false, false, 0)

	// Scrolled window for the palette
	scrolledWindow, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolledWindow.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scrolledWindow.SetSizeRequest(200, -1)

	// Create list store with two columns: display name and operation name
	listStore, _ := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)

	// Populate with all operations
	operations := GetOperations()
	for _, op := range operations {
		iter := listStore.Append()
		listStore.SetValue(iter, 0, op.Name)
		listStore.SetValue(iter, 1, op.Name)
	}

	// Create tree view for the palette
	treeView, _ := gtk.TreeViewNew()
	tc.paletteTree = treeView
	treeView.SetModel(listStore)

	// Add column for operation name
	renderer, _ := gtk.CellRendererTextNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute("Operation", renderer, "text", 0)
	treeView.AppendColumn(column)

	// Set properties
	treeView.SetHeadersVisible(false)

	// Enable drag source
	targetEntry, _ := gtk.TargetEntryNew("text/plain", gtk.TARGET_SAME_APP, 0)
	targets := []gtk.TargetEntry{*targetEntry}
	treeView.DragSourceSet(gdk.BUTTON1_MASK, targets, gdk.ACTION_COPY)

	scrolledWindow.Add(treeView)
	paletteBox.PackStart(scrolledWindow, true, true, 0)

	return paletteBox
}

func (tc *TextCleaner) createPipelinePanel() *gtk.Box {
	panel, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	panel.SetMarginTop(5)
	panel.SetMarginBottom(5)
	panel.SetMarginStart(5)
	panel.SetMarginEnd(5)

	// Create horizontal paned for palette and pipeline sections
	mainPaned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	mainPaned.SetPosition(220) // Palette width

	// ===== LEFT SECTION: Operations Palette =====
	palettePanel := tc.createOperationsPalette()
	mainPaned.Add1(palettePanel)

	// ===== RIGHT SECTION: Controls and Tree =====
	rightPanel, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)

	// Create paned layout for controls and tree
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	paned.SetPosition(380) // Controls panel height

	// Top: Node Controls
	nodeControls := tc.createNodeControls()
	controlsFrame, _ := gtk.FrameNew("Node Controls")
	controlsFrame.Add(nodeControls)
	paned.Add1(controlsFrame)

	// Bottom: Pipeline Tree
	treePanel, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)

	// Title label
	titleLabel, _ := gtk.LabelNew("Pipeline Tree")
	titleLabel.SetMarkup("<b>Pipeline Tree</b>")
	treePanel.PackStart(titleLabel, false, false, 0)

	// Scrolled window for the tree
	scrolledWindow, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolledWindow.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scrolledWindow.SetSizeRequest(300, -1)

	// Create tree store with two columns: display text and node ID
	treeStore, _ := gtk.TreeStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)
	tc.treeStore = treeStore

	// Create tree view
	treeView, _ := gtk.TreeViewNew()
	tc.pipelineTree = treeView
	treeView.SetModel(treeStore)

	// Add column for display text (column 0)
	renderer, _ := gtk.CellRendererTextNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute("Node", renderer, "text", 0)
	treeView.AppendColumn(column)

	// Set properties
	treeView.SetHeadersVisible(false)

	// Enable drag destination for the pipeline tree
	targetEntry, _ := gtk.TargetEntryNew("text/plain", gtk.TARGET_SAME_APP, 0)
	targets := []gtk.TargetEntry{*targetEntry}
	treeView.DragDestSet(gtk.DEST_DEFAULT_ALL, targets, gdk.ACTION_COPY)

	scrolledWindow.Add(treeView)
	treePanel.PackStart(scrolledWindow, true, true, 0)

	paned.Add2(treePanel)

	rightPanel.PackStart(paned, true, true, 0)
	mainPaned.Add2(rightPanel)

	panel.PackStart(mainPaned, true, true, 0)

	return panel
}

func (tc *TextCleaner) createTextPane(title string, isInput bool) *gtk.Frame {
	frame, _ := gtk.FrameNew(title)

	scrolledWindow, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolledWindow.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)

	textView, _ := gtk.TextViewNew()
	textView.SetWrapMode(gtk.WRAP_WORD)
	textView.SetMonospace(true)

	// Get the text buffer
	buffer, _ := textView.GetBuffer()

	if isInput {
		tc.inputView = textView
		tc.inputBuffer = buffer
		textView.SetEditable(true)
	} else {
		tc.outputView = textView
		tc.outputBuffer = buffer
		textView.SetEditable(false)
	}

	scrolledWindow.Add(textView)
	frame.Add(scrolledWindow)

	return frame
}

// setupEventHandlers wires up all event handlers
func (tc *TextCleaner) setupEventHandlers() {
	// Input buffer changed - process text in real-time
	tc.inputBuffer.Connect("changed", func() {
		tc.processText()
	})

	// Copy button
	tc.copyButton.Connect("clicked", func() {
		tc.copyToClipboard()
	})

	// Node type changed
	tc.nodeTypeCombo.Connect("changed", func() {
		tc.updateNodeTypeUI()
	})

	// Tree selection changed - update button states
	tc.pipelineTree.Connect("cursor-changed", func() {
		tc.updateTreeSelection()
	})

	// Tree row activated (double-click) - open node for editing
	tc.pipelineTree.Connect("row-activated", func() {
		tc.openNodeForEditing()
	})

	// Drag and drop handlers for operations palette
	tc.setupDragAndDrop()

	// Create Node button
	tc.createNodeButton.Connect("clicked", func() {
		tc.createNewNode()
	})

	// Edit Node button
	tc.editNodeButton.Connect("clicked", func() {
		tc.updateSelectedNode()
	})

	// Add Child button
	tc.addChildButton.Connect("clicked", func() {
		tc.addChildNode()
	})

	// Delete Node button
	tc.deleteNodeButton.Connect("clicked", func() {
		tc.deleteSelectedNode()
	})

	// Indent button
	tc.indentButton.Connect("clicked", func() {
		tc.indentSelectedNode()
	})

	// Unindent button
	tc.unindentButton.Connect("clicked", func() {
		tc.unindentSelectedNode()
	})

	// Move Up button
	tc.moveUpButton.Connect("clicked", func() {
		tc.moveSelectedNodeUp()
	})

	// Move Down button
	tc.moveDownButton.Connect("clicked", func() {
		tc.moveSelectedNodeDown()
	})

	// ===== REAL-TIME NODE EDITING =====
	// Wire up auto-update handlers for node property fields
	// Only update when in editing mode (after double-click)

	// Node name field - auto-update when edited (only in editing mode)
	tc.nodeNameEntry.Connect("changed", func() {
		if tc.editingMode {
			tc.updateNodeFromUIFields()
		}
	})

	// Operation combo - auto-update when changed (only in editing mode)
	tc.operationCombo.Connect("changed", func() {
		if tc.editingMode && tc.commands.GetSelectedNodeID() != "" {
			tc.updateNodeFromUIFields()
		}
	})

	// Argument 1 - auto-update when edited (only in editing mode)
	tc.argument1.Connect("changed", func() {
		if tc.editingMode {
			tc.updateNodeFromUIFields()
		}
	})

	// Argument 2 - auto-update when edited (only in editing mode)
	tc.argument2.Connect("changed", func() {
		if tc.editingMode {
			tc.updateNodeFromUIFields()
		}
	})

	// Condition field - auto-update when edited (only in editing mode)
	tc.conditionEntry.Connect("changed", func() {
		if tc.editingMode {
			tc.updateNodeFromUIFields()
		}
	})
}

func (tc *TextCleaner) setupDragAndDrop() {
	// Palette drag source: provide operation name when dragging
	tc.paletteTree.Connect("drag-data-get", func(widget *gtk.TreeView, context *gdk.DragContext, data *gtk.SelectionData, info uint, time uint) {
		// Get selected operation from palette
		selection, _ := widget.GetSelection()
		model, iter, ok := selection.GetSelected()
		if !ok {
			return
		}

		// Get operation name from column 1
		val, _ := model.(*gtk.TreeModel).GetValue(iter, 1)
		operationName, _ := val.GetString()

		// Set the selection data to the operation name
		data.SetText(operationName)
	})

	// Pipeline tree drag source: provide node ID when dragging
	targetEntry, _ := gtk.TargetEntryNew("text/plain", gtk.TARGET_SAME_APP, 0)
	targets := []gtk.TargetEntry{*targetEntry}
	tc.pipelineTree.DragSourceSet(gdk.BUTTON1_MASK, targets, gdk.ACTION_MOVE)

	tc.pipelineTree.Connect("drag-data-get", func(widget *gtk.TreeView, context *gdk.DragContext, data *gtk.SelectionData, info uint, time uint) {
		// Get selected node from pipeline tree
		selection, _ := widget.GetSelection()
		_, iter, ok := selection.GetSelected()
		if !ok {
			return
		}

		// Get node ID from column 1
		val, _ := tc.treeStore.GetValue(iter, 1)
		nodeID, _ := val.GetString()

		// Set the selection data to "NODE:nodeID" to distinguish from palette drags
		data.SetText("NODE:" + nodeID)
	})

	// Pipeline tree drag destination: create node from palette or move existing node
	tc.pipelineTree.Connect("drag-data-received", func(widget *gtk.TreeView, context *gdk.DragContext, x int, y int, data *gtk.SelectionData, info uint, time uint) {
		dragData := data.GetText()
		if dragData == "" {
			return
		}

		// Get the drop position
		path, pos, ok := widget.GetDestRowAtPos(x, y)
		_ = ok // We proceed whether or not we got a valid position

		// Determine if this is a palette drag (operation name) or tree drag (NODE:nodeID)
		isTreeDrag := len(dragData) > 5 && dragData[:5] == "NODE:"

		if isTreeDrag {
			// Handle tree node reordering
			nodeID := dragData[5:]

			// Prevent dragging a node to an invalid location
			if !ok || path == nil {
				// Dropping in empty space - add to root at the end
				err := tc.commands.MoveNodeToPosition(nodeID, "", -1)
				if err != nil {
					return
				}
			} else {
				// Get the target node
				iter, _ := tc.treeStore.GetIter(path)
				val, _ := tc.treeStore.GetValue(iter, 1)
				targetID, _ := val.GetString()

				// Prevent dragging into itself
				if nodeID == targetID {
					return
				}

				// Determine new parent and position based on drop location
				var newParentID string
				var newPosition int

				// Check drop position relative to target row
				switch pos {
				case gtk.TREE_VIEW_DROP_INTO_OR_BEFORE, gtk.TREE_VIEW_DROP_INTO_OR_AFTER:
					// Drop INTO the target node - make it a child
					newParentID = targetID
					newPosition = 0 // Add at beginning of children

				case gtk.TREE_VIEW_DROP_BEFORE:
					// Drop BEFORE the target - make it a sibling
					// Find parent and position of target
					parentNode := tc.findParentNode(targetID)
					if parentNode != nil {
						newParentID = parentNode.ID
					}
					// Find position of target in parent's children
					targetParent := tc.commands.GetNode(newParentID)
					if targetParent != nil || newParentID == "" {
						var childrenList []PipelineNode
						if newParentID == "" {
							childrenList = tc.commands.GetPipeline()
						} else {
							childrenList = targetParent.Children
						}
						for i, child := range childrenList {
							if child.ID == targetID {
								newPosition = i
								break
							}
						}
					}

				case gtk.TREE_VIEW_DROP_AFTER:
					// Drop AFTER the target - make it a sibling, positioned after
					// Find parent and position of target
					parentNode := tc.findParentNode(targetID)
					if parentNode != nil {
						newParentID = parentNode.ID
					}
					// Find position of target in parent's children
					targetParent := tc.commands.GetNode(newParentID)
					if targetParent != nil || newParentID == "" {
						var childrenList []PipelineNode
						if newParentID == "" {
							childrenList = tc.commands.GetPipeline()
						} else {
							childrenList = targetParent.Children
						}
						for i, child := range childrenList {
							if child.ID == targetID {
								newPosition = i + 1
								break
							}
						}
					}
				}

				// Move the node
				err := tc.commands.MoveNodeToPosition(nodeID, newParentID, newPosition)
				if err != nil {
					return
				}
			}

			// Refresh UI - keep the node selected
			tc.refreshPipelineTree()
			tc.updateTextDisplay()
			tc.commands.SelectNode(nodeID)
			tc.updateTreeSelection()

		} else {
			// Handle palette drag: create new node
			operationName := dragData

			var parentID string
			if path != nil {
				// Get the node ID at the drop position
				iter, _ := tc.treeStore.GetIter(path)
				val, _ := tc.treeStore.GetValue(iter, 1)
				nodeID, _ := val.GetString()

				// If dropping on a node, determine if it should be a child or sibling
				if pos == gtk.TREE_VIEW_DROP_INTO_OR_BEFORE || pos == gtk.TREE_VIEW_DROP_INTO_OR_AFTER {
					// Drop as child of the target node
					parentID = nodeID
				} else {
					// Drop as sibling - find parent of target node
					parentNode := tc.findParentNode(nodeID)
					if parentNode != nil {
						parentID = parentNode.ID
					}
				}
			}

			// Create the new node
			var newNodeID string
			if parentID != "" {
				// Add as child node
				newNodeID, _ = tc.commands.AddChildNode(
					parentID,
					"operation",
					operationName,
					operationName,
					"",
					"",
					"",
				)
			} else {
				// Add as root node
				newNodeID = tc.commands.CreateNode(
					"operation",
					operationName,
					operationName,
					"",
					"",
					"",
				)
			}

			// Refresh UI
			tc.refreshPipelineTree()
			tc.updateTextDisplay()

			// Select and enter editing mode for the new node
			tc.commands.SelectNode(newNodeID)
			node := tc.commands.GetNode(newNodeID)
			if node != nil {
				tc.loadNodeToUI(node)
				tc.editingMode = true
				tc.updateTreeEditingIndicators()
			}
			tc.updateButtonStates()
		}
	})

	// Pipeline tree drag motion: provide visual feedback during drag
	tc.pipelineTree.Connect("drag-motion", func(widget *gtk.TreeView, context *gdk.DragContext, x int, y int, time uint) bool {
		path, pos, ok := widget.GetDestRowAtPos(x, y)

		if ok && path != nil {
			// Highlight the drop target row
			widget.SetDragDestRow(path, pos)
		}

		return true
	})

	// Pipeline tree drag leave: clear visual feedback
	tc.pipelineTree.Connect("drag-leave", func(widget *gtk.TreeView, context *gdk.DragContext, time uint) {
		// Clear any drag highlighting
		widget.SetDragDestRow(nil, gtk.TREE_VIEW_DROP_BEFORE)
	})
}

func (tc *TextCleaner) findParentNode(nodeID string) *PipelineNode {
	pipeline := tc.commands.GetPipeline()
	return tc.findParentNodeRecursive(&pipeline, nodeID, nil)
}

func (tc *TextCleaner) findParentNodeRecursive(nodes *[]PipelineNode, targetID string, parent *PipelineNode) *PipelineNode {
	for i := range *nodes {
		node := &(*nodes)[i]

		// Check if this node is the target
		if node.ID == targetID {
			return parent
		}

		// Search in children
		if result := tc.findParentNodeRecursive(&node.Children, targetID, node); result != nil {
			return result
		}

		// Search in else children
		if result := tc.findParentNodeRecursive(&node.ElseChildren, targetID, node); result != nil {
			return result
		}
	}
	return nil
}

func (tc *TextCleaner) updateNodeTypeUI() {
	nodeType := tc.nodeTypeCombo.GetActiveText()

	// Show/hide fields based on node type
	switch nodeType {
	case "Operation":
		tc.operationCombo.ShowAll()
		tc.argument1.ShowAll()
		tc.argument2.ShowAll()
		tc.conditionEntry.Hide()
	case "If (Conditional)":
		tc.operationCombo.Hide()
		tc.argument1.Hide()
		tc.argument2.Hide()
		tc.conditionEntry.ShowAll()
	case "ForEachLine":
		tc.operationCombo.Hide()
		tc.argument1.Hide()
		tc.argument2.Hide()
		tc.conditionEntry.Hide()
	case "Group":
		tc.operationCombo.Hide()
		tc.argument1.Hide()
		tc.argument2.Hide()
		tc.conditionEntry.Hide()
	}
}

// openNodeForEditing opens the currently selected node for editing
func (tc *TextCleaner) openNodeForEditing() {
	selection, _ := tc.pipelineTree.GetSelection()
	_, iter, ok := selection.GetSelected()
	if !ok {
		return
	}

	// Get the node ID from column 1 of the tree
	val, _ := tc.treeStore.GetValue(iter, 1)
	nodeID, _ := val.GetString()

	// Find the node by ID in the pipeline
	foundNode := tc.commands.GetNode(nodeID)
	if foundNode != nil {
		tc.commands.SelectNode(nodeID)
		tc.loadNodeToUI(foundNode)
		tc.updateButtonStates()
		// Enter editing mode - real-time updates will now be active
		tc.editingMode = true
		// Update the tree to show editing indicator (✏️)
		tc.updateSingleNodeDisplay(nodeID)
	}
}

func (tc *TextCleaner) updateTreeSelection() {
	// Save the previously selected node ID before changing selection
	oldSelectedID := tc.commands.GetSelectedNodeID()

	// Single-click stops editing mode
	tc.editingMode = false
	tc.clearNodeInputs()

	selection, _ := tc.pipelineTree.GetSelection()
	_, iter, ok := selection.GetSelected()
	if !ok {
		tc.commands.SelectNode("")
		tc.updateButtonStates()
		// Remove editing indicator from previously selected node
		if oldSelectedID != "" {
			tc.updateSingleNodeDisplay(oldSelectedID)
		}
		// Show full pipeline output when no node is selected
		tc.updateTextDisplay()
		return
	}

	// Get the node ID from column 1 of the tree
	val, _ := tc.treeStore.GetValue(iter, 1)
	nodeID, _ := val.GetString()

	// Find the node by ID in the pipeline
	foundNode := tc.commands.GetNode(nodeID)
	if foundNode != nil {
		tc.commands.SelectNode(nodeID)
	} else {
		tc.commands.SelectNode("")
	}

	tc.updateButtonStates()

	// Remove the editing indicator from the previously selected node
	if oldSelectedID != "" && oldSelectedID != nodeID {
		tc.updateSingleNodeDisplay(oldSelectedID)
	}

	// Update the newly selected node (without editing indicator since editingMode is false)
	if nodeID != "" {
		tc.updateSingleNodeDisplay(nodeID)
	}

	// Show the text output up to and including the selected node
	if nodeID != "" {
		tc.updateTextDisplayAtNode(nodeID)
	}
}

func (tc *TextCleaner) loadNodeToUI(node *PipelineNode) {
	// Set node type
	for i := 0; i < 4; i++ {
		if tc.nodeTypeCombo.GetActiveText() == "" {
			break
		}
	}

	switch node.Type {
	case "operation":
		tc.nodeTypeCombo.SetActive(0)
		// Find and set operation
		operations := GetOperations()
		for i, op := range operations {
			if op.Name == node.Operation {
				tc.operationCombo.SetActive(i)
				break
			}
		}
	case "if":
		tc.nodeTypeCombo.SetActive(1)
		tc.conditionEntry.SetText(node.Condition)
	case "foreach":
		tc.nodeTypeCombo.SetActive(2)
	case "group":
		tc.nodeTypeCombo.SetActive(3)
	}

	tc.nodeNameEntry.SetText(node.Name)
	tc.argument1.SetText(node.Arg1)
	tc.argument2.SetText(node.Arg2)
	tc.updateNodeTypeUI()
}

func (tc *TextCleaner) createNewNode() {
	// For Operation nodes, default to "Untitled" name and "Identity" operation
	// For other types, use the default empty name
	nodeType := tc.nodeTypeCombo.GetActiveText()

	// Default to "Untitled" if no name provided
	nodeName := "Untitled"
	operation := ""
	arg1 := ""
	arg2 := ""
	condition := ""

	if nodeType == "Operation" {
		// For operation nodes, default to "Identity" (no-op)
		operation = "Identity"
	} else if nodeType == "If (Conditional)" {
		// Conditional nodes have no operation
		nodeName = "If"
	} else if nodeType == "ForEachLine" {
		nodeName = "ForEach"
	} else if nodeType == "Group" {
		nodeName = "Group"
	}

	// Convert UI node type to core node type
	coreNodeType := tc.getNodeTypeFromUI(nodeType)

	// Create node via commands interface (works with both local core and socket wrapper)
	nodeID := tc.commands.CreateNode(
		coreNodeType,
		nodeName,
		operation,
		arg1,
		arg2,
		condition,
	)

	// Refresh UI
	tc.refreshPipelineTree()
	tc.updateTextDisplay()

	// Select and enter editing mode for the newly created node
	tc.commands.SelectNode(nodeID)
	node := tc.commands.GetNode(nodeID)
	if node != nil {
		tc.loadNodeToUI(node)
		tc.editingMode = true
		tc.updateTreeEditingIndicators()
	}
	tc.updateButtonStates()

	// Clear the form inputs (they will be refilled from the node)
	// Don't clear here since we just loaded the node values above
}

func (tc *TextCleaner) updateSelectedNode() {
	if tc.commands.GetSelectedNodeID() == "" {
		return
	}

	nodeType := tc.nodeTypeCombo.GetActiveText()
	nodeName, _ := tc.nodeNameEntry.GetText()
	operation := ""
	arg1 := ""
	arg2 := ""
	condition := ""

	if nodeType == "Operation" {
		operation = tc.operationCombo.GetActiveText()
		arg1, _ = tc.argument1.GetText()
		arg2, _ = tc.argument2.GetText()
	} else if nodeType == "If (Conditional)" {
		condition, _ = tc.conditionEntry.GetText()
	}

	selectedID := tc.commands.GetSelectedNodeID()

	// Update node via commands interface (works with both local core and socket wrapper)
	err := tc.commands.UpdateNode(
		selectedID,
		nodeName,
		operation,
		arg1,
		arg2,
		condition,
	)

	if err != nil {
		return
	}

	// Refresh UI
	tc.refreshPipelineTree()
	tc.updateTextDisplay()

	// Reload the updated node into the UI so user can test changes
	updatedNode := tc.commands.GetNode(tc.commands.GetSelectedNodeID())
	if updatedNode != nil {
		tc.loadNodeToUI(updatedNode)
	}
}

// updateNodeFromUIFields reads current UI field values and updates the selected node in real-time
// This is called whenever a field is edited to provide immediate feedback
func (tc *TextCleaner) updateNodeFromUIFields() {
	selectedID := tc.commands.GetSelectedNodeID()
	if selectedID == "" {
		return
	}

	nodeType := tc.nodeTypeCombo.GetActiveText()
	nodeName, _ := tc.nodeNameEntry.GetText()
	operation := ""
	arg1 := ""
	arg2 := ""
	condition := ""

	if nodeType == "Operation" {
		operation = tc.operationCombo.GetActiveText()
		arg1, _ = tc.argument1.GetText()
		arg2, _ = tc.argument2.GetText()
	} else if nodeType == "If (Conditional)" {
		condition, _ = tc.conditionEntry.GetText()
	}

	// Update node via commands interface (works with both local core and socket wrapper)
	err := tc.commands.UpdateNode(
		selectedID,
		nodeName,
		operation,
		arg1,
		arg2,
		condition,
	)

	if err != nil {
		return
	}

	// Update UI to show changes in real-time
	// Only update the single node display and output, don't refresh entire tree
	// (to avoid segfault from modifying tree during signal handling)
	node := tc.commands.GetNode(selectedID)
	if node != nil {
		tc.updateSingleNodeDisplay(selectedID)
	}
	tc.updateTextDisplay()
}

func (tc *TextCleaner) deleteSelectedNode() {
	if tc.commands.GetSelectedNodeID() == "" {
		return
	}

	selectedID := tc.commands.GetSelectedNodeID()

	// Delete via commands interface (works with both local core and socket wrapper)
	err := tc.commands.DeleteNode(selectedID)
	if err != nil {
		return
	}

	// Refresh UI
	tc.refreshPipelineTree()
	tc.updateTextDisplay()
	tc.clearNodeInputs()
	tc.updateButtonStates()
}

func (tc *TextCleaner) addChildNode() {
	if tc.commands.GetSelectedNodeID() == "" {
		return
	}

	nodeType := tc.nodeTypeCombo.GetActiveText()
	nodeName, _ := tc.nodeNameEntry.GetText()
	operation := ""
	arg1 := ""
	arg2 := ""
	condition := ""

	if nodeType == "Operation" {
		operation = tc.operationCombo.GetActiveText()
		arg1, _ = tc.argument1.GetText()
		arg2, _ = tc.argument2.GetText()
	} else if nodeType == "If (Conditional)" {
		condition, _ = tc.conditionEntry.GetText()
	}

	coreNodeType := tc.getNodeTypeFromUI(nodeType)
	parentID := tc.commands.GetSelectedNodeID()

	// Add child node via commands interface (works with both local core and socket wrapper)
	_, err := tc.commands.AddChildNode(
		parentID,
		coreNodeType,
		nodeName,
		operation,
		arg1,
		arg2,
		condition,
	)

	if err != nil {
		return
	}

	// Refresh UI
	tc.refreshPipelineTree()
	tc.updateTextDisplay()

	// Clear inputs
	tc.clearNodeInputs()
}

func (tc *TextCleaner) indentSelectedNode() {
	selectedID := tc.commands.GetSelectedNodeID()
	if selectedID == "" {
		return
	}

	if err := tc.commands.IndentNode(selectedID); err != nil {
		return
	}

	// Refresh UI
	tc.refreshPipelineTree()
	tc.updateTextDisplay()
	tc.updateButtonStates()
}

func (tc *TextCleaner) unindentSelectedNode() {
	selectedID := tc.commands.GetSelectedNodeID()
	if selectedID == "" {
		return
	}

	if err := tc.commands.UnindentNode(selectedID); err != nil {
		return
	}

	// Refresh UI
	tc.refreshPipelineTree()
	tc.updateTextDisplay()
	tc.updateButtonStates()
}

func (tc *TextCleaner) moveSelectedNodeUp() {
	selectedID := tc.commands.GetSelectedNodeID()
	if selectedID == "" {
		return
	}

	if err := tc.commands.MoveNodeUp(selectedID); err != nil {
		return
	}

	// Refresh UI
	tc.refreshPipelineTree()
	tc.updateTextDisplay()
	tc.updateButtonStates()
}

func (tc *TextCleaner) moveSelectedNodeDown() {
	selectedID := tc.commands.GetSelectedNodeID()
	if selectedID == "" {
		return
	}

	if err := tc.commands.MoveNodeDown(selectedID); err != nil {
		return
	}

	// Refresh UI
	tc.refreshPipelineTree()
	tc.updateTextDisplay()
	tc.updateButtonStates()
}

func (tc *TextCleaner) updateButtonStates() {
	selectedID := tc.commands.GetSelectedNodeID()
	hasSelection := selectedID != ""

	tc.editNodeButton.SetSensitive(hasSelection)
	tc.deleteNodeButton.SetSensitive(hasSelection)
	tc.addChildButton.SetSensitive(hasSelection)
	tc.indentButton.SetSensitive(hasSelection && tc.commands.CanIndentNode(selectedID))
	tc.unindentButton.SetSensitive(hasSelection && tc.commands.CanUnindentNode(selectedID))
	tc.moveUpButton.SetSensitive(hasSelection && tc.commands.CanMoveNodeUp(selectedID))
	tc.moveDownButton.SetSensitive(hasSelection && tc.commands.CanMoveNodeDown(selectedID))
}

func (tc *TextCleaner) clearNodeInputs() {
	tc.nodeNameEntry.SetText("")
	tc.argument1.SetText("")
	tc.argument2.SetText("")
	tc.conditionEntry.SetText("")
	tc.operationCombo.SetActive(0)
	tc.nodeTypeCombo.SetActive(0)
}

func (tc *TextCleaner) refreshPipelineTree() {
	tc.treeStore.Clear()

	// Add all root-level nodes from core
	pipeline := tc.commands.GetPipeline()
	for i, node := range pipeline {
		tc.addNodeToTree(&node, nil, i)
	}

	tc.pipelineTree.ShowAll()

	// Expand all nodes
	tc.pipelineTree.ExpandAll()

	// Update visual editing indicators
	tc.updateTreeEditingIndicators()
}

// updateSingleNodeDisplay updates the display text of a single node in the tree
// without clearing the entire tree (safe to call from signal handlers)
func (tc *TextCleaner) updateSingleNodeDisplay(nodeID string) {
	node := tc.commands.GetNode(nodeID)
	if node == nil {
		return
	}

	// Find the node in the tree and update its display
	tc.updateNodeDisplayInTree(nil, nodeID, node)
}

// updateNodeDisplayInTree recursively finds and updates a node's display text in the tree store
func (tc *TextCleaner) updateNodeDisplayInTree(parentIter *gtk.TreeIter, nodeID string, node *PipelineNode) bool {
	var iter gtk.TreeIter
	hasIter := tc.treeStore.IterChildren(parentIter, &iter)

	for hasIter {
		val, _ := tc.treeStore.GetValue(&iter, 1)
		currentNodeID, _ := val.GetString()

		if currentNodeID == nodeID {
			// Found the node - update its display
			displayText := tc.getNodeDisplayText(node)
			// Only add emoji if in editing mode
			if tc.editingMode && nodeID == tc.commands.GetSelectedNodeID() {
				displayText = "✏️ " + displayText
			}
			tc.treeStore.SetValue(&iter, 0, displayText)
			return true
		}

		// Recursively search children
		if tc.updateNodeDisplayInTree(&iter, nodeID, node) {
			return true
		}

		hasIter = tc.treeStore.IterNext(&iter)
	}

	return false
}

// updateTreeEditingIndicators updates the display of nodes in the tree to show which node is being edited
func (tc *TextCleaner) updateTreeEditingIndicators() {
	selectedID := tc.commands.GetSelectedNodeID()
	if selectedID == "" {
		return // No node selected, nothing to highlight
	}

	// Walk through the tree store and update the display text for the selected node
	tc.updateNodeDisplayWithIndicator(nil, selectedID)
}

// updateNodeDisplayWithIndicator recursively updates tree nodes to add/remove editing indicator
func (tc *TextCleaner) updateNodeDisplayWithIndicator(parentIter *gtk.TreeIter, selectedID string) bool {
	var iter gtk.TreeIter
	hasIter := tc.treeStore.IterChildren(parentIter, &iter)

	for hasIter {
		// Get node ID from column 1
		val, _ := tc.treeStore.GetValue(&iter, 1)
		nodeID, _ := val.GetString()

		if nodeID == selectedID {
			// Found the selected node - update its display with indicator
			foundNode := tc.commands.GetNode(nodeID)
			if foundNode != nil {
				displayText := tc.getNodeDisplayText(foundNode)
				displayText = "✏️ " + displayText // Add pencil emoji indicator
				tc.treeStore.SetValue(&iter, 0, displayText)
			}
			return true
		}

		// Recursively search children
		if tc.updateNodeDisplayWithIndicator(&iter, selectedID) {
			return true
		}

		// Move to next sibling
		hasIter = tc.treeStore.IterNext(&iter)
	}

	return false
}

// buildTreePathForNodeID builds a GTK TreePath for a node anywhere in the tree
func (tc *TextCleaner) buildTreePathForNodeID(nodeID string) *gtk.TreePath {
	// Find path indices to this node
	pipeline := tc.commands.GetPipeline()
	indices := tc.findNodePathIndices(&pipeline, nodeID)
	if len(indices) == 0 {
		return nil
	}

	// Build path string like "0:1:2" for nested nodes
	pathStr := ""
	for i, idx := range indices {
		if i > 0 {
			pathStr += ":"
		}
		pathStr += fmt.Sprintf("%d", idx)
	}

	path, _ := gtk.TreePathNewFromString(pathStr)
	return path
}

// findNodePathIndices finds the indices path to a node (e.g., [0, 2] for root child 0, then grandchild 2)
func (tc *TextCleaner) findNodePathIndices(nodes *[]PipelineNode, nodeID string) []int {
	for i, node := range *nodes {
		if node.ID == nodeID {
			return []int{i}
		}

		// Search children
		if childIndices := tc.findNodePathIndices(&node.Children, nodeID); len(childIndices) > 0 {
			return append([]int{i}, childIndices...)
		}

		// Search else children
		if childIndices := tc.findNodePathIndices(&node.ElseChildren, nodeID); len(childIndices) > 0 {
			return append([]int{i}, childIndices...)
		}
	}
	return []int{}
}

func (tc *TextCleaner) addNodeToTree(node *PipelineNode, parentIter *gtk.TreeIter, nodeIdx int) {
	displayText := tc.getNodeDisplayText(node)

	var iter *gtk.TreeIter
	if parentIter == nil {
		iter = tc.treeStore.Append(nil)
	} else {
		iter = tc.treeStore.Append(parentIter)
	}

	// Store both display text (column 0) and node ID (column 1)
	tc.treeStore.SetValue(iter, 0, displayText)
	tc.treeStore.SetValue(iter, 1, node.ID)

	// Add children
	for _, child := range node.Children {
		tc.addNodeToTree(&child, iter, nodeIdx)
	}
}

func (tc *TextCleaner) getNodeDisplayText(node *PipelineNode) string {
	text := ""

	switch node.Type {
	case "operation":
		text = fmt.Sprintf("[OP] %s", node.Name)
		if node.Arg1 != "" {
			text += fmt.Sprintf(" (%s", node.Arg1)
			if node.Arg2 != "" {
				text += fmt.Sprintf(", %s", node.Arg2)
			}
			text += ")"
		}
	case "if":
		text = fmt.Sprintf("[IF] %s", node.Name)
	case "foreach":
		text = fmt.Sprintf("[LOOP] %s", node.Name)
	case "group":
		text = fmt.Sprintf("[GROUP] %s", node.Name)
	default:
		text = node.Name
	}

	return text
}

func (tc *TextCleaner) getNodeTypeFromUI(nodeTypeText string) string {
	switch nodeTypeText {
	case "Operation":
		return "operation"
	case "If (Conditional)":
		return "if"
	case "ForEachLine":
		return "foreach"
	case "Group":
		return "group"
	}
	return "operation"
}

// updateTextDisplay is called after core operations to update the output display
func (tc *TextCleaner) updateTextDisplay() {
	// Update output buffer from core
	tc.outputBuffer.SetText(tc.commands.GetOutputText())
}

// updateTextDisplayAtNode updates the output display to show text after processing through a specific node
// This allows users to see intermediate results as they navigate the pipeline
func (tc *TextCleaner) updateTextDisplayAtNode(nodeID string) {
	// Update output buffer with text processed up to the selected node
	tc.outputBuffer.SetText(tc.commands.GetOutputTextAtNode(nodeID))
}

func (tc *TextCleaner) processText() {
	// Get input text from GTK buffer
	startIter, endIter := tc.inputBuffer.GetBounds()
	input, _ := tc.inputBuffer.GetText(startIter, endIter, true)

	// Process via commands interface (works with both local core and socket wrapper)
	tc.commands.SetInputText(input)

	// Update output buffer - if a node is selected, show output at that node
	// Otherwise show the full pipeline output
	selectedNodeID := tc.commands.GetSelectedNodeID()
	if selectedNodeID != "" {
		tc.outputBuffer.SetText(tc.commands.GetOutputTextAtNode(selectedNodeID))
	} else {
		tc.outputBuffer.SetText(tc.commands.GetOutputText())
	}
}

func (tc *TextCleaner) copyToClipboard() {
	clipboard, err := gtk.ClipboardGet(gdk.GdkAtomIntern("CLIPBOARD", true))
	if err != nil {
		log.Println("Failed to get clipboard:", err)
		return
	}

	startIter, endIter := tc.outputBuffer.GetBounds()
	text, _ := tc.outputBuffer.GetText(startIter, endIter, true)

	clipboard.SetText(text)
}

// refreshUIFromCore is called when socket commands modify the core
// It refreshes all UI elements to reflect the current state of the core
func (tc *TextCleaner) refreshUIFromCore() {
	// Refresh the pipeline tree view to show any structural changes
	tc.refreshPipelineTree()

	// Refresh the output text display (in case text processing changed)
	tc.updateTextDisplay()

	// Update button states based on selection
	tc.updateButtonStates()

	// If a node is selected, refresh its display in the node controls
	selectedID := tc.commands.GetSelectedNodeID()
	if selectedID != "" {
		node := tc.commands.GetNode(selectedID)
		if node != nil {
			tc.loadNodeToUI(node)
		}
	}
}

// generateRandomSocketPath generates a random socket path in XDG_RUNTIME_DIR
func generateRandomSocketPath() string {
	// Use XDG_RUNTIME_DIR if available, otherwise fall back to /tmp
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = "/tmp"
	}

	// Generate a random 8-byte hex string
	randBytes := make([]byte, 8)
	if _, err := rand.Read(randBytes); err != nil {
		// Fallback to a simple random suffix if rand.Read fails
		randBytes = []byte(fmt.Sprintf("%d", time.Now().UnixNano()))
	}
	randomSuffix := hex.EncodeToString(randBytes)

	return filepath.Join(runtimeDir, fmt.Sprintf("textcleaner-%s.sock", randomSuffix))
}

// startHeadlessChildProcess spawns the current executable as a headless socket server
func startHeadlessChildProcess(socketPath string) (*os.Process, error) {
	// Get the path to the current executable
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Start the headless server in a child process
	cmd := exec.Command(exePath, "--headless", "--socket", socketPath)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start child process: %w", err)
	}

	return cmd.Process, nil
}

// waitForSocketServer waits for a socket server to become available
func waitForSocketServer(socketPath string, timeout time.Duration) (*SocketClient, error) {
	deadline := time.Now().Add(timeout)

	for {
		// Try to connect
		client, err := NewSocketClient(socketPath)
		if err == nil {
			return client, nil
		}

		// Check if we've exceeded the timeout
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for socket server at %s", socketPath)
		}

		// Wait a bit before retrying
		time.Sleep(100 * time.Millisecond)
	}
}
