package main

import (
	"encoding/json"
)

// Command represents a JSON command for AI agents
type Command struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
}

// Response represents a JSON response from command execution
type Response struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ExecuteCommand executes a JSON command and returns a JSON response
func (tc *TextCleanerCore) ExecuteCommand(cmdJSON string) string {
	var cmd Command
	if err := json.Unmarshal([]byte(cmdJSON), &cmd); err != nil {
		return tc.errorResponse("Invalid JSON: " + err.Error())
	}

	switch cmd.Action {
	case "create_node":
		return tc.cmdCreateNode(cmd.Params)
	case "update_node":
		return tc.cmdUpdateNode(cmd.Params)
	case "delete_node":
		return tc.cmdDeleteNode(cmd.Params)
	case "add_child_node":
		return tc.cmdAddChildNode(cmd.Params)
	case "select_node":
		return tc.cmdSelectNode(cmd.Params)
	case "set_input_text":
		return tc.cmdSetInputText(cmd.Params)
	case "get_input_text":
		return tc.cmdGetInputText(cmd.Params)
	case "get_output_text":
		return tc.cmdGetOutputText(cmd.Params)
	case "get_pipeline":
		return tc.cmdGetPipeline(cmd.Params)
	case "export_pipeline":
		return tc.cmdExportPipeline(cmd.Params)
	case "import_pipeline":
		return tc.cmdImportPipeline(cmd.Params)
	case "get_node":
		return tc.cmdGetNode(cmd.Params)
	case "get_selected_node_id":
		return tc.cmdGetSelectedNodeID(cmd.Params)
	case "list_nodes":
		return tc.cmdListNodes(cmd.Params)
	case "indent_node":
		return tc.cmdIndentNode(cmd.Params)
	case "unindent_node":
		return tc.cmdUnindentNode(cmd.Params)
	case "move_node_up":
		return tc.cmdMoveNodeUp(cmd.Params)
	case "move_node_down":
		return tc.cmdMoveNodeDown(cmd.Params)
	case "can_indent_node":
		return tc.cmdCanIndentNode(cmd.Params)
	case "can_unindent_node":
		return tc.cmdCanUnindentNode(cmd.Params)
	case "can_move_node_up":
		return tc.cmdCanMoveNodeUp(cmd.Params)
	case "can_move_node_down":
		return tc.cmdCanMoveNodeDown(cmd.Params)
	case "list_node_types":
		return tc.cmdListNodeTypes(cmd.Params)
	default:
		return tc.errorResponse("Unknown action: " + cmd.Action)
	}
}

// ============================================================================
// Command Handlers
// ============================================================================

// cmdCreateNode creates a new root-level node
func (tc *TextCleanerCore) cmdCreateNode(params map[string]interface{}) string {
	nodeType := getStr(params, "type", "")
	name := getStr(params, "name", "")
	operation := getStr(params, "operation", "")
	arg1 := getStr(params, "arg1", "")
	arg2 := getStr(params, "arg2", "")
	condition := getStr(params, "condition", "")

	if nodeType == "" {
		return tc.errorResponse("Missing required parameter: type")
	}

	nodeID := tc.CreateNode(nodeType, name, operation, arg1, arg2, condition)
	return tc.successResponse(map[string]interface{}{
		"node_id": nodeID,
	})
}

// cmdUpdateNode updates an existing node
func (tc *TextCleanerCore) cmdUpdateNode(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	name := getStr(params, "name", "")
	operation := getStr(params, "operation", "")
	arg1 := getStr(params, "arg1", "")
	arg2 := getStr(params, "arg2", "")
	condition := getStr(params, "condition", "")

	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	if err := tc.UpdateNode(nodeID, name, operation, arg1, arg2, condition); err != nil {
		return tc.errorResponse(err.Error())
	}

	return tc.successResponse(map[string]interface{}{
		"success": true,
	})
}

// cmdDeleteNode deletes a node by ID
func (tc *TextCleanerCore) cmdDeleteNode(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	if err := tc.DeleteNode(nodeID); err != nil {
		return tc.errorResponse(err.Error())
	}

	return tc.successResponse(map[string]interface{}{
		"success": true,
	})
}

