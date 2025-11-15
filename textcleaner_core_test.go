package main

import (
	"strings"
	"testing"
)

// ============================================================================
// Node Management Tests
// ============================================================================

// TestCreateNode tests basic node creation
func TestCreateNode(t *testing.T) {
	core := NewTextCleanerCore()

	nodeID := core.CreateNode("operation", "Uppercase", "Uppercase", "", "", "")
	if nodeID != "node_0" {
		t.Errorf("Expected nodeID 'node_0', got '%s'", nodeID)
	}

	node := core.GetNode(nodeID)
	if node == nil {
		t.Fatal("Node should exist")
	}
	if node.Type != "operation" {
		t.Errorf("Expected type 'operation', got '%s'", node.Type)
	}
	if node.Operation != "Uppercase" {
		t.Errorf("Expected operation 'Uppercase', got '%s'", node.Operation)
	}
}

// TestCreateMultipleNodes tests creating multiple nodes
func TestCreateMultipleNodes(t *testing.T) {
	core := NewTextCleanerCore()

	id1 := core.CreateNode("operation", "Op1", "Uppercase", "", "", "")
	id2 := core.CreateNode("operation", "Op2", "Lowercase", "", "", "")
	id3 := core.CreateNode("operation", "Op3", "Replace Text", "a", "b", "")

	if id1 != "node_0" || id2 != "node_1" || id3 != "node_2" {
		t.Errorf("Expected sequential IDs, got %s, %s, %s", id1, id2, id3)
	}

	pipeline := core.GetPipeline()
	if len(pipeline) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(pipeline))
	}
}

// TestUpdateNode tests updating an existing node
func TestUpdateNode(t *testing.T) {
	core := NewTextCleanerCore()
	nodeID := core.CreateNode("operation", "Test", "Uppercase", "", "", "")

	err := core.UpdateNode(nodeID, "Updated", "Replace Text", "arg1", "arg2", "")
	if err != nil {
		t.Fatalf("Update should succeed, got error: %v", err)
	}

	node := core.GetNode(nodeID)
	if node.Name != "Updated" {
		t.Errorf("Expected name 'Updated', got '%s'", node.Name)
	}
	if node.Operation != "Replace Text" {
		t.Errorf("Expected operation 'Replace Text', got '%s'", node.Operation)
	}
	if node.Arg1 != "arg1" {
		t.Errorf("Expected Arg1 'arg1', got '%s'", node.Arg1)
	}
}

// TestUpdateNonexistentNode tests updating a node that doesn't exist
func TestUpdateNonexistentNode(t *testing.T) {
	core := NewTextCleanerCore()

	err := core.UpdateNode("nonexistent", "Name", "Op", "", "", "")
	if err == nil {
		t.Error("Expected error when updating nonexistent node")
	}
}

// TestDeleteNode tests deleting a root-level node
func TestDeleteNode(t *testing.T) {
	core := NewTextCleanerCore()
	id1 := core.CreateNode("operation", "Op1", "Uppercase", "", "", "")
	id2 := core.CreateNode("operation", "Op2", "Lowercase", "", "", "")

	err := core.DeleteNode(id1)
	if err != nil {
		t.Fatalf("Delete should succeed, got error: %v", err)
	}

	if core.GetNode(id1) != nil {
		t.Error("Deleted node should not exist")
	}

	if core.GetNode(id2) == nil {
		t.Error("Remaining node should still exist")
	}

	pipeline := core.GetPipeline()
	if len(pipeline) != 1 {
		t.Errorf("Expected 1 node after deletion, got %d", len(pipeline))
	}
}

// TestDeleteNonexistentNode tests deleting a node that doesn't exist
func TestDeleteNonexistentNode(t *testing.T) {
	core := NewTextCleanerCore()

	err := core.DeleteNode("nonexistent")
	if err == nil {
		t.Error("Expected error when deleting nonexistent node")
	}
}

