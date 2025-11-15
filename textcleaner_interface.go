package main

// TextCleanerCommands defines the interface for all TextCleaner operations.
// Both TextCleanerCore (direct implementation) and SocketClientCommands (socket wrapper)
// implement this interface, ensuring feature parity between direct and socket-based access.
type TextCleanerCommands interface {
	// =========================================================================
	// Node Management - Create, update, delete, and select nodes
	// =========================================================================

	// CreateNode creates a new root-level node and returns its ID
	CreateNode(nodeType, name, operation, arg1, arg2, condition string) string

	// UpdateNode updates an existing node's properties
	UpdateNode(nodeID, name, operation, arg1, arg2, condition string) error

	// DeleteNode removes a node and its subtree
	DeleteNode(nodeID string) error

	// AddChildNode adds a child node to a specified parent and returns its ID
	AddChildNode(parentID, nodeType, name, operation, arg1, arg2, condition string) (string, error)

	// SelectNode marks a node as the currently selected node
	SelectNode(nodeID string) error

	// =========================================================================
	// Tree Operations - Manipulate node structure and check constraints
	// =========================================================================

	// IndentNode moves a node to become a child of its previous sibling
	IndentNode(nodeID string) error

	// UnindentNode moves a node up to be a sibling of its parent
	UnindentNode(nodeID string) error

	// MoveNodeUp moves a node earlier in its parent's children list
	MoveNodeUp(nodeID string) error

	// MoveNodeDown moves a node later in its parent's children list
	MoveNodeDown(nodeID string) error

	// MoveNodeToPosition moves a node to a specific position with a new parent
	// parentID: "" for root level, or the parent node ID
	// position: index in the new parent's children list
	MoveNodeToPosition(nodeID, newParentID string, position int) error

	// CanIndentNode returns whether a node can be indented
	CanIndentNode(nodeID string) bool

	// CanUnindentNode returns whether a node can be unindented
	CanUnindentNode(nodeID string) bool

	// CanMoveNodeUp returns whether a node can be moved up
	CanMoveNodeUp(nodeID string) bool

	// CanMoveNodeDown returns whether a node can be moved down
	CanMoveNodeDown(nodeID string) bool

	// =========================================================================
	// Text Processing - Set and get input/output text
	// =========================================================================

	// SetInputText sets the input text to be processed by the pipeline
	SetInputText(text string)

	// GetInputText returns the current input text
	GetInputText() string

	// GetOutputText returns the result of processing input through the pipeline
	GetOutputText() string

	// =========================================================================
	// Query Operations - Retrieve pipeline information
	// =========================================================================

	// GetNode returns a specific node by ID, or nil if not found
	GetNode(nodeID string) *PipelineNode

	// GetSelectedNodeID returns the ID of the currently selected node, or empty string if none selected
	GetSelectedNodeID() string

	// GetPipeline returns all root-level nodes in the pipeline
	GetPipeline() []PipelineNode

	// =========================================================================
	// Import/Export - Serialize and deserialize pipeline
	// =========================================================================

	// ExportPipeline returns the pipeline as a JSON string
	ExportPipeline() (string, error)

	// ImportPipeline loads a pipeline from a JSON string
	ImportPipeline(jsonStr string) error
}