// cmdAddChildNode adds a child node to a parent
func (tc *TextCleanerCore) cmdAddChildNode(params map[string]interface{}) string {
	parentID := getStr(params, "parent_id", "")
	nodeType := getStr(params, "type", "")
	name := getStr(params, "name", "")
	operation := getStr(params, "operation", "")
	arg1 := getStr(params, "arg1", "")
	arg2 := getStr(params, "arg2", "")
	condition := getStr(params, "condition", "")

	if parentID == "" {
		return tc.errorResponse("Missing required parameter: parent_id")
	}
	if nodeType == "" {
		return tc.errorResponse("Missing required parameter: type")
	}

	childID, err := tc.AddChildNode(parentID, nodeType, name, operation, arg1, arg2, condition)
	if err != nil {
		return tc.errorResponse(err.Error())
	}

	return tc.successResponse(map[string]interface{}{
		"node_id": childID,
	})
}

// cmdSelectNode sets the selected node
func (tc *TextCleanerCore) cmdSelectNode(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	if err := tc.SelectNode(nodeID); err != nil {
		return tc.errorResponse(err.Error())
	}

	return tc.successResponse(map[string]interface{}{
		"success": true,
	})
}

// cmdSetInputText sets the input text and processes it
func (tc *TextCleanerCore) cmdSetInputText(params map[string]interface{}) string {
	text := getStr(params, "text", "")
	tc.SetInputText(text)
	return tc.successResponse(map[string]interface{}{
		"success": true,
	})
}

// cmdGetInputText returns the current input text
func (tc *TextCleanerCore) cmdGetInputText(params map[string]interface{}) string {
	return tc.successResponse(map[string]interface{}{
		"text": tc.GetInputText(),
	})
}

// cmdGetOutputText returns the current output text
func (tc *TextCleanerCore) cmdGetOutputText(params map[string]interface{}) string {
	return tc.successResponse(map[string]interface{}{
		"output": tc.GetOutputText(),
	})
}

// cmdGetSelectedNodeID returns the currently selected node ID
func (tc *TextCleanerCore) cmdGetSelectedNodeID(params map[string]interface{}) string {
	nodeID := tc.GetSelectedNodeID()
	return tc.successResponse(map[string]interface{}{
		"node_id": nodeID,
	})
}

// cmdGetPipeline returns the full pipeline structure
func (tc *TextCleanerCore) cmdGetPipeline(params map[string]interface{}) string {
	pipeline := tc.GetPipeline()
	return tc.successResponse(map[string]interface{}{
		"pipeline": pipeline,
	})
}

// cmdExportPipeline exports the pipeline as JSON
func (tc *TextCleanerCore) cmdExportPipeline(params map[string]interface{}) string {
	jsonStr, err := tc.ExportPipeline()
	if err != nil {
		return tc.errorResponse(err.Error())
	}

	// Parse the JSON string back to return as an object
	var pipeline interface{}
	if err := json.Unmarshal([]byte(jsonStr), &pipeline); err != nil {
		return tc.errorResponse(err.Error())
	}

	return tc.successResponse(map[string]interface{}{
		"pipeline": pipeline,
	})
}

// cmdImportPipeline imports a pipeline from JSON
func (tc *TextCleanerCore) cmdImportPipeline(params map[string]interface{}) string {
	jsonData, ok := params["json"]
	if !ok {
		return tc.errorResponse("Missing required parameter: json")
	}

	// Convert the parameter to JSON string
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return tc.errorResponse("Invalid json parameter: " + err.Error())
	}

	if err := tc.ImportPipeline(string(jsonBytes)); err != nil {
		return tc.errorResponse(err.Error())
	}

	return tc.successResponse(map[string]interface{}{
		"success": true,
	})
}

// cmdGetNode returns a single node by ID
func (tc *TextCleanerCore) cmdGetNode(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	node := tc.GetNode(nodeID)
	if node == nil {
		return tc.errorResponse("node not found: " + nodeID)
	}

	return tc.successResponse(map[string]interface{}{
		"node": node,
	})
}

