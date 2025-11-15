# Tree View & Tree Operations - Remaining Improvements

## Current Status
✅ **Implemented:**
- Tree structure with parent-child relationships
- Node creation with placeholder names
- Node selection and editing
- Node deletion
- Add child nodes
- Auto-name from operations
- All nodes expanded by default
- Real-time pipeline execution
- 114 text operations available

❌ **Not Implemented:**
- See categories below

---

## Priority 1: Core Node Operations

### 1.1 Indent/Unindent (Currently stubbed - lines 605-612 in main.go)
**Description:** Move nodes in/out of nesting hierarchy
- **Indent:** Make selected node a child of the previous sibling
- **Unindent:** Move selected node to be a sibling of its parent
**Impact:** Essential for tree organization without delete/recreate
**Complexity:** Medium - requires finding node relationships

### 1.2 Reorder Nodes (Move Up/Down)
**Description:** Change order of nodes at same level
- Move node up one position
- Move node down one position
**Impact:** High - users need to reorder operations
**Complexity:** Medium - requires array manipulation
**Suggested Implementation:** Add "Move Up" / "Move Down" buttons

### 1.3 Drag and Drop Reordering
**Description:** Visually drag nodes to reorder or nest
**Impact:** High - intuitive for rearrangement
**Complexity:** High - requires GTK drag-drop handlers
**Note:** Can be added after basic reordering

---

## Priority 2: Conditional Branch Support

### 2.1 Else Branch UI
**Description:** Display and configure else children of If nodes
- Show "Else:" section in tree under If nodes
- Allow adding children to else branch
- UI to switch between If branch and Else branch
**Impact:** Medium-High - needed for full conditional logic
**Complexity:** High - requires tree display refactoring
**Current Status:** ElseChildren field exists in PipelineNode but no UI

### 2.2 Else Condition Display
**Description:** Show else branches visually in tree
- Mark else nodes differently in tree display
- Allow selecting and editing else branch nodes
**Impact:** Medium - quality of life
**Complexity:** Medium - tree display logic

---

## Priority 3: Data Persistence

### 3.1 Save Pipeline to File
**Description:** Export pipeline to JSON file
- Button or menu: "Save Pipeline"
- File dialog to choose location
- Save pipeline array as JSON using PipelineNode.json tags
**Impact:** Critical - users lose work on close
**Complexity:** Low - JSON marshaling already in struct
**Suggested File Format:** `.pipeline.json`

### 3.2 Load Pipeline from File
**Description:** Import pipeline from JSON file
- Button or menu: "Load Pipeline"
- File dialog to select file
- Parse JSON and populate tree
- Validate loaded pipeline before executing
**Impact:** Critical - enables workflow saving
**Complexity:** Low - JSON unmarshaling
**Dependencies:** Requires Save Pipeline first

### 3.3 Auto-Save / Session Recovery
**Description:** Periodically save pipeline to temp file
- Auto-save every N seconds while editing
- On app start, offer to recover last session
**Impact:** Medium - safety feature
**Complexity:** Medium - file I/O + timer
**Optional Enhancement:** Undo/Redo history

### 3.4 Recent Pipelines
**Description:** Track and quickly load recent files
- "Recent Files" menu
- Store last 5-10 used pipelines
- Quick access to common workflows
**Impact:** Low - convenience
**Complexity:** Low

---

## Priority 4: Editing Improvements

### 4.1 Copy/Paste Nodes
**Description:** Duplicate node configurations
- Right-click: "Copy Node"
- Right-click: "Paste Node" (as sibling or child)
- Clipboard for node data (not system clipboard)
**Impact:** Medium - useful for repetitive pipelines
**Complexity:** Medium - clipboard management
**Suggested Implementation:** Store copied node in TextCleaner struct

### 4.2 Cut/Move Nodes
**Description:** Move nodes via cut/paste
- Right-click: "Cut Node"
- Right-click: "Paste Node"
- Similar to copy but removes from source
**Impact:** Medium - alternative to drag-drop
**Complexity:** Medium

