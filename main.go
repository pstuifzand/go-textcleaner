package main

import (
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const (
	appTitle  = "TextCleaner"
	appWidth  = 1000
	appHeight = 600
)

// PipelineOperation represents a single operation in the processing pipeline
type PipelineOperation struct {
	Name      string
	Arg1      string
	Arg2      string
	LineBased bool // If true, operation is applied to each line individually
}

type TextCleaner struct {
	window            *gtk.Window
	inputView         *gtk.TextView
	outputView        *gtk.TextView
	inputBuffer       *gtk.TextBuffer
	outputBuffer      *gtk.TextBuffer
	operationBox      *gtk.ComboBoxText
	argument1         *gtk.Entry
	argument2         *gtk.Entry
	lineBasedCheckbox *gtk.CheckButton
	copyButton        *gtk.Button
	pipeline          []PipelineOperation
	pipelineListBox   *gtk.ListBox
	addButton         *gtk.Button
	updateButton      *gtk.Button
	removeButton      *gtk.Button
	moveUpButton      *gtk.Button
	moveDownButton    *gtk.Button
	selectedPipeline  int // -1 means no selection
}

func main() {
	gtk.Init(nil)

	app := &TextCleaner{}
	app.selectedPipeline = -1 // Initialize with no selection
	app.pipeline = []PipelineOperation{}
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

	// Create toolbar area with just copy button
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
	mainPaned.SetPosition(250) // Pipeline panel width

	// Create pipeline panel (left side)
	pipelinePanel := tc.createPipelinePanel()
	mainPaned.Add1(pipelinePanel)

	// Create horizontal paned for input/output (right side)
	textPaned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	textPaned.SetPosition((appWidth - 250) / 2)

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

	// Wire up event handlers for real-time processing
	tc.setupEventHandlers()
}

func (tc *TextCleaner) createOperationControls() *gtk.Box {
	controlsBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)

	// Operation label
	label, _ := gtk.LabelNew("Operation:")
	controlsBox.PackStart(label, false, false, 0)

	// Operation combo box
	operationBox, _ := gtk.ComboBoxTextNew()
	tc.operationBox = operationBox

	// Add all available operations
	operations := GetOperations()
	for _, op := range operations {
		operationBox.AppendText(op.Name)
	}
	operationBox.SetActive(0)

	controlsBox.PackStart(operationBox, false, false, 0)

	// Argument 1
	arg1Label, _ := gtk.LabelNew("Arg1:")
	controlsBox.PackStart(arg1Label, false, false, 0)

	arg1Entry, _ := gtk.EntryNew()
	tc.argument1 = arg1Entry
	arg1Entry.SetWidthChars(20)
	controlsBox.PackStart(arg1Entry, false, false, 0)

	// Argument 2
	arg2Label, _ := gtk.LabelNew("Arg2:")
	controlsBox.PackStart(arg2Label, false, false, 0)

	arg2Entry, _ := gtk.EntryNew()
	tc.argument2 = arg2Entry
	arg2Entry.SetWidthChars(20)
	controlsBox.PackStart(arg2Entry, false, false, 0)

	// Line-based checkbox
	lineBasedCheckbox, _ := gtk.CheckButtonNewWithLabel("Line-based")
	tc.lineBasedCheckbox = lineBasedCheckbox
	controlsBox.PackStart(lineBasedCheckbox, false, false, 0)

	// Add to Pipeline button
	addButton, _ := gtk.ButtonNewWithLabel("Add to Pipeline")
	tc.addButton = addButton
	controlsBox.PackStart(addButton, false, false, 0)

	// Update button (hidden/shown based on selection)
	updateButton, _ := gtk.ButtonNewWithLabel("Update Selected")
	tc.updateButton = updateButton
	updateButton.SetSensitive(false)
	controlsBox.PackStart(updateButton, false, false, 0)

	return controlsBox
}