// TestSelectNode tests selecting a node
func TestSelectNode(t *testing.T) {
	core := NewTextCleanerCore()
	nodeID := core.CreateNode("operation", "Test", "Uppercase", "", "", "")

	err := core.SelectNode(nodeID)
	if err != nil {
		t.Fatalf("Select should succeed, got error: %v", err)
	}

	if core.GetSelectedNodeID() != nodeID {
		t.Errorf("Expected selected node '%s', got '%s'", nodeID, core.GetSelectedNodeID())
	}
}

// TestSelectNonexistentNode tests selecting a node that doesn't exist
func TestSelectNonexistentNode(t *testing.T) {
	core := NewTextCleanerCore()

	err := core.SelectNode("nonexistent")
	if err == nil {
		t.Error("Expected error when selecting nonexistent node")
	}
}

// ============================================================================
// Child Node Tests
// ============================================================================

// TestAddChildNode tests adding a child to a parent node
func TestAddChildNode(t *testing.T) {
	core := NewTextCleanerCore()
	parentID := core.CreateNode("if", "IfNode", "", "", "", "pattern")

	childID, err := core.AddChildNode(parentID, "operation", "Child", "Uppercase", "", "", "")
	if err != nil {
		t.Fatalf("AddChild should succeed, got error: %v", err)
	}

	parent := core.GetNode(parentID)
	if parent == nil {
		t.Fatal("Parent node should exist")
	}

	if len(parent.Children) != 1 {
		t.Errorf("Parent should have 1 child, got %d", len(parent.Children))
	}

	child := core.GetNode(childID)
	if child == nil {
		t.Error("Child should exist")
	}
	if child.Type != "operation" {
		t.Errorf("Expected child type 'operation', got '%s'", child.Type)
	}
}

// TestAddChildToNonexistentParent tests adding a child when parent doesn't exist
func TestAddChildToNonexistentParent(t *testing.T) {
	core := NewTextCleanerCore()

	_, err := core.AddChildNode("nonexistent", "operation", "Child", "Op", "", "", "")
	if err == nil {
		t.Error("Expected error when adding child to nonexistent parent")
	}
}

// TestAddMultipleChildren tests adding multiple children to a parent
func TestAddMultipleChildren(t *testing.T) {
	core := NewTextCleanerCore()
	parentID := core.CreateNode("if", "IfNode", "", "", "", "pattern")

	for i := 0; i < 3; i++ {
		_, err := core.AddChildNode(parentID, "operation", "Child", "Uppercase", "", "", "")
		if err != nil {
			t.Fatalf("AddChild failed: %v", err)
		}
	}

	parent := core.GetNode(parentID)
	if len(parent.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(parent.Children))
	}
}

// ============================================================================
// Text Processing Tests
// ============================================================================

// TestSetInputText tests setting and retrieving input text
func TestSetInputText(t *testing.T) {
	core := NewTextCleanerCore()

	text := "Hello World"
	core.SetInputText(text)

	if core.GetInputText() != text {
		t.Errorf("Expected input '%s', got '%s'", text, core.GetInputText())
	}
}

// TestSimpleTextProcessing tests text processing with a single operation
func TestSimpleTextProcessing(t *testing.T) {
	core := NewTextCleanerCore()
	core.CreateNode("operation", "Uppercase", "Uppercase", "", "", "")

	core.SetInputText("hello world")

	output := core.GetOutputText()
	if output != "HELLO WORLD" {
		t.Errorf("Expected 'HELLO WORLD', got '%s'", output)
	}
}

// TestChainedOperations tests multiple operations in sequence
func TestChainedOperations(t *testing.T) {
	core := NewTextCleanerCore()
	core.CreateNode("operation", "Uppercase", "Uppercase", "", "", "")
	core.CreateNode("operation", "Replace Text", "Replace Text", "L", "X", "")

	core.SetInputText("hello world")

	output := core.GetOutputText()
	if output != "HEXXO WORXD" {
		t.Errorf("Expected 'HEXXO WORXD', got '%s'", output)
	}
}

