# GUI Implementation Analysis - go-textcleaner

## Overview
The go-textcleaner application uses GTK3 (via gotk3 bindings) for its graphical interface. The GUI is built around a headless core (`TextCleanerCore`) that handles text processing logic, while the GUI (`TextCleaner` struct) manages the user interface.

---

## 1. Main GUI Structure

### Key GUI Component (TextCleaner struct)
**Location:** `/home/peter/work/go-textcleaner/main.go` lines 20-45

The `TextCleaner` struct contains:
- **Core reference:** `core *TextCleanerCore` - handles all text processing
- **Windows & views:**
  - `window` - Main GTK window
  - `inputView` / `outputView` - Text display areas
  - `pipelineTree` - Tree view showing node hierarchy
- **Control widgets:**
  - Combo boxes: `nodeTypeCombo`, `operationCombo`
  - Text entries: `nodeNameEntry`, `argument1`, `argument2`, `conditionEntry`
  - Buttons: `createNodeButton`, `editNodeButton`, `deleteNodeButton`, and tree operation buttons
- **Data storage:**
  - `treeStore` - GTK TreeStore backing the tree view
  - `selectedNode` - Currently selected tree path

### Layout
The UI is built with a paned layout:
1. **Left panel (Pipeline Panel):** Node controls + tree view
   - Controls: Node creation form with type/operation/name/args selectors
   - Tree: Hierarchical view of pipeline nodes
2. **Right panel (Text Panes):** Input and output areas
   - Input: Editable text area for test data
   - Output: Read-only display of processed results

**Key Layout Functions:**
- `BuildUI()` - Main UI builder (line 192)
- `createPipelinePanel()` - Left side panel creation (line 409)
- `createNodeControls()` - Node creation/editing controls (line 257)
- `createTextPane()` - Input/output text areas (line 466)

---

## 2. Tree View Structure

### Data Model
**TreeStore Structure:** 2 columns
- Column 0: Display text (human-readable node description)
- Column 1: Node ID (internal identifier for looking up in core)

**PipelineNode Structure** (`/home/peter/work/go-textcleaner/processor.go` lines 29-39):
```go
type PipelineNode struct {
    ID           string          // Unique identifier (e.g., "node_0", "node_1_child_0")
    Type         string          // "operation", "if", "foreach", "group"
    Name         string          // Display name
    Operation    string          // Operation type (e.g., "Uppercase", "Replace Text")
    Arg1, Arg2   string          // Operation arguments
    Condition    string          // For If nodes: regex pattern
    Children     []PipelineNode  // Child nodes (for nesting)
    ElseChildren []PipelineNode  // Else branch for If nodes
}
```

### Tree Building Functions
- `refreshPipelineTree()` (line 904) - Clears and rebuilds entire tree from core
  - Gets pipeline from core
  - Recursively adds nodes via `addNodeToTree()`
  - Expands all nodes by default
  
- `addNodeToTree()` (line 961) - Recursively adds node and children
  - Creates GTK TreeIter
  - Stores display text in column 0 and node ID in column 1
  - Recursively adds Children nodes
  
- `getNodeDisplayText()` (line 981) - Generates human-readable node text
  - Format: `[TYPE] Name (arg1, arg2)` for operations
  - Example: `[OP] Uppercase`, `[IF] Condition`, `[LOOP] For Each Line`

---

## 3. Node Selection & Interaction

### Selection Callbacks
**Lines 512-520 in setupEventHandlers():**

1. **"cursor-changed" signal** (line 513) → `updateTreeSelection()`
   - Fires when user clicks to select a node
   - Updates core's selected node ID
   - Enables/disables buttons based on selection state

2. **"row-activated" signal** (line 518) → `openNodeForEditing()`
   - Fires on double-click (tree row activation)
   - **Already implemented for double-click editing!**
   - Loads the node into the control form

### Selection Handling Functions

**`updateTreeSelection()` (lines 612-634)**
- Gets GTK tree selection
- Extracts node ID from column 1 of tree
- Updates core's selected node ID
- Calls `updateButtonStates()` to enable/disable buttons

**`openNodeForEditing()` (lines 591-610)** ← **KEY FUNCTION FOR DOUBLE-CLICK**
- Gets currently selected node from tree
- Retrieves node ID from column 1
- Finds node in core via `core.GetNode(nodeID)`
- Calls `loadNodeToUI(foundNode)` to populate controls
- Updates button states

