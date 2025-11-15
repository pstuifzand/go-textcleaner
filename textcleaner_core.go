package main

import (
	"encoding/json"
	"fmt"
	"sync"
)

// TextCleanerCore is the headless core for text processing with no GTK dependencies
type TextCleanerCore struct {
	mu             sync.RWMutex // Protects all fields below for thread-safe concurrent access
	pipeline       []PipelineNode
	selectedNodeID string
	inputText      string
	outputText     string
	nodeCounter    int // For generating unique IDs
}

// NewTextCleanerCore creates a new TextCleanerCore instance
func NewTextCleanerCore() *TextCleanerCore {
	return &TextCleanerCore{
		pipeline:       []PipelineNode{},
		selectedNodeID: "",
		inputText:      "",
		outputText:     "",
		nodeCounter:    0,
	}
}

// ============================================================================
// Node Management Methods
// ============================================================================

// CreateNode creates a new root-level node and returns its ID
func (tc *TextCleanerCore) CreateNode(nodeType, name, operation, arg1, arg2, condition string) string {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	nodeID := tc.generateNodeID()

	node := PipelineNode{
		ID:        nodeID,
		Type:      tc.normalizeNodeType(nodeType),
		Name:      name,
		Operation: operation,
		Arg1:      arg1,
		Arg2:      arg2,
		Condition: condition,
		Children:  []PipelineNode{},
	}

	// Set defaults if name is empty
	if node.Name == "" || node.Name == "[Empty]" {
		switch node.Type {
		case "operation":
			node.Name = operation
		case "if":
			node.Name = "If: " + condition
		case "foreach":
			node.Name = "For Each Line"
		case "group":
			node.Name = "Group"
		}
	}

	tc.pipeline = append(tc.pipeline, node)
	return nodeID
}