### 4.3 Duplicate Subtree
**Description:** Copy entire branch with all children
- Right-click: "Duplicate Node and Children"
- Create deep copy of node and all descendants
**Impact:** Low-Medium - convenience
**Complexity:** Medium - recursive copy

### 4.4 Inline Name Editing in Tree
**Description:** Edit node name directly in tree
- Double-click node name to edit
- Press Enter to save, Esc to cancel
**Impact:** Low - convenience
**Complexity:** Medium - GTK editing mode

---

## Priority 5: User Interface Enhancements

### 5.1 Right-Click Context Menu
**Description:** Context menu for node operations
**Menu Items:**
- Edit Node (select it)
- Create Child
- Delete
- Copy/Cut/Paste
- Indent/Unindent
- Move Up/Down
**Impact:** Medium - discoverability
**Complexity:** Medium - GTK context menu handling

### 5.2 Keyboard Shortcuts
**Description:** Keyboard navigation and operations
**Suggested Shortcuts:**
- Ctrl+N: Create Node
- Ctrl+D: Delete Node
- Ctrl+C/V: Copy/Paste
- Delete key: Delete selected node
- Enter: Edit selected node
- Arrow keys: Navigate tree
**Impact:** Medium - power user feature
**Complexity:** Low-Medium

### 5.3 Node Icons/Visual Types
**Description:** Different icons for node types
- [OP] → Operation icon (green)
- [IF] → Conditional icon (blue)
- [LOOP] → ForEach icon (orange)
- [GROUP] → Group icon (gray)
**Impact:** Low - visual clarity
**Complexity:** Low - text styling only (or GTK icons)

### 5.4 Tree Search/Filter
**Description:** Search for operations in tree
- Search box: "Find operation..."
- Filter tree to show matching nodes
- Highlight matches
**Impact:** Low-Medium - large pipelines
**Complexity:** Medium

### 5.5 Tooltips
**Description:** Hover tooltips for nodes
- Show operation description
- Show arguments being used
**Impact:** Low - help
**Complexity:** Low - GTK tooltips

---

## Priority 6: History and Undo

### 6.1 Undo/Redo System
**Description:** Step backwards through changes
- Undo: Ctrl+Z
- Redo: Ctrl+Y
- Track all node operations
**Impact:** Medium-High - users expect this
**Complexity:** High - requires command history
**Suggested Implementation:** Command pattern with history stack

### 6.2 Change Indicators
**Description:** Show if pipeline has unsaved changes
- Window title: "TextCleaner - pipeline.json *" (asterisk = unsaved)
- Prompt on close if unsaved
**Impact:** Medium - prevents accidental loss
**Complexity:** Low

---

## Priority 7: Node Properties & Validation

### 7.1 Operation Description/Help
**Description:** Show what each operation does
- Help text in UI when operation selected
- Argument descriptions
- Usage examples
**Impact:** Low - documentation
**Complexity:** Low - add descriptions to GetOperations()

### 7.2 Argument Validation
**Description:** Validate arguments before allowing update
- Check regex patterns for If conditions
- Check numeric arguments are numbers
- Warn user of invalid input
**Impact:** Medium - error prevention
**Complexity:** Medium

### 7.3 Pipeline Validation
**Description:** Check pipeline structure is valid
- If nodes have children
- ForEach nodes have children
- No infinite loops (if not applicable)
- Suggest fixes for common errors
**Impact:** Low-Medium - safety
**Complexity:** Medium

---

## Priority 8: Advanced Features

### 8.1 Multi-Select
**Description:** Select multiple nodes at once
- Ctrl+Click to select multiple
- Shift+Click to select range
- Apply operations to all selected
**Impact:** Low - batch operations
**Complexity:** High - GTK selection changes