// TestEmptyPipelineProcessing tests processing with no operations
func TestEmptyPipelineProcessing(t *testing.T) {
	core := NewTextCleanerCore()

	core.SetInputText("hello world")

	output := core.GetOutputText()
	if output != "hello world" {
		t.Errorf("Expected input unchanged, got '%s'", output)
	}
}

// TestIfNodeTrueBranch tests if node that matches the pattern
func TestIfNodeTrueBranch(t *testing.T) {
	core := NewTextCleanerCore()
	ifNodeID := core.CreateNode("if", "Check", "", "", "", "hello")
	core.AddChildNode(ifNodeID, "operation", "Uppercase", "Uppercase", "", "", "")

	core.SetInputText("hello world")

	output := core.GetOutputText()
	if output != "HELLO WORLD" {
		t.Errorf("Expected 'HELLO WORLD', got '%s'", output)
	}
}

// TestIfNodeFalseBranch tests if node that doesn't match the pattern
func TestIfNodeFalseBranch(t *testing.T) {
	core := NewTextCleanerCore()
	ifNodeID := core.CreateNode("if", "Check", "", "", "", "goodbye")
	core.AddChildNode(ifNodeID, "operation", "Uppercase", "Uppercase", "", "", "")

	core.SetInputText("hello world")

	output := core.GetOutputText()
	if output != "hello world" {
		t.Errorf("Expected 'hello world' (unchanged), got '%s'", output)
	}
}

// TestForEachLineOperation tests foreach line operation
func TestForEachLineOperation(t *testing.T) {
	core := NewTextCleanerCore()
	forEachID := core.CreateNode("foreach", "ForEach", "", "", "", "")
	core.AddChildNode(forEachID, "operation", "Uppercase", "Uppercase", "", "", "")

	core.SetInputText("hello\nworld")

	output := core.GetOutputText()
	expected := "HELLO\nWORLD"
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}

// TestGroupOperation tests group node structure
func TestGroupOperation(t *testing.T) {
	core := NewTextCleanerCore()
	groupID := core.CreateNode("group", "Group", "", "", "", "")
	core.AddChildNode(groupID, "operation", "Uppercase", "Uppercase", "", "", "")
	core.AddChildNode(groupID, "operation", "Replace Text", "Replace Text", "A", "B", "")

	core.SetInputText("apple banana")

	output := core.GetOutputText()
	if output != "BPPLE BBNBNB" {
		t.Errorf("Expected 'BPPLE BBNBNB', got '%s'", output)
	}
}

// ============================================================================
// Import/Export Tests
// ============================================================================

// TestExportPipeline tests exporting pipeline to JSON
func TestExportPipeline(t *testing.T) {
	core := NewTextCleanerCore()
	core.CreateNode("operation", "Uppercase", "Uppercase", "", "", "")

	exported, err := core.ExportPipeline()
	if err != nil {
		t.Fatalf("Export should succeed, got error: %v", err)
	}

	if exported == "" {
		t.Error("Exported JSON should not be empty")
	}

	if !strings.Contains(exported, "Uppercase") {
		t.Error("Exported JSON should contain operation name")
	}
}

// TestImportPipeline tests importing pipeline from JSON
func TestImportPipeline(t *testing.T) {
	core := NewTextCleanerCore()
	core.CreateNode("operation", "Uppercase", "Uppercase", "", "", "")

	exported, _ := core.ExportPipeline()

	core2 := NewTextCleanerCore()
	err := core2.ImportPipeline(exported)
	if err != nil {
		t.Fatalf("Import should succeed, got error: %v", err)
	}

	pipeline := core2.GetPipeline()
	if len(pipeline) != 1 {
		t.Errorf("Expected 1 node after import, got %d", len(pipeline))
	}

	if pipeline[0].Operation != "Uppercase" {
		t.Errorf("Expected operation 'Uppercase', got '%s'", pipeline[0].Operation)
	}
}