// UpdateNode updates an existing node by ID
func (tc *TextCleanerCore) UpdateNode(nodeID, name, operation, arg1, arg2, condition string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	node := tc.findNodeByID(nodeID)
	if node == nil {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	node.Name = name
	node.Operation = operation
	node.Arg1 = arg1
	node.Arg2 = arg2
	node.Condition = condition

	// Auto-fill name if empty
	if node.Name == "" || node.Name == "[Empty]" {
		switch node.Type {
		case "operation":
			node.Name = operation
		case "if":
			node.Name = "If: " + condition
		case "foreach":
			node.Name = "For Each Line"
		case "group":
			node.Name = "Group"
		}
	}

	tc.processText()
	return nil
}

// DeleteNode deletes a node by ID from anywhere in the pipeline
func (tc *TextCleanerCore) DeleteNode(nodeID string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Try to delete from root level
	for i := range tc.pipeline {
		if tc.pipeline[i].ID == nodeID {
			tc.pipeline = append(tc.pipeline[:i], tc.pipeline[i+1:]...)
			tc.selectedNodeID = ""
			tc.processText()
			return nil
		}
	}

	// Try to delete from nested children
	if tc.deleteNodeByID(&tc.pipeline, nodeID) {
		tc.selectedNodeID = ""
		tc.processText()
		return nil
	}

	return fmt.Errorf("node not found: %s", nodeID)
}

// AddChildNode adds a child node to a parent node
func (tc *TextCleanerCore) AddChildNode(parentID, nodeType, name, operation, arg1, arg2, condition string) (string, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	parentNode := tc.findNodeByID(parentID)
	if parentNode == nil {
		return "", fmt.Errorf("parent node not found: %s", parentID)
	}

	// Generate child node ID based on parent
	childID := fmt.Sprintf("%s_child_%d", parentID, len(parentNode.Children))

	child := PipelineNode{
		ID:        childID,
		Type:      tc.normalizeNodeType(nodeType),
		Name:      name,
		Operation: operation,
		Arg1:      arg1,
		Arg2:      arg2,
		Condition: condition,
		Children:  []PipelineNode{},
	}

	// Set defaults if name is empty
	if child.Name == "" || child.Name == "[Empty]" {
		switch child.Type {
		case "operation":
			child.Name = operation
		case "if":
			child.Name = "If: " + condition
		case "foreach":
			child.Name = "For Each Line"
		case "group":
			child.Name = "Group"
		}
	}

	parentNode.Children = append(parentNode.Children, child)
	tc.processText()
	return childID, nil
}

// ============================================================================
// Tree Operations (Indent, Unindent, Move Up, Move Down)
// ============================================================================

// IndentNode moves a node to become a child of its previous sibling
func (tc *TextCleanerCore) IndentNode(nodeID string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Check if node is at root level
	rootIdx := -1
	for i := range tc.pipeline {
		if tc.pipeline[i].ID == nodeID {
			rootIdx = i
			break
		}
	}

	if rootIdx >= 0 {
		// Handle root-level node
		if rootIdx == 0 {
			return fmt.Errorf("cannot indent: no previous sibling")
		}

		prevSibling := &tc.pipeline[rootIdx-1]
		node := tc.pipeline[rootIdx]

		// Move the node to the previous sibling's children
		prevSibling.Children = append(prevSibling.Children, node)

		// Remove from root pipeline
		tc.pipeline = append(tc.pipeline[:rootIdx], tc.pipeline[rootIdx+1:]...)

		tc.processText()
		return nil
	}

	// Handle nested node
	parentNode, idx := tc.findNodeParentAndIndex(&tc.pipeline, nodeID)
	if parentNode == nil || idx < 0 {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	// Check if there's a previous sibling
	if idx == 0 {
		return fmt.Errorf("cannot indent: no previous sibling")
	}

	prevSibling := &parentNode.Children[idx-1]
	node := parentNode.Children[idx]

	// Move the node to the previous sibling's children
	prevSibling.Children = append(prevSibling.Children, node)

	// Remove from parent's children
	parentNode.Children = append(parentNode.Children[:idx], parentNode.Children[idx+1:]...)

	tc.processText()
	return nil
}

// UnindentNode moves a node to become a sibling of its parent
func (tc *TextCleanerCore) UnindentNode(nodeID string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Find the node's parent
	parentNode, idx := tc.findNodeParentAndIndex(&tc.pipeline, nodeID)
	if parentNode == nil || idx < 0 {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	// Find the parent's parent and index
	grandParentNode, parentIdx := tc.findNodeParentAndIndex(&tc.pipeline, parentNode.ID)
	if grandParentNode == nil {
		// Parent is at root level
		// Move node to root level after the parent
		node := parentNode.Children[idx]
		parentNode.Children = append(parentNode.Children[:idx], parentNode.Children[idx+1:]...)

		// Find parent index in root pipeline
		parentRootIdx := -1
		for i := range tc.pipeline {
			if tc.pipeline[i].ID == parentNode.ID {
				parentRootIdx = i
				break
			}
		}

		if parentRootIdx >= 0 {
			// Insert after the parent
			newPipeline := append([]PipelineNode{}, tc.pipeline[:parentRootIdx+1]...)
			newPipeline = append(newPipeline, node)
			newPipeline = append(newPipeline, tc.pipeline[parentRootIdx+1:]...)
			tc.pipeline = newPipeline
		}
	} else {
		// Parent is not at root level
		// Move node from parent's children to grandparent's children
		node := parentNode.Children[idx]
		parentNode.Children = append(parentNode.Children[:idx], parentNode.Children[idx+1:]...)

		// Insert after the parent in grandparent's children
		grandChildrenList := &grandParentNode.Children
		newChildren := append([]PipelineNode{}, (*grandChildrenList)[:parentIdx+1]...)
		newChildren = append(newChildren, node)
		newChildren = append(newChildren, (*grandChildrenList)[parentIdx+1:]...)
		*grandChildrenList = newChildren
	}

	tc.processText()
	return nil
}

// MoveNodeUp moves a node up one position in its sibling list
func (tc *TextCleanerCore) MoveNodeUp(nodeID string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Check if node is at root level
	rootIdx := -1
	for i := range tc.pipeline {
		if tc.pipeline[i].ID == nodeID {
			rootIdx = i
			break
		}
	}

	if rootIdx >= 0 {
		// Handle root-level node
		if rootIdx == 0 {
			return fmt.Errorf("cannot move up: already at top")
		}

		// Swap with previous sibling
		tc.pipeline[rootIdx], tc.pipeline[rootIdx-1] = tc.pipeline[rootIdx-1], tc.pipeline[rootIdx]

		tc.processText()
		return nil
	}

	// Handle nested node
	parentNode, idx := tc.findNodeParentAndIndex(&tc.pipeline, nodeID)
	if parentNode == nil || idx < 0 {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	if idx == 0 {
		return fmt.Errorf("cannot move up: already at top")
	}

	// Swap with previous sibling
	parentNode.Children[idx], parentNode.Children[idx-1] = parentNode.Children[idx-1], parentNode.Children[idx]

	tc.processText()
	return nil
}

// MoveNodeDown moves a node down one position in its sibling list
func (tc *TextCleanerCore) MoveNodeDown(nodeID string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Check if node is at root level
	rootIdx := -1
	for i := range tc.pipeline {
		if tc.pipeline[i].ID == nodeID {
			rootIdx = i
			break
		}
	}

	if rootIdx >= 0 {
		// Handle root-level node
		if rootIdx >= len(tc.pipeline)-1 {
			return fmt.Errorf("cannot move down: already at bottom")
		}

		// Swap with next sibling
		tc.pipeline[rootIdx], tc.pipeline[rootIdx+1] = tc.pipeline[rootIdx+1], tc.pipeline[rootIdx]

		tc.processText()
		return nil
	}

	// Handle nested node
	parentNode, idx := tc.findNodeParentAndIndex(&tc.pipeline, nodeID)
	if parentNode == nil || idx < 0 {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	if idx >= len(parentNode.Children)-1 {
		return fmt.Errorf("cannot move down: already at bottom")
	}

	// Swap with next sibling
	parentNode.Children[idx], parentNode.Children[idx+1] = parentNode.Children[idx+1], parentNode.Children[idx]

	tc.processText()
	return nil
}

// CanIndentNode checks if a node can be indented
func (tc *TextCleanerCore) CanIndentNode(nodeID string) bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Check if node is at root level
	rootIdx := -1
	for i := range tc.pipeline {
		if tc.pipeline[i].ID == nodeID {
			rootIdx = i
			break
		}
	}

	if rootIdx >= 0 {
		// Can indent if not the first root node
		return rootIdx > 0
	}

	// Check nested nodes
	parentNode, idx := tc.findNodeParentAndIndex(&tc.pipeline, nodeID)
	return parentNode != nil && idx > 0
}

// CanUnindentNode checks if a node can be unindented
func (tc *TextCleanerCore) CanUnindentNode(nodeID string) bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	parentNode, _ := tc.findNodeParentAndIndex(&tc.pipeline, nodeID)
	// Can unindent if node has a parent (i.e., is not at root level)
	return parentNode != nil
}

// CanMoveNodeUp checks if a node can be moved up
func (tc *TextCleanerCore) CanMoveNodeUp(nodeID string) bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Check if node is at root level
	rootIdx := -1
	for i := range tc.pipeline {
		if tc.pipeline[i].ID == nodeID {
			rootIdx = i
			break
		}
	}

	if rootIdx >= 0 {
		return rootIdx > 0
	}

	// Check nested nodes
	parentNode, idx := tc.findNodeParentAndIndex(&tc.pipeline, nodeID)
	return parentNode != nil && idx > 0
}

// CanMoveNodeDown checks if a node can be moved down
func (tc *TextCleanerCore) CanMoveNodeDown(nodeID string) bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Check if node is at root level
	rootIdx := -1
	for i := range tc.pipeline {
		if tc.pipeline[i].ID == nodeID {
			rootIdx = i
			break
		}
	}

	if rootIdx >= 0 {
		return rootIdx >= 0 && rootIdx < len(tc.pipeline)-1
	}

	// Check nested nodes
	parentNode, idx := tc.findNodeParentAndIndex(&tc.pipeline, nodeID)
	return parentNode != nil && idx >= 0 && idx < len(parentNode.Children)-1
}

// SelectNode sets the currently selected node
func (tc *TextCleanerCore) SelectNode(nodeID string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	node := tc.findNodeByID(nodeID)
	if node == nil {
		return fmt.Errorf("node not found: %s", nodeID)
	}
	tc.selectedNodeID = nodeID
	return nil
}

// ============================================================================
// Text Processing Methods
// ============================================================================

// SetInputText sets the input text and processes it through the pipeline
func (tc *TextCleanerCore) SetInputText(text string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.inputText = text
	tc.processText()
}

// GetInputText returns the current input text
func (tc *TextCleanerCore) GetInputText() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return tc.inputText
}

