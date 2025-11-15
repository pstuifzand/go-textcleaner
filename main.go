package main

import (
	"fmt"
	"log"

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
	core              *TextCleanerCore // Headless core for business logic
	window            *gtk.Window
	inputView         *gtk.TextView
	outputView        *gtk.TextView
	inputBuffer       *gtk.TextBuffer
	outputBuffer      *gtk.TextBuffer
	copyButton        *gtk.Button
	pipelineTree      *gtk.TreeView
	treeStore         *gtk.TreeStore
	selectedNode      *gtk.TreePath
	nodeTypeCombo     *gtk.ComboBoxText
	operationCombo    *gtk.ComboBoxText
	argument1         *gtk.Entry
	argument2         *gtk.Entry
	conditionEntry    *gtk.Entry
	nodeNameEntry     *gtk.Entry
	createNodeButton  *gtk.Button
	editNodeButton    *gtk.Button
	deleteNodeButton  *gtk.Button
	indentButton      *gtk.Button
	unindentButton    *gtk.Button
	addChildButton    *gtk.Button
}

func main() {
	gtk.Init(nil)

	app := &TextCleaner{
		core: NewTextCleanerCore(),
	}
	app.BuildUI()

	gtk.Main()
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
	mainPaned.SetPosition(350) // Pipeline panel width

	// Create pipeline panel (left side)
	pipelinePanel := tc.createPipelinePanel()
	mainPaned.Add1(pipelinePanel)

	// Create horizontal paned for input/output (right side)
	textPaned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	textPaned.SetPosition((appWidth - 350) / 2)

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

	// Add child row
	addChildRow, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)

	addChildButton, _ := gtk.ButtonNewWithLabel("Add Child")
	tc.addChildButton = addChildButton
	addChildButton.SetSensitive(false)
	addChildRow.PackStart(addChildButton, true, true, 0)

	indentButton, _ := gtk.ButtonNewWithLabel("Indent")
	tc.indentButton = indentButton
	indentButton.SetSensitive(false)
	addChildRow.PackStart(indentButton, true, true, 0)

	unindentButton, _ := gtk.ButtonNewWithLabel("Unindent")
	tc.unindentButton = unindentButton
	unindentButton.SetSensitive(false)
	addChildRow.PackStart(unindentButton, true, true, 0)

	deleteNodeButton, _ := gtk.ButtonNewWithLabel("Delete")
	tc.deleteNodeButton = deleteNodeButton
	deleteNodeButton.SetSensitive(false)
	addChildRow.PackStart(deleteNodeButton, true, true, 0)

	controlsBox.PackStart(addChildRow, false, false, 0)

	return controlsBox
}

func (tc *TextCleaner) createPipelinePanel() *gtk.Box {
	panel, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	panel.SetMarginTop(5)
	panel.SetMarginBottom(5)
	panel.SetMarginStart(5)
	panel.SetMarginEnd(5)

	// Create paned layout for controls and tree
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	paned.SetPosition(320) // Controls panel height

	// ===== TOP SECTION: Node Controls =====
	nodeControls := tc.createNodeControls()
	controlsFrame, _ := gtk.FrameNew("Node Controls")
	controlsFrame.Add(nodeControls)
	paned.Add1(controlsFrame)

	// ===== BOTTOM SECTION: Pipeline Tree =====
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

	scrolledWindow.Add(treeView)
	treePanel.PackStart(scrolledWindow, true, true, 0)

	paned.Add2(treePanel)

	panel.PackStart(paned, true, true, 0)

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
	foundNode := tc.core.GetNode(nodeID)
	if foundNode != nil {
		tc.core.SelectNode(nodeID)
		tc.loadNodeToUI(foundNode)
		tc.updateButtonStates()
	}
}

func (tc *TextCleaner) updateTreeSelection() {
	selection, _ := tc.pipelineTree.GetSelection()
	_, iter, ok := selection.GetSelected()
	if !ok {
		tc.core.SelectNode("")
		tc.updateButtonStates()
		return
	}

	// Get the node ID from column 1 of the tree
	val, _ := tc.treeStore.GetValue(iter, 1)
	nodeID, _ := val.GetString()

	// Find the node by ID in the pipeline
	foundNode := tc.core.GetNode(nodeID)
	if foundNode != nil {
		tc.core.SelectNode(nodeID)
	} else {
		tc.core.SelectNode("")
	}

	tc.updateButtonStates()
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

	// Create node via core
	nodeID := tc.core.CreateNode(
		nodeType,
		nodeName,
		operation,
		arg1,
		arg2,
		condition,
	)

	// Refresh UI
	tc.refreshPipelineTree()
	tc.updateTextDisplay()

	// Select the newly created node
	tc.core.SelectNode(nodeID)
	node := tc.core.GetNode(nodeID)
	if node != nil {
		tc.loadNodeToUI(node)
	}
	tc.updateButtonStates()

	// Clear inputs
	tc.clearNodeInputs()
}

func (tc *TextCleaner) updateSelectedNode() {
	if tc.core.GetSelectedNodeID() == "" {
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

	// Update node via core
	err := tc.core.UpdateNode(
		tc.core.GetSelectedNodeID(),
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
	updatedNode := tc.core.GetNode(tc.core.GetSelectedNodeID())
	if updatedNode != nil {
		tc.loadNodeToUI(updatedNode)
	}
}

func (tc *TextCleaner) deleteSelectedNode() {
	if tc.core.GetSelectedNodeID() == "" {
		return
	}

	// Delete via core
	err := tc.core.DeleteNode(tc.core.GetSelectedNodeID())
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
	if tc.core.GetSelectedNodeID() == "" {
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

	// Add child node via core
	_, err := tc.core.AddChildNode(
		tc.core.GetSelectedNodeID(),
		nodeType,
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
	// This would move node to be a child of the previous node
	// Simplified for now
}

func (tc *TextCleaner) unindentSelectedNode() {
	// This would move node to be a sibling of its parent
	// Simplified for now
}

func (tc *TextCleaner) updateButtonStates() {
	hasSelection := tc.core.GetSelectedNodeID() != ""
	tc.editNodeButton.SetSensitive(hasSelection)
	tc.deleteNodeButton.SetSensitive(hasSelection)
	tc.addChildButton.SetSensitive(hasSelection)
	tc.indentButton.SetSensitive(false) // TODO: implement
	tc.unindentButton.SetSensitive(false) // TODO: implement
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
	pipeline := tc.core.GetPipeline()
	for i, node := range pipeline {
		tc.addNodeToTree(&node, nil, i)
	}

	tc.pipelineTree.ShowAll()

	// Expand all nodes
	tc.pipelineTree.ExpandAll()
}

// buildTreePathForNodeID builds a GTK TreePath for a node anywhere in the tree
func (tc *TextCleaner) buildTreePathForNodeID(nodeID string) *gtk.TreePath {
	// Find path indices to this node
	pipeline := tc.core.GetPipeline()
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
	tc.outputBuffer.SetText(tc.core.GetOutputText())
}

func (tc *TextCleaner) processText() {
	// Get input text from GTK buffer
	startIter, endIter := tc.inputBuffer.GetBounds()
	input, _ := tc.inputBuffer.GetText(startIter, endIter, true)

	// Process via core
	tc.core.SetInputText(input)

	// Update output buffer
	tc.outputBuffer.SetText(tc.core.GetOutputText())
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