// TestRoundTripExportImport tests export/import round trip
func TestRoundTripExportImport(t *testing.T) {
	core := NewTextCleanerCore()
	core.CreateNode("operation", "Uppercase", "Uppercase", "", "", "")
	core.CreateNode("operation", "Replace", "Replace", "A", "B", "")

	core.SetInputText("apple apple")
	expectedOutput := core.GetOutputText()

	exported, _ := core.ExportPipeline()

	core2 := NewTextCleanerCore()
	core2.ImportPipeline(exported)
	core2.SetInputText("apple apple")
	actualOutput := core2.GetOutputText()

	if actualOutput != expectedOutput {
		t.Errorf("Round-trip failed: expected '%s', got '%s'", expectedOutput, actualOutput)
	}
}

// TestImportInvalidJSON tests importing invalid JSON
func TestImportInvalidJSON(t *testing.T) {
	core := NewTextCleanerCore()

	err := core.ImportPipeline("invalid json")
	if err == nil {
		t.Error("Expected error when importing invalid JSON")
	}
}

// ============================================================================
// Edge Cases Tests
// ============================================================================

// TestEmptyNameDefault tests that empty names get default values
func TestEmptyNameDefault(t *testing.T) {
	core := NewTextCleanerCore()

	nodeID := core.CreateNode("operation", "", "Uppercase", "", "", "")
	node := core.GetNode(nodeID)

	if node.Name == "" || node.Name == "[Empty]" {
		t.Errorf("Expected default name, got '%s'", node.Name)
	}
}

// TestConditionalNameDefault tests that if node gets default condition-based name
func TestConditionalNameDefault(t *testing.T) {
	core := NewTextCleanerCore()

	nodeID := core.CreateNode("if", "", "", "", "", "testpattern")
	node := core.GetNode(nodeID)

	if !strings.Contains(node.Name, "testpattern") {
		t.Errorf("Expected name to contain 'testpattern', got '%s'", node.Name)
	}
}

// TestComplexNestedStructure tests complex nested pipeline
func TestComplexNestedStructure(t *testing.T) {
	core := NewTextCleanerCore()

	// Create root nodes
	if1 := core.CreateNode("if", "IF1", "", "", "", "test")
	upper1, _ := core.AddChildNode(if1, "operation", "Upper", "Uppercase", "", "", "")

	if2 := core.CreateNode("if", "IF2", "", "", "", "hello")
	_, _ = core.AddChildNode(if2, "operation", "Lower", "Lowercase", "", "", "")

	// Add nested children
	core.AddChildNode(upper1, "operation", "Nested", "Replace", "A", "X", "")

	// Verify structure
	pipeline := core.GetPipeline()
	if len(pipeline) != 2 {
		t.Errorf("Expected 2 root nodes, got %d", len(pipeline))
	}

	if1Node := core.GetNode(if1)
	if len(if1Node.Children) != 1 {
		t.Errorf("Expected 1 child for if1, got %d", len(if1Node.Children))
	}
}

