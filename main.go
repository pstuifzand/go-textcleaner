package main

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
)

const (
	appTitle  = "TextCleaner"
	appWidth  = 1000
	appHeight = 600
)

type TextCleaner struct {
	window       *gtk.Window
	inputView    *gtk.TextView
	outputView   *gtk.TextView
	inputBuffer  *gtk.TextBuffer
	outputBuffer *gtk.TextBuffer
	operationBox *gtk.ComboBoxText
	argument1    *gtk.Entry
	argument2    *gtk.Entry
	copyButton   *gtk.Button
}

func main() {
	gtk.Init(nil)

	app := &TextCleaner{}
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

	// Create toolbar area
	toolbar := tc.createToolbar()
	mainBox.PackStart(toolbar, false, false, 0)

	// Create horizontal paned for input/output
	paned, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	paned.SetPosition(appWidth / 2)

	// Create input pane (left side)
	inputFrame := tc.createTextPane("Input", true)
	paned.Add1(inputFrame)

	// Create output pane (right side)
	outputFrame := tc.createTextPane("Output", false)
	paned.Add2(outputFrame)

	mainBox.PackStart(paned, true, true, 0)

	tc.window.Add(mainBox)
	tc.window.ShowAll()

	// Wire up event handlers for real-time processing
	tc.setupEventHandlers()
}

func (tc *TextCleaner) createToolbar() *gtk.Box {
	toolbar, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)

	// Operation label
	label, _ := gtk.LabelNew("Operation:")
	toolbar.PackStart(label, false, false, 0)

	// Operation combo box
	operationBox, _ := gtk.ComboBoxTextNew()
	tc.operationBox = operationBox

	// Add all available operations
	operations := GetOperations()
	for _, op := range operations {
		operationBox.AppendText(op.Name)
	}
	operationBox.SetActive(0)

	toolbar.PackStart(operationBox, false, false, 0)

	// Argument 1
	arg1Label, _ := gtk.LabelNew("Arg1:")
	toolbar.PackStart(arg1Label, false, false, 0)

	arg1Entry, _ := gtk.EntryNew()
	tc.argument1 = arg1Entry
	arg1Entry.SetWidthChars(20)
	toolbar.PackStart(arg1Entry, false, false, 0)

	// Argument 2
	arg2Label, _ := gtk.LabelNew("Arg2:")
	toolbar.PackStart(arg2Label, false, false, 0)

	arg2Entry, _ := gtk.EntryNew()
	tc.argument2 = arg2Entry
	arg2Entry.SetWidthChars(20)
	toolbar.PackStart(arg2Entry, false, false, 0)

	// Spacer
	spacer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	toolbar.PackStart(spacer, true, true, 0)

	// Copy button
	copyButton, _ := gtk.ButtonNewWithLabel("Copy to Clipboard")
	tc.copyButton = copyButton
	toolbar.PackStart(copyButton, false, false, 0)

	return toolbar
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

	// Argument 1 changed - reprocess
	tc.argument1.Connect("changed", func() {
		tc.processText()
	})

	// Argument 2 changed - reprocess
	tc.argument2.Connect("changed", func() {
		tc.processText()
	})

	// Operation changed - reprocess
	tc.operationBox.Connect("changed", func() {
		tc.processText()
	})

	// Copy button - copy output to clipboard
	tc.copyButton.Connect("clicked", func() {
		tc.copyToClipboard()
	})
}

// processText processes the input text with the selected operation and updates output
func (tc *TextCleaner) processText() {
	// Get input text
	startIter, endIter := tc.inputBuffer.GetBounds()
	input, _ := tc.inputBuffer.GetText(startIter, endIter, true)

	// Get selected operation
	operationName := tc.operationBox.GetActiveText()

	// Get arguments
	arg1, _ := tc.argument1.GetText()
	arg2, _ := tc.argument2.GetText()

	// Process the text
	output := ProcessText(input, operationName, arg1, arg2)

	// Update output buffer
	tc.outputBuffer.SetText(output)
}

// copyToClipboard copies the output text to clipboard
func (tc *TextCleaner) copyToClipboard() {
	clipboard, err := gtk.ClipboardGet(glib.GdkAtomIntern("CLIPBOARD", true))
	if err != nil {
		log.Println("Failed to get clipboard:", err)
		return
	}

	startIter, endIter := tc.outputBuffer.GetBounds()
	text, _ := tc.outputBuffer.GetText(startIter, endIter, true)

	clipboard.SetText(text)
}