func (tc *TextCleaner) createPipelinePanel() *gtk.Box {
	panel, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)

	// Add operation controls at the top
	operationControls := tc.createOperationControls()
	panel.PackStart(operationControls, false, false, 0)

	// Title label
	titleLabel, _ := gtk.LabelNew("Operations Pipeline")
	titleLabel.SetMarkup("<b>Operations Pipeline</b>")
	panel.PackStart(titleLabel, false, false, 5)

	// Scrolled window for the list
	scrolledWindow, _ := gtk.ScrolledWindowNew(nil, nil)
	scrolledWindow.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scrolledWindow.SetSizeRequest(200, -1)

	// ListBox for operations
	listBox, _ := gtk.ListBoxNew()
	tc.pipelineListBox = listBox
	listBox.SetSelectionMode(gtk.SELECTION_SINGLE)

	scrolledWindow.Add(listBox)
	panel.PackStart(scrolledWindow, true, true, 0)

	// Button box for controls
	buttonBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)

	// Remove button
	removeButton, _ := gtk.ButtonNewFromIconName("list-remove", gtk.ICON_SIZE_BUTTON)
	tc.removeButton = removeButton
	removeButton.SetSensitive(false)
	removeButton.SetTooltipText("Remove")
	buttonBox.PackStart(removeButton, true, true, 0)

	// Move Up button
	moveUpButton, _ := gtk.ButtonNewFromIconName("go-up", gtk.ICON_SIZE_BUTTON)
	tc.moveUpButton = moveUpButton
	moveUpButton.SetSensitive(false)
	moveUpButton.SetTooltipText("Move Up")
	buttonBox.PackStart(moveUpButton, true, true, 0)

	// Move Down button
	moveDownButton, _ := gtk.ButtonNewFromIconName("go-down", gtk.ICON_SIZE_BUTTON)
	tc.moveDownButton = moveDownButton
	moveDownButton.SetSensitive(false)
	moveDownButton.SetTooltipText("Move Down")
	buttonBox.PackStart(moveDownButton, true, true, 0)

	panel.PackStart(buttonBox, false, false, 0)

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

// setupEventHandlers wires up all event handlers for real-time processing
func (tc *TextCleaner) setupEventHandlers() {
	// Input buffer changed - process text in real-time
	tc.inputBuffer.Connect("changed", func() {
		tc.processText()
	})

	// Copy button - copy output to clipboard
	tc.copyButton.Connect("clicked", func() {
		tc.copyToClipboard()
	})

	// Add to Pipeline button
	tc.addButton.Connect("clicked", func() {
		tc.addToPipeline()
	})

	// Pipeline list selection
	tc.pipelineListBox.Connect("row-selected", func() {
		tc.updatePipelineSelection()
	})

	// Pipeline list double-click to edit
	tc.pipelineListBox.Connect("row-activated", func(listBox *gtk.ListBox, row *gtk.ListBoxRow) {
		tc.loadOperationForEdit()
	})

	// Update button
	tc.updateButton.Connect("clicked", func() {
		tc.updateSelectedOperation()
	})

	// Remove button
	tc.removeButton.Connect("clicked", func() {
		tc.removeFromPipeline()
	})

	// Move Up button
	tc.moveUpButton.Connect("clicked", func() {
		tc.moveOperationUp()
	})

	// Move Down button
	tc.moveDownButton.Connect("clicked", func() {
		tc.moveOperationDown()
	})
}

// processText processes the input text through the pipeline and updates output
func (tc *TextCleaner) processText() {
	// Get input text
	startIter, endIter := tc.inputBuffer.GetBounds()
	input, _ := tc.inputBuffer.GetText(startIter, endIter, true)

	// Process through pipeline
	output := input
	for _, pipeOp := range tc.pipeline {
		output = ProcessTextWithMode(output, pipeOp.Name, pipeOp.Arg1, pipeOp.Arg2, pipeOp.LineBased)
	}

	// Update output buffer
	tc.outputBuffer.SetText(output)
}

// copyToClipboard copies the output text to clipboard
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

// addToPipeline adds the current operation to the pipeline
func (tc *TextCleaner) addToPipeline() {
	operationName := tc.operationBox.GetActiveText()
	arg1, _ := tc.argument1.GetText()
	arg2, _ := tc.argument2.GetText()
	lineBased := tc.lineBasedCheckbox.GetActive()

	pipeOp := PipelineOperation{
		Name:      operationName,
		Arg1:      arg1,
		Arg2:      arg2,
		LineBased: lineBased,
	}

	tc.pipeline = append(tc.pipeline, pipeOp)
	tc.refreshPipelineList()
	tc.processText()

	// Clear arguments and checkbox for next operation
	tc.argument1.SetText("")
	tc.argument2.SetText("")
	tc.lineBasedCheckbox.SetActive(false)
}

// removeFromPipeline removes the selected operation from the pipeline
func (tc *TextCleaner) removeFromPipeline() {
	if tc.selectedPipeline < 0 || tc.selectedPipeline >= len(tc.pipeline) {
		return
	}

	// Remove the operation
	tc.pipeline = append(tc.pipeline[:tc.selectedPipeline], tc.pipeline[tc.selectedPipeline+1:]...)
	tc.selectedPipeline = -1
	tc.refreshPipelineList()
	tc.processText()
}

// moveOperationUp moves the selected operation up in the pipeline
func (tc *TextCleaner) moveOperationUp() {
	if tc.selectedPipeline <= 0 || tc.selectedPipeline >= len(tc.pipeline) {
		return
	}

	// Swap with previous operation
	tc.pipeline[tc.selectedPipeline], tc.pipeline[tc.selectedPipeline-1] = tc.pipeline[tc.selectedPipeline-1], tc.pipeline[tc.selectedPipeline]

	tc.selectedPipeline--
	tc.refreshPipelineList()
	tc.processText()
}