// TestNodeCounterAfterImport tests that node counter is reset properly after import
func TestNodeCounterAfterImport(t *testing.T) {
	core := NewTextCleanerCore()
	core.CreateNode("operation", "Op1", "Uppercase", "", "", "")

	exported, _ := core.ExportPipeline()

	core2 := NewTextCleanerCore()
	core2.ImportPipeline(exported)

	// Create a new node and verify it gets a unique ID
	newID := core2.CreateNode("operation", "Op2", "Uppercase", "", "", "")
	if newID == "node_0" {
		t.Errorf("New node should not reuse ID from imported pipeline, got %s", newID)
	}

	pipeline := core2.GetPipeline()
	if len(pipeline) != 2 {
		t.Errorf("Expected 2 nodes after import and create, got %d", len(pipeline))
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

// TestCompleteWorkflow tests a complete workflow from start to finish
func TestCompleteWorkflow(t *testing.T) {
	core := NewTextCleanerCore()

	// Create a complete pipeline
	if1 := core.CreateNode("if", "Has space", "", "", "", " ")
	core.AddChildNode(if1, "operation", "Replace space", "Replace Text", " ", "_", "")

	core.CreateNode("operation", "Uppercase", "Uppercase", "", "", "")

	// Process text
	core.SetInputText("hello world")
	output := core.GetOutputText()

	if output != "HELLO_WORLD" {
		t.Errorf("Expected 'HELLO_WORLD', got '%s'", output)
	}

	// Export and verify
	exported, _ := core.ExportPipeline()
	if !strings.Contains(exported, "Has space") {
		t.Error("Export should contain node names")
	}
}

// TestMultipleSelectionsAndUpdates tests selecting and updating multiple nodes
func TestMultipleSelectionsAndUpdates(t *testing.T) {
	core := NewTextCleanerCore()

	id1 := core.CreateNode("operation", "Op1", "Uppercase", "", "", "")
	id2 := core.CreateNode("operation", "Op2", "Lowercase", "", "", "")

	core.SelectNode(id1)
	if core.GetSelectedNodeID() != id1 {
		t.Error("Selected node should be id1")
	}

	core.UpdateNode(id1, "Updated1", "Replace", "a", "b", "")

	core.SelectNode(id2)
	if core.GetSelectedNodeID() != id2 {
		t.Error("Selected node should be id2")
	}

	core.UpdateNode(id2, "Updated2", "Uppercase", "", "", "")

	node1 := core.GetNode(id1)
	if node1.Name != "Updated1" {
		t.Errorf("Expected node1 name 'Updated1', got '%s'", node1.Name)
	}

	node2 := core.GetNode(id2)
	if node2.Name != "Updated2" {
		t.Errorf("Expected node2 name 'Updated2', got '%s'", node2.Name)
	}
}

// TestLargeTextProcessing tests processing of larger text
func TestLargeTextProcessing(t *testing.T) {
	core := NewTextCleanerCore()
	core.CreateNode("operation", "Uppercase", "Uppercase", "", "", "")

	largeText := strings.Repeat("hello world\n", 100)
	core.SetInputText(largeText)

	output := core.GetOutputText()
	expectedLine := "HELLO WORLD"
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 100 {
		t.Errorf("Expected 100 lines, got %d", len(lines))
	}

	for _, line := range lines {
		if line != expectedLine {
			t.Errorf("Expected line '%s', got '%s'", expectedLine, line)
		}
	}
}

// TestIndentNode tests indenting a node to become a child of previous sibling
func TestIndentNode(t *testing.T) {
	core := NewTextCleanerCore()

	// Create two root nodes
	id1 := core.CreateNode("operation", "Op1", "Uppercase", "", "", "")
	id2 := core.CreateNode("operation", "Op2", "Lowercase", "", "", "")

	// Indent id2 to become child of id1
	err := core.IndentNode(id2)
	if err != nil {
		t.Fatalf("Failed to indent node: %v", err)
	}

	// Verify id1 now has id2 as a child
	node1 := core.GetNode(id1)
	if len(node1.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(node1.Children))
	}

	if node1.Children[0].ID != id2 {
		t.Errorf("Expected child to be id2, got %s", node1.Children[0].ID)
	}

	// Verify pipeline no longer contains id2 at root
	pipeline := core.GetPipeline()
	if len(pipeline) != 1 {
		t.Errorf("Expected 1 root node, got %d", len(pipeline))
	}
}

// TestUnindentNode tests unindenting a node to become sibling of parent
func TestUnindentNode(t *testing.T) {
	core := NewTextCleanerCore()

	// Create root node
	id1 := core.CreateNode("operation", "Op1", "Uppercase", "", "", "")

	// Add child to id1
	id2, err := core.AddChildNode(id1, "operation", "Op2", "Lowercase", "", "", "")
	if err != nil {
		t.Fatalf("Failed to add child: %v", err)
	}

	// Unindent id2
	err = core.UnindentNode(id2)
	if err != nil {
		t.Fatalf("Failed to unindent node: %v", err)
	}

	// Verify id1 has no children
	node1 := core.GetNode(id1)
	if len(node1.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(node1.Children))
	}

	// Verify pipeline has 2 root nodes
	pipeline := core.GetPipeline()
	if len(pipeline) != 2 {
		t.Errorf("Expected 2 root nodes, got %d", len(pipeline))
	}

	// Verify id2 is now at root level after id1
	if pipeline[1].ID != id2 {
		t.Errorf("Expected second root node to be id2, got %s", pipeline[1].ID)
	}
}