**`loadNodeToUI()` (lines 636-668)**
- Populates the control form with node properties
- Sets node type combo based on node.Type
- Sets operation combo for operation nodes
- Sets name, arguments, and condition fields
- Calls `updateNodeTypeUI()` to show/hide appropriate fields

**`updateNodeTypeUI()` (lines 563-589)**
- Shows/hides control fields based on node type
- Operation nodes: show operation combo + arg1 + arg2
- If nodes: show condition field only
- ForEach/Group nodes: show minimal controls

---

## 4. Node Editing

### Current Edit Mechanism
The application already has a complete edit flow:

1. **User double-clicks a node in tree**
   - `row-activated` signal fires
   - `openNodeForEditing()` loads node into controls

2. **User modifies fields in control panel**
   - Changes name, operation, arguments, etc.

3. **User clicks "Update Node" button**
   - `updateSelectedNode()` is called (line 712)
   - Collects values from controls
   - Calls `core.UpdateNode(nodeID, ...)`
   - Refreshes tree and output display

**Key Edit Functions:**
- `updateSelectedNode()` (line 712) - Executes the update
- `clearNodeInputs()` (line 895) - Resets form fields
- `createNewNode()` (line 670) - For creating new nodes

### Edit Button State
- "Update Node" button is only enabled when a node is selected
- Button sensitivity managed by `updateButtonStates()` (line 882)

---

## 5. Button/Command Callbacks

All callbacks wired in `setupEventHandlers()` (line 496):

| Button | Signal | Handler Function | Action |
|--------|--------|-----------------|--------|
| Create Node | clicked | `createNewNode()` | Creates new root node |
| Update Node | clicked | `updateSelectedNode()` | Updates selected node |
| Delete | clicked | `deleteSelectedNode()` | Deletes selected node |
| Add Child | clicked | `addChildNode()` | Adds child to selected |
| Indent | clicked | `indentSelectedNode()` | Makes selected a child of prev sibling |
| Unindent | clicked | `unindentSelectedNode()` | Moves selected to parent's level |
| Move Up | clicked | `moveSelectedNodeUp()` | Moves selected up in sibling order |
| Move Down | clicked | `moveSelectedNodeDown()` | Moves selected down in sibling order |
| Tree row-activated | (double-click) | `openNodeForEditing()` | Loads node to editing form |
| Tree cursor-changed | (single-click) | `updateTreeSelection()` | Selects node, updates buttons |
| Input changed | changed | `processText()` | Processes input through pipeline |
| Copy button | clicked | `copyToClipboard()` | Copies output to clipboard |

---

## 6. Core Integration Points

### TextCleanerCore Methods Used
The GUI calls these core methods to perform operations:

**Node Management:**
- `core.CreateNode(type, name, operation, arg1, arg2, condition)` → node ID
- `core.UpdateNode(nodeID, name, operation, arg1, arg2, condition)` → error
- `core.DeleteNode(nodeID)` → error
- `core.AddChildNode(parentID, type, name, operation, arg1, arg2, condition)` → (nodeID, error)
- `core.GetNode(nodeID)` → *PipelineNode
- `core.GetPipeline()` → []PipelineNode

**Tree Operations:**
- `core.IndentNode(nodeID)` → error
- `core.UnindentNode(nodeID)` → error
- `core.MoveNodeUp(nodeID)` → error
- `core.MoveNodeDown(nodeID)` → error
- `core.CanIndentNode(nodeID)` → bool
- `core.CanUnindentNode(nodeID)` → bool
- `core.CanMoveNodeUp(nodeID)` → bool
- `core.CanMoveNodeDown(nodeID)` → bool

**Selection:**
- `core.SelectNode(nodeID)` - Sets selected node
- `core.GetSelectedNodeID()` → string

**Text Processing:**
- `core.SetInputText(text)` - Sets input to process
- `core.GetOutputText()` → string

---

## 7. Data Flow for Node Operations

### Example: Creating and Editing a Node

```
User clicks "Create Node"
    ↓
createNewNode() called
    ↓
Collects form values (type, name, operation, args)
    ↓
core.CreateNode(...) called → returns nodeID
    ↓
refreshPipelineTree() called
    ↓
Tree rebuilt from core.GetPipeline()
    ↓
GUI displays new node in tree
    ↓
updateTextDisplay() updates output
```

### Example: Double-Click Editing