### 8.2 Node Groups/Categories
**Description:** Organize operations by category
- Group operations in UI dropdowns
- Tree categories (Text, Pattern, Math, etc.)
**Impact:** Low - organization
**Complexity:** Low-Medium - GetOperations() refactor

### 8.3 Pipeline Templates
**Description:** Save and reuse common patterns
- "Save as Template"
- "Load from Template"
- Common templates (clean text, extract emails, etc.)
**Impact:** Low - convenience
**Complexity:** Medium

### 8.4 Conditional Else-If Chains
**Description:** Multiple conditions in sequence
- Allow multiple if branches with different conditions
- UI to show if/else-if/else structure
**Impact:** Low - advanced
**Complexity:** High - significant data structure change

---

## Priority 9: Performance & Stability

### 9.1 Large Pipeline Optimization
**Description:** Handle very deep/wide trees
- Virtual scrolling for large trees
- Lazy loading of tree nodes
- Performance testing with 1000+ nodes
**Impact:** Low - most pipelines small
**Complexity:** High

### 9.2 Error Handling
**Description:** Graceful error handling
- Catch panics in operation execution
- Show error messages in output pane
- Prevent crashes from bad operations
**Impact:** Medium - stability
**Complexity:** Medium

### 9.3 Memory Management
**Description:** Prevent memory leaks
- Proper GTK object cleanup
- Test with very large inputs
**Impact:** Low - not critical for typical use
**Complexity:** Low-Medium

---

## Priority 10: Documentation & Help

### 10.1 User Manual
**Description:** In-app help and documentation
- Help button → embedded documentation
- Operation reference
- Tutorial walkthrough
**Impact:** Low - external docs can work
**Complexity:** Low

### 10.2 Example Pipelines
**Description:** Built-in example pipelines
- CSV to JSON converter
- Email extractor
- Log analyzer
- Load from examples in UI
**Impact:** Low-Medium - learning
**Complexity:** Low

---

## Implementation Roadmap (Suggested Priority Order)

### Phase 1 (Critical - Enable basic workflows)
1. Save/Load Pipeline (Priority 3.1, 3.2)
2. Indent/Unindent (Priority 1.1)
3. Reorder Nodes (Priority 1.2)

### Phase 2 (High Value - Improve usability)
1. Else branch UI (Priority 2.1)
2. Copy/Paste nodes (Priority 4.1)
3. Right-click context menu (Priority 5.1)
4. Change indicators + Undo/Redo basic (Priority 6.2, 6.1 basic)

### Phase 3 (Medium - Quality of life)
1. Keyboard shortcuts (Priority 5.2)
2. Node icons/visual types (Priority 5.3)
3. Operation descriptions (Priority 7.1)
4. Reorder via drag-drop (Priority 1.3)

### Phase 4 (Polish - Nice to have)
1. Context menu enhancements
2. Tree search/filter
3. Templates
4. Tooltips
5. Advanced validation

---

## Known Limitations

1. **ElseChildren not exposed in UI** - If nodes can have else branches in code but no UI to configure
2. **Single file workflow** - Only one pipeline open at a time
3. **No pipeline validation** - Can create invalid structures (if with no children, etc.)
4. **No operation descriptions** - 114 operations available but users must learn from use
5. **Tree selection only supports root level for some operations** - Fixed by using NodeID
6. **GTK3 limitations** - Some modern features not available (custom rendering, animations)

---

## Testing Checklist for Complete Implementation

- [ ] Create, edit, delete operations at all nesting levels
- [ ] Save pipeline with 50+ operations to file
- [ ] Load complex pipeline and verify execution
- [ ] Undo/Redo chain of 10+ operations
- [ ] Copy/Paste subtrees with children
- [ ] Reorder nodes and verify execution order
- [ ] Use If/Else branches properly
- [ ] Keyboard navigation through large tree
- [ ] Handle invalid input gracefully
- [ ] Performance with 200+ nodes
- [ ] Session recovery after crash