// TestMoveNodeUp tests moving a node up in sibling list
func TestMoveNodeUp(t *testing.T) {
	core := NewTextCleanerCore()

	// Create parent
	parentID := core.CreateNode("operation", "Parent", "Uppercase", "", "", "")

	// Add two children
	child1, _ := core.AddChildNode(parentID, "operation", "Child1", "Lowercase", "", "", "")
	child2, _ := core.AddChildNode(parentID, "operation", "Child2", "Uppercase", "", "", "")

	// Move child2 up
	err := core.MoveNodeUp(child2)
	if err != nil {
		t.Fatalf("Failed to move node up: %v", err)
	}

	// Verify order is swapped
	parent := core.GetNode(parentID)
	if len(parent.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(parent.Children))
	}

	if parent.Children[0].ID != child2 {
		t.Errorf("Expected first child to be child2, got %s", parent.Children[0].ID)
	}

	if parent.Children[1].ID != child1 {
		t.Errorf("Expected second child to be child1, got %s", parent.Children[1].ID)
	}
}

// TestMoveNodeDown tests moving a node down in sibling list
func TestMoveNodeDown(t *testing.T) {
	core := NewTextCleanerCore()

	// Create parent
	parentID := core.CreateNode("operation", "Parent", "Uppercase", "", "", "")

	// Add two children
	child1, _ := core.AddChildNode(parentID, "operation", "Child1", "Lowercase", "", "", "")
	child2, _ := core.AddChildNode(parentID, "operation", "Child2", "Uppercase", "", "", "")

	// Move child1 down
	err := core.MoveNodeDown(child1)
	if err != nil {
		t.Fatalf("Failed to move node down: %v", err)
	}

	// Verify order is swapped
	parent := core.GetNode(parentID)
	if len(parent.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(parent.Children))
	}

	if parent.Children[0].ID != child2 {
		t.Errorf("Expected first child to be child2, got %s", parent.Children[0].ID)
	}

	if parent.Children[1].ID != child1 {
		t.Errorf("Expected second child to be child1, got %s", parent.Children[1].ID)
	}
}

// TestCanIndentNode tests the CanIndentNode predicate
func TestCanIndentNode(t *testing.T) {
	core := NewTextCleanerCore()

	id1 := core.CreateNode("operation", "Op1", "Uppercase", "", "", "")
	id2 := core.CreateNode("operation", "Op2", "Lowercase", "", "", "")

	// Can indent id2 (has previous sibling)
	if !core.CanIndentNode(id2) {
		t.Error("Should be able to indent id2")
	}

	// Cannot indent id1 (no previous sibling)
	if core.CanIndentNode(id1) {
		t.Error("Should not be able to indent id1")
	}

	// Cannot indent root level node that doesn't exist
	if core.CanIndentNode("nonexistent") {
		t.Error("Should not be able to indent nonexistent node")
	}
}

// TestCanUnindentNode tests the CanUnindentNode predicate
func TestCanUnindentNode(t *testing.T) {
	core := NewTextCleanerCore()

	parentID := core.CreateNode("operation", "Parent", "Uppercase", "", "", "")
	childID, _ := core.AddChildNode(parentID, "operation", "Child", "Lowercase", "", "", "")

	// Can unindent child
	if !core.CanUnindentNode(childID) {
		t.Error("Should be able to unindent child")
	}

	// Cannot unindent root level node
	if core.CanUnindentNode(parentID) {
		t.Error("Should not be able to unindent root level node")
	}
}