```
User double-clicks node in tree
    ↓
row-activated signal fires
    ↓
openNodeForEditing() called
    ↓
Gets node ID from tree column 1
    ↓
core.GetNode(nodeID) retrieves node data
    ↓
loadNodeToUI(node) populates controls
    ↓
User modifies fields
    ↓
User clicks "Update Node"
    ↓
updateSelectedNode() called
    ↓
core.UpdateNode(nodeID, ...) updates core
    ↓
refreshPipelineTree() rebuilds tree view
    ↓
updateTextDisplay() shows new output
```

---

## 8. GTK/gotk3 Patterns Used

### Common GTK Operations
- **TreeView Selection:** `treeView.GetSelection()` → get selected row
- **TreeStore Data:** `treeStore.GetValue(iter, column)` → read value
- **TreeStore Data:** `treeStore.SetValue(iter, column, value)` → write value
- **Signals:** `widget.Connect(signal, callback)` → register event handler
- **Button State:** `button.SetSensitive(bool)` → enable/disable
- **Text Entry:** `entry.GetText()` / `entry.SetText()` → read/write
- **Combo Box:** `combo.GetActiveText()` / `combo.SetActive(index)` → get/set
- **Text Buffer:** `buffer.GetText(start, end, true)` → read all text

### Widget Types Used
- `gtk.Window` - Main window
- `gtk.Box` (VERTICAL/HORIZONTAL) - Layout containers
- `gtk.Paned` - Resizable panels
- `gtk.TreeView` - Node tree display
- `gtk.TreeStore` - Tree data backend
- `gtk.TreeViewColumn` - Column in tree
- `gtk.CellRendererText` - Text rendering in tree
- `gtk.TextView` - Multi-line text input/output
- `gtk.Entry` - Single-line text input
- `gtk.ComboBoxText` - Dropdown selection
- `gtk.Button` - Clickable buttons
- `gtk.ScrolledWindow` - Scrollable container
- `gtk.Frame` - Labeled container

---

## 9. Key Files & Line References

### Main GUI Implementation
**File:** `/home/peter/work/go-textcleaner/main.go`
- Lines 20-45: `TextCleaner` struct definition
- Lines 47-111: `main()` entry point
- Lines 192-255: `BuildUI()` - Main UI builder
- Lines 257-407: `createNodeControls()` - Control form creation
- Lines 409-464: `createPipelinePanel()` - Left panel with tree
- Lines 466-493: `createTextPane()` - Text input/output panes
- Lines 496-561: `setupEventHandlers()` - All signal connections
- Lines 563-589: `updateNodeTypeUI()` - Show/hide fields based on type
- Lines 591-610: **`openNodeForEditing()`** - Double-click handler
- Lines 612-634: `updateTreeSelection()` - Click selection handler
- Lines 636-668: `loadNodeToUI()` - Populate controls from node
- Lines 670-710: `createNewNode()` - Create operation
- Lines 712-755: `updateSelectedNode()` - Update operation
- Lines 757-773: `deleteSelectedNode()` - Delete operation
- Lines 775-816: `addChildNode()` - Add child operation
- Lines 818-880: Tree operation handlers (indent, unindent, move up/down)
- Lines 882-893: `updateButtonStates()` - Button sensitivity logic
- Lines 895-917: Tree building functions (refresh, add, display text)
- Lines 1053-1073: `refreshUIFromCore()` - Refresh all UI from core state

### Core Implementation
**File:** `/home/peter/work/go-textcleaner/textcleaner_core.go`
- Lines 1-27: `TextCleanerCore` struct definition
- Lines 34-68: `CreateNode()`
- Lines 70-102: `UpdateNode()`
- Lines 104-127: `DeleteNode()`
- Lines 129-170: `AddChildNode()`
- Lines 176+: Tree operations (Indent, Unindent, Move)

### Data Structures
**File:** `/home/peter/work/go-textcleaner/processor.go`
- Lines 22-26: `Operation` struct
- Lines 28-39: `PipelineNode` struct (with JSON tags)
- Lines 41+: `GetOperations()` - All available operations

---

## 10. How to Add Double-Click Editing Functionality

### Current State
**Double-click editing is ALREADY IMPLEMENTED!**

The mechanism exists at line 518 in `main.go`:
```go
tc.pipelineTree.Connect("row-activated", func() {
    tc.openNodeForEditing()
})
```

This signal fires when user double-clicks a tree row, calling `openNodeForEditing()` which:
1. Gets the selected node from tree
2. Loads it into the control form
3. User can edit in the form
4. User clicks "Update Node" to save