// moveOperationDown moves the selected operation down in the pipeline
func (tc *TextCleaner) moveOperationDown() {
	if tc.selectedPipeline < 0 || tc.selectedPipeline >= len(tc.pipeline)-1 {
		return
	}

	// Swap with next operation
	tc.pipeline[tc.selectedPipeline], tc.pipeline[tc.selectedPipeline+1] = tc.pipeline[tc.selectedPipeline+1], tc.pipeline[tc.selectedPipeline]

	tc.selectedPipeline++
	tc.refreshPipelineList()
	tc.processText()
}

// refreshPipelineList refreshes the pipeline list display
func (tc *TextCleaner) refreshPipelineList() {
	// Clear existing rows
	tc.pipelineListBox.GetChildren().Foreach(func(item interface{}) {
		widget := item.(*gtk.Widget)
		tc.pipelineListBox.Remove(widget)
	})

	// Add new rows
	for _, pipeOp := range tc.pipeline {
		label := pipeOp.Name
		if pipeOp.Arg1 != "" {
			label += " (" + pipeOp.Arg1
			if pipeOp.Arg2 != "" {
				label += ", " + pipeOp.Arg2
			}
			label += ")"
		}
		if pipeOp.LineBased {
			label += " [Line-based]"
		}

		row, _ := gtk.LabelNew(label)
		row.SetXAlign(0) // Left align
		row.SetMarginStart(5)
		row.SetMarginEnd(5)
		row.SetMarginTop(3)
		row.SetMarginBottom(3)

		tc.pipelineListBox.Add(row)
	}

	tc.pipelineListBox.ShowAll()

	// Restore selection if valid
	if tc.selectedPipeline >= 0 && tc.selectedPipeline < len(tc.pipeline) {
		row := tc.pipelineListBox.GetRowAtIndex(tc.selectedPipeline)
		tc.pipelineListBox.SelectRow(row)
	}

	tc.updateButtonStates()
}

// updatePipelineSelection updates the selected pipeline index
func (tc *TextCleaner) updatePipelineSelection() {
	selectedRow := tc.pipelineListBox.GetSelectedRow()
	if selectedRow != nil {
		tc.selectedPipeline = selectedRow.GetIndex()
	} else {
		tc.selectedPipeline = -1
	}
	tc.updateButtonStates()
}

// updateButtonStates updates the enabled/disabled state of pipeline buttons
func (tc *TextCleaner) updateButtonStates() {
	hasSelection := tc.selectedPipeline >= 0 && tc.selectedPipeline < len(tc.pipeline)
	tc.removeButton.SetSensitive(hasSelection)
	tc.updateButton.SetSensitive(hasSelection)

	canMoveUp := hasSelection && tc.selectedPipeline > 0
	tc.moveUpButton.SetSensitive(canMoveUp)

	canMoveDown := hasSelection && tc.selectedPipeline < len(tc.pipeline)-1
	tc.moveDownButton.SetSensitive(canMoveDown)
}

// loadOperationForEdit loads the selected operation into the toolbar for editing
func (tc *TextCleaner) loadOperationForEdit() {
	if tc.selectedPipeline < 0 || tc.selectedPipeline >= len(tc.pipeline) {
		return
	}

	pipeOp := tc.pipeline[tc.selectedPipeline]

	// Find and set the operation in the combo box
	operations := GetOperations()
	for i, op := range operations {
		if op.Name == pipeOp.Name {
			tc.operationBox.SetActive(i)
			break
		}
	}

	// Set the arguments
	tc.argument1.SetText(pipeOp.Arg1)
	tc.argument2.SetText(pipeOp.Arg2)
	tc.lineBasedCheckbox.SetActive(pipeOp.LineBased)
}

// updateSelectedOperation updates the selected operation with current toolbar values
func (tc *TextCleaner) updateSelectedOperation() {
	if tc.selectedPipeline < 0 || tc.selectedPipeline >= len(tc.pipeline) {
		return
	}

	operationName := tc.operationBox.GetActiveText()
	arg1, _ := tc.argument1.GetText()
	arg2, _ := tc.argument2.GetText()
	lineBased := tc.lineBasedCheckbox.GetActive()

	tc.pipeline[tc.selectedPipeline] = PipelineOperation{
		Name:      operationName,
		Arg1:      arg1,
		Arg2:      arg2,
		LineBased: lineBased,
	}

	tc.refreshPipelineList()
	tc.processText()

	// Clear arguments and checkbox
	tc.argument1.SetText("")
	tc.argument2.SetText("")
	tc.lineBasedCheckbox.SetActive(false)
}