// cmdListNodes returns all root-level nodes
func (tc *TextCleanerCore) cmdListNodes(params map[string]interface{}) string {
	pipeline := tc.GetPipeline()
	return tc.successResponse(map[string]interface{}{
		"nodes": pipeline,
	})
}

// cmdIndentNode indents a node (makes it a child of previous sibling)
func (tc *TextCleanerCore) cmdIndentNode(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	if err := tc.IndentNode(nodeID); err != nil {
		return tc.errorResponse(err.Error())
	}

	return tc.successResponse(map[string]interface{}{
		"success": true,
	})
}

// cmdUnindentNode unindents a node (makes it a sibling of its parent)
func (tc *TextCleanerCore) cmdUnindentNode(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	if err := tc.UnindentNode(nodeID); err != nil {
		return tc.errorResponse(err.Error())
	}

	return tc.successResponse(map[string]interface{}{
		"success": true,
	})
}

// cmdMoveNodeUp moves a node up in its parent's children list
func (tc *TextCleanerCore) cmdMoveNodeUp(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	if err := tc.MoveNodeUp(nodeID); err != nil {
		return tc.errorResponse(err.Error())
	}

	return tc.successResponse(map[string]interface{}{
		"success": true,
	})
}

// cmdMoveNodeDown moves a node down in its parent's children list
func (tc *TextCleanerCore) cmdMoveNodeDown(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	if err := tc.MoveNodeDown(nodeID); err != nil {
		return tc.errorResponse(err.Error())
	}

	return tc.successResponse(map[string]interface{}{
		"success": true,
	})
}

// cmdCanIndentNode returns whether a node can be indented
func (tc *TextCleanerCore) cmdCanIndentNode(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	canIndent := tc.CanIndentNode(nodeID)
	return tc.successResponse(map[string]interface{}{
		"can_indent": canIndent,
	})
}

// cmdCanUnindentNode returns whether a node can be unindented
func (tc *TextCleanerCore) cmdCanUnindentNode(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	canUnindent := tc.CanUnindentNode(nodeID)
	return tc.successResponse(map[string]interface{}{
		"can_unindent": canUnindent,
	})
}

// cmdCanMoveNodeUp returns whether a node can be moved up
func (tc *TextCleanerCore) cmdCanMoveNodeUp(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	canMove := tc.CanMoveNodeUp(nodeID)
	return tc.successResponse(map[string]interface{}{
		"can_move_up": canMove,
	})
}

// cmdCanMoveNodeDown returns whether a node can be moved down
func (tc *TextCleanerCore) cmdCanMoveNodeDown(params map[string]interface{}) string {
	nodeID := getStr(params, "node_id", "")
	if nodeID == "" {
		return tc.errorResponse("Missing required parameter: node_id")
	}

	canMove := tc.CanMoveNodeDown(nodeID)
	return tc.successResponse(map[string]interface{}{
		"can_move_down": canMove,
	})
}

// cmdListNodeTypes returns available node types and operations
func (tc *TextCleanerCore) cmdListNodeTypes(params map[string]interface{}) string {
	nodeTypes := []string{"operation", "if", "foreach", "group"}

	operations := GetOperations()
	operationNames := make([]string, len(operations))
	for i, op := range operations {
		operationNames[i] = op.Name
	}

	return tc.successResponse(map[string]interface{}{
		"node_types": nodeTypes,
		"operations": operationNames,
	})
}

// ============================================================================
// Helper Functions
// ============================================================================

// getStr safely extracts a string parameter, with a default value
func getStr(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// toJSON converts a value to JSON string
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// successResponse creates a successful response
func (tc *TextCleanerCore) successResponse(result interface{}) string {
	resp := Response{
		Success: true,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

// errorResponse creates an error response
func (tc *TextCleanerCore) errorResponse(errorMsg string) string {
	resp := Response{
		Success: false,
		Error:   errorMsg,
	}
	data, _ := json.Marshal(resp)
	return string(data)
}