### Enhancement Ideas (If Further Improvement Needed)

#### 1. **Inline Tree Editing** (Edit Name Directly in Tree)
Currently, all editing happens in the control form below the tree. To add inline editing:

**Required Changes:**
- Replace GTK text cell renderer with editable renderer
- Add "edited" signal handler
- Update node in core when inline edit completes
- Refresh tree display

**Code Location:** `createPipelinePanel()` (around line 449)

```go
// Current:
renderer, _ := gtk.CellRendererTextNew()
column, _ := gtk.TreeViewColumnNewWithAttribute("Node", renderer, "text", 0)

// Enhanced:
renderer, _ := gtk.CellRendererTextNew()
renderer.SetProperty("editable", true)  // Make editable
renderer.Connect("edited", func(path, text string) {
    // Update node in core
    // Refresh tree
})
column, _ := gtk.TreeViewColumnNewWithAttribute("Node", renderer, "text", 0)
```

#### 2. **Modal Edit Dialog** (Pop-up Window for Editing)
Instead of using the form below tree, open a dialog:

**Required Changes:**
- Create `gtk.Dialog` or custom window
- Copy node controls into dialog
- Show dialog on double-click
- Update core on OK button
- Close dialog

**Code Location:** Add new function `openNodeEditDialog()` near `openNodeForEditing()`

#### 3. **Context Menu** (Right-Click Menu)
Add right-click menu for quick operations:

**Required Changes:**
- Add "button-press-event" handler to tree
- Create `gtk.Menu` with options
- Pop up menu on right-click
- Add handlers for menu items

---

## 11. Summary of Key Functions for Double-Click Implementation

| Function | Purpose | Location | Input | Output |
|----------|---------|----------|-------|--------|
| `openNodeForEditing()` | Load node for editing | 591 | (none, uses tree selection) | Populates form controls |
| `loadNodeToUI()` | Populate controls from node | 636 | *PipelineNode | (none, sets UI state) |
| `updateNodeTypeUI()` | Show/hide fields based on type | 563 | (none, reads combo) | Updates visibility |
| `updateTreeSelection()` | Handle tree selection | 612 | (none, from tree signal) | Updates core selected node |
| `updateSelectedNode()` | Save edited node to core | 712 | (none, reads form) | Updates core, refreshes UI |
| `refreshPipelineTree()` | Rebuild entire tree from core | 904 | (none, reads core) | Updates tree view |
| `updateButtonStates()` | Enable/disable buttons | 882 | (none, reads core) | Updates button sensitivity |
| `updateTextDisplay()` | Update output from core | 1023 | (none, reads core) | Updates output pane |

---

## 12. Architectural Insights

### Thread Safety
- **Core is thread-safe:** Uses `sync.RWMutex` for concurrent access
- **GUI is single-threaded:** GTK is not thread-safe, all UI ops on main thread
- **Socket integration:** Uses `glib.IdleAdd()` to queue GUI updates from socket thread

### Data Consistency
- Tree display always built from core state via `refreshPipelineTree()`
- Never modify tree directly; always update core first, then refresh tree
- This ensures GUI always matches core state

### Callback Pattern
- All user actions trigger core operations
- Core operations complete, then UI is refreshed
- No attempt to keep parallel state in GUI

---

## 13. Current Limitations & Future Improvements

See `TREE_IMPROVEMENTS.md` for complete roadmap. Key planned features:

1. **Else branch UI** - Currently not exposed in GUI
2. **Save/Load pipelines** - To files
3. **Copy/Paste nodes** - Duplicate configurations
4. **Right-click context menu** - Quick operations
5. **Keyboard shortcuts** - Power user features
6. **Undo/Redo** - Command history
7. **Inline name editing** - Edit directly in tree
8. **Drag-drop reordering** - Intuitive tree manipulation

---

## Summary

The go-textcleaner GUI is well-structured with clean separation between:
- **Business logic** (TextCleanerCore)
- **UI rendering** (TextCleaner struct and GTK)
- **Data models** (PipelineNode struct)

Double-click editing functionality **already exists** - users can double-click nodes to load them into the editing form, modify properties, and click "Update Node" to save.

The architecture makes it straightforward to add further UI enhancements following the existing patterns of:
1. Detect user action (signal)
2. Get data from core or UI
3. Perform operation on core
4. Refresh affected UI elements