// GetOutputText returns the current output text
func (tc *TextCleanerCore) GetOutputText() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return tc.outputText
}

// processText executes the pipeline on the input text and updates outputText
// This is a private method called automatically by SetInputText and other operations
func (tc *TextCleanerCore) processText() {
	output := tc.inputText
	for i := range tc.pipeline {
		output = ExecuteNode(&tc.pipeline[i], output)
	}
	tc.outputText = output
}

// ============================================================================
// Query Methods
// ============================================================================

// GetNode returns a node by ID, or nil if not found
func (tc *TextCleanerCore) GetNode(nodeID string) *PipelineNode {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return tc.findNodeByID(nodeID)
}

// GetSelectedNodeID returns the ID of the currently selected node
func (tc *TextCleanerCore) GetSelectedNodeID() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return tc.selectedNodeID
}

// GetPipeline returns a copy of the current pipeline
func (tc *TextCleanerCore) GetPipeline() []PipelineNode {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return append([]PipelineNode{}, tc.pipeline...)
}

// ============================================================================
// Import/Export Methods
// ============================================================================

// ExportPipeline exports the pipeline as a JSON string
func (tc *TextCleanerCore) ExportPipeline() (string, error) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	data, err := json.MarshalIndent(tc.pipeline, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ImportPipeline imports a pipeline from JSON string
func (tc *TextCleanerCore) ImportPipeline(jsonStr string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	var pipeline []PipelineNode
	if err := json.Unmarshal([]byte(jsonStr), &pipeline); err != nil {
		return err
	}

	tc.pipeline = pipeline
	tc.selectedNodeID = ""

	// Reset node counter to max ID + 1
	tc.nodeCounter = tc.calculateMaxNodeCounter() + 1

	tc.processText()
	return nil
}

// ============================================================================
// Helper Methods (Private)
// ============================================================================

// generateNodeID generates a unique node ID
func (tc *TextCleanerCore) generateNodeID() string {
	id := fmt.Sprintf("node_%d", tc.nodeCounter)
	tc.nodeCounter++
	return id
}

// normalizeNodeType converts UI node type names to internal representation
func (tc *TextCleanerCore) normalizeNodeType(nodeTypeText string) string {
	switch nodeTypeText {
	case "Operation":
		return "operation"
	case "If (Conditional)":
		return "if"
	case "ForEachLine":
		return "foreach"
	case "Group":
		return "group"
	default:
		// If it's already normalized, return as-is
		return nodeTypeText
	}
}

// findNodeByID searches for a node by ID in the entire pipeline (handles nested nodes)
func (tc *TextCleanerCore) findNodeByID(nodeID string) *PipelineNode {
	for i := range tc.pipeline {
		if node := tc.searchNodeByID(&tc.pipeline[i], nodeID); node != nil {
			return node
		}
	}
	return nil
}

// searchNodeByID recursively searches for a node by ID
func (tc *TextCleanerCore) searchNodeByID(node *PipelineNode, nodeID string) *PipelineNode {
	if node.ID == nodeID {
		return node
	}

	// Search in children
	for i := range node.Children {
		if found := tc.searchNodeByID(&node.Children[i], nodeID); found != nil {
			return found
		}
	}

	// Search in else children
	for i := range node.ElseChildren {
		if found := tc.searchNodeByID(&node.ElseChildren[i], nodeID); found != nil {
			return found
		}
	}

	return nil
}

// deleteNodeByID recursively deletes a node by ID from the pipeline
func (tc *TextCleanerCore) deleteNodeByID(nodes *[]PipelineNode, nodeID string) bool {
	for i := range *nodes {
		if (*nodes)[i].ID == nodeID {
			*nodes = append((*nodes)[:i], (*nodes)[i+1:]...)
			return true
		}

		// Search in children
		if tc.deleteNodeByID(&(*nodes)[i].Children, nodeID) {
			return true
		}

		// Search in else children
		if tc.deleteNodeByID(&(*nodes)[i].ElseChildren, nodeID) {
			return true
		}
	}
	return false
}

// findNodeIndexByID finds the index of a root-level node by ID
func (tc *TextCleanerCore) findNodeIndexByID(nodeID string) int {
	for i := range tc.pipeline {
		if tc.pipeline[i].ID == nodeID {
			return i
		}
	}
	return -1
}

// calculateMaxNodeCounter calculates the maximum node counter value from existing IDs
func (tc *TextCleanerCore) calculateMaxNodeCounter() int {
	maxCounter := 0
	tc.findMaxCounter(&tc.pipeline, &maxCounter)
	return maxCounter
}

// findMaxCounter recursively finds the maximum node counter in the pipeline
func (tc *TextCleanerCore) findMaxCounter(nodes *[]PipelineNode, maxCounter *int) {
	for _, node := range *nodes {
		// Extract counter from node ID like "node_5"
		var counter int
		if _, err := fmt.Sscanf(node.ID, "node_%d", &counter); err == nil {
			if counter > *maxCounter {
				*maxCounter = counter
			}
		}

		// Search in children
		tc.findMaxCounter(&node.Children, maxCounter)

		// Search in else children
		tc.findMaxCounter(&node.ElseChildren, maxCounter)
	}
}

// findNodeParentAndIndex finds a node's parent and its index in the parent's children list
// Returns (parentNode, index) or (nil, -1) if node is root-level or not found
func (tc *TextCleanerCore) findNodeParentAndIndex(nodes *[]PipelineNode, nodeID string) (*PipelineNode, int) {
	for i := range *nodes {
		// Check in children of current node
		for j := range (*nodes)[i].Children {
			if (*nodes)[i].Children[j].ID == nodeID {
				return &(*nodes)[i], j
			}
		}

		// Check in else children of current node
		for j := range (*nodes)[i].ElseChildren {
			if (*nodes)[i].ElseChildren[j].ID == nodeID {
				return &(*nodes)[i], j
			}
		}

		// Recursively search in nested children
		if parent, idx := tc.findNodeParentAndIndex(&(*nodes)[i].Children, nodeID); parent != nil {
			return parent, idx
		}

		// Recursively search in else children
		if parent, idx := tc.findNodeParentAndIndex(&(*nodes)[i].ElseChildren, nodeID); parent != nil {
			return parent, idx
		}
	}
	return nil, -1
}
