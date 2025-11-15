package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// SocketClientCommands wraps a SocketClient to implement the TextCleanerCommands interface.
// This allows GUI code to use the same interface whether connected to a socket server
// or using TextCleanerCore directly.
type SocketClientCommands struct {
	client *SocketClient
}

// NewSocketClientCommands creates a new socket client wrapper
func NewSocketClientCommands(client *SocketClient) *SocketClientCommands {
	return &SocketClientCommands{client: client}
}

// ============================================================================
// Node Management Methods
// ============================================================================

// CreateNode implements TextCleanerCommands.CreateNode
func (s *SocketClientCommands) CreateNode(nodeType, name, operation, arg1, arg2, condition string) string {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "create_node",
		"params": map[string]interface{}{
			"type":      nodeType,
			"name":      name,
			"operation": operation,
			"arg1":      arg1,
			"arg2":      arg2,
			"condition": condition,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("CreateNode socket error: %v", err)
		return ""
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if nodeID, ok := result["node_id"].(string); ok {
				return nodeID
			}
		}
	}

	if errMsg, ok := resp["error"].(string); ok {
		log.Printf("CreateNode error: %s", errMsg)
	}

	return ""
}

// UpdateNode implements TextCleanerCommands.UpdateNode
func (s *SocketClientCommands) UpdateNode(nodeID, name, operation, arg1, arg2, condition string) error {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "update_node",
		"params": map[string]interface{}{
			"node_id":   nodeID,
			"name":      name,
			"operation": operation,
			"arg1":      arg1,
			"arg2":      arg2,
			"condition": condition,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		return nil
	}

	if errMsg, ok := resp["error"].(string); ok {
		return fmt.Errorf("update_node error: %s", errMsg)
	}

	return fmt.Errorf("update_node failed with unknown error")
}

// DeleteNode implements TextCleanerCommands.DeleteNode
func (s *SocketClientCommands) DeleteNode(nodeID string) error {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "delete_node",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		return nil
	}

	if errMsg, ok := resp["error"].(string); ok {
		return fmt.Errorf("delete_node error: %s", errMsg)
	}

	return fmt.Errorf("delete_node failed with unknown error")
}

// AddChildNode implements TextCleanerCommands.AddChildNode
func (s *SocketClientCommands) AddChildNode(parentID, nodeType, name, operation, arg1, arg2, condition string) (string, error) {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "add_child_node",
		"params": map[string]interface{}{
			"parent_id":  parentID,
			"type":       nodeType,
			"name":       name,
			"operation":  operation,
			"arg1":       arg1,
			"arg2":       arg2,
			"condition":  condition,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return "", fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if nodeID, ok := result["node_id"].(string); ok {
				return nodeID, nil
			}
		}
	}

	if errMsg, ok := resp["error"].(string); ok {
		return "", fmt.Errorf("add_child_node error: %s", errMsg)
	}

	return "", fmt.Errorf("add_child_node failed with unknown error")
}

// SelectNode implements TextCleanerCommands.SelectNode
func (s *SocketClientCommands) SelectNode(nodeID string) error {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "select_node",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		return nil
	}

	if errMsg, ok := resp["error"].(string); ok {
		return fmt.Errorf("select_node error: %s", errMsg)
	}

	return fmt.Errorf("select_node failed with unknown error")
}

// ============================================================================
// Tree Operations Methods
// ============================================================================

// IndentNode implements TextCleanerCommands.IndentNode
func (s *SocketClientCommands) IndentNode(nodeID string) error {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "indent_node",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		return nil
	}

	if errMsg, ok := resp["error"].(string); ok {
		return fmt.Errorf("indent_node error: %s", errMsg)
	}

	return fmt.Errorf("indent_node failed with unknown error")
}

// UnindentNode implements TextCleanerCommands.UnindentNode
func (s *SocketClientCommands) UnindentNode(nodeID string) error {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "unindent_node",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		return nil
	}

	if errMsg, ok := resp["error"].(string); ok {
		return fmt.Errorf("unindent_node error: %s", errMsg)
	}

	return fmt.Errorf("unindent_node failed with unknown error")
}

// MoveNodeUp implements TextCleanerCommands.MoveNodeUp
func (s *SocketClientCommands) MoveNodeUp(nodeID string) error {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "move_node_up",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		return nil
	}

	if errMsg, ok := resp["error"].(string); ok {
		return fmt.Errorf("move_node_up error: %s", errMsg)
	}

	return fmt.Errorf("move_node_up failed with unknown error")
}

// MoveNodeDown implements TextCleanerCommands.MoveNodeDown
func (s *SocketClientCommands) MoveNodeDown(nodeID string) error {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "move_node_down",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		return nil
	}

	if errMsg, ok := resp["error"].(string); ok {
		return fmt.Errorf("move_node_down error: %s", errMsg)
	}

	return fmt.Errorf("move_node_down failed with unknown error")
}

// MoveNodeToPosition implements TextCleanerCommands.MoveNodeToPosition
func (s *SocketClientCommands) MoveNodeToPosition(nodeID, newParentID string, position int) error {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "move_node_to_position",
		"params": map[string]interface{}{
			"node_id":      nodeID,
			"new_parent_id": newParentID,
			"position":     position,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		return nil
	}

	if errMsg, ok := resp["error"].(string); ok {
		return fmt.Errorf("move_node_to_position error: %s", errMsg)
	}

	return fmt.Errorf("move_node_to_position failed with unknown error")
}

// CanIndentNode implements TextCleanerCommands.CanIndentNode
func (s *SocketClientCommands) CanIndentNode(nodeID string) bool {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "can_indent_node",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("CanIndentNode socket error: %v", err)
		return false
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if canIndent, ok := result["can_indent"].(bool); ok {
				return canIndent
			}
		}
	}

	return false
}

// CanUnindentNode implements TextCleanerCommands.CanUnindentNode
func (s *SocketClientCommands) CanUnindentNode(nodeID string) bool {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "can_unindent_node",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("CanUnindentNode socket error: %v", err)
		return false
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if canUnindent, ok := result["can_unindent"].(bool); ok {
				return canUnindent
			}
		}
	}

	return false
}

// CanMoveNodeUp implements TextCleanerCommands.CanMoveNodeUp
func (s *SocketClientCommands) CanMoveNodeUp(nodeID string) bool {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "can_move_node_up",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("CanMoveNodeUp socket error: %v", err)
		return false
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if canMove, ok := result["can_move_up"].(bool); ok {
				return canMove
			}
		}
	}

	return false
}

// CanMoveNodeDown implements TextCleanerCommands.CanMoveNodeDown
func (s *SocketClientCommands) CanMoveNodeDown(nodeID string) bool {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "can_move_node_down",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("CanMoveNodeDown socket error: %v", err)
		return false
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if canMove, ok := result["can_move_down"].(bool); ok {
				return canMove
			}
		}
	}

	return false
}

// ============================================================================
// Text Processing Methods
// ============================================================================

// SetInputText implements TextCleanerCommands.SetInputText
func (s *SocketClientCommands) SetInputText(text string) {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "set_input_text",
		"params": map[string]interface{}{
			"text": text,
		},
	})

	_, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("SetInputText socket error: %v", err)
	}
}

// GetInputText implements TextCleanerCommands.GetInputText
func (s *SocketClientCommands) GetInputText() string {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "get_input_text",
		"params": map[string]interface{}{},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("GetInputText socket error: %v", err)
		return ""
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if text, ok := result["text"].(string); ok {
				return text
			}
		}
	}

	return ""
}

// GetOutputText implements TextCleanerCommands.GetOutputText
func (s *SocketClientCommands) GetOutputText() string {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "get_output_text",
		"params": map[string]interface{}{},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("GetOutputText socket error: %v", err)
		return ""
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if output, ok := result["output"].(string); ok {
				return output
			}
		}
	}

	return ""
}

// GetOutputTextAtNode implements TextCleanerCommands.GetOutputTextAtNode
func (s *SocketClientCommands) GetOutputTextAtNode(nodeID string) string {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "get_output_text_at_node",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("GetOutputTextAtNode socket error: %v", err)
		return ""
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if output, ok := result["output"].(string); ok {
				return output
			}
		}
	}

	return ""
}

// ============================================================================
// Query Operations Methods
// ============================================================================

// GetNode implements TextCleanerCommands.GetNode
func (s *SocketClientCommands) GetNode(nodeID string) *PipelineNode {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "get_node",
		"params": map[string]interface{}{
			"node_id": nodeID,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("GetNode socket error: %v", err)
		return nil
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if nodeData, ok := result["node"]; ok {
				// Convert interface{} back to PipelineNode
				nodeJSON, _ := json.Marshal(nodeData)
				var node PipelineNode
				if err := json.Unmarshal(nodeJSON, &node); err == nil {
					return &node
				}
			}
		}
	}

	return nil
}

// GetSelectedNodeID implements TextCleanerCommands.GetSelectedNodeID
func (s *SocketClientCommands) GetSelectedNodeID() string {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "get_selected_node_id",
		"params": map[string]interface{}{},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("GetSelectedNodeID socket error: %v", err)
		return ""
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if nodeID, ok := result["node_id"].(string); ok {
				return nodeID
			}
		}
	}

	return ""
}

// GetPipeline implements TextCleanerCommands.GetPipeline
func (s *SocketClientCommands) GetPipeline() []PipelineNode {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "get_pipeline",
		"params": map[string]interface{}{},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		log.Printf("GetPipeline socket error: %v", err)
		return []PipelineNode{}
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if pipelineData, ok := result["pipeline"]; ok {
				// Convert interface{} back to []PipelineNode
				pipelineJSON, _ := json.Marshal(pipelineData)
				var pipeline []PipelineNode
				if err := json.Unmarshal(pipelineJSON, &pipeline); err == nil {
					return pipeline
				}
			}
		}
	}

	return []PipelineNode{}
}

// ============================================================================
// Import/Export Methods
// ============================================================================

// ExportPipeline implements TextCleanerCommands.ExportPipeline
func (s *SocketClientCommands) ExportPipeline() (string, error) {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "export_pipeline",
		"params": map[string]interface{}{},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return "", fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		if result, ok := resp["result"].(map[string]interface{}); ok {
			if pipeline, ok := result["pipeline"]; ok {
				pipelineJSON, _ := json.Marshal(pipeline)
				return string(pipelineJSON), nil
			}
		}
	}

	if errMsg, ok := resp["error"].(string); ok {
		return "", fmt.Errorf("export_pipeline error: %s", errMsg)
	}

	return "", fmt.Errorf("export_pipeline failed with unknown error")
}

// ImportPipeline implements TextCleanerCommands.ImportPipeline
func (s *SocketClientCommands) ImportPipeline(jsonStr string) error {
	cmdJSON, _ := json.Marshal(map[string]interface{}{
		"action": "import_pipeline",
		"params": map[string]interface{}{
			"json": jsonStr,
		},
	})

	resp, err := s.client.Execute(string(cmdJSON))
	if err != nil {
		return fmt.Errorf("socket error: %w", err)
	}

	if success, ok := resp["success"].(bool); ok && success {
		return nil
	}

	if errMsg, ok := resp["error"].(string); ok {
		return fmt.Errorf("import_pipeline error: %s", errMsg)
	}

	return fmt.Errorf("import_pipeline failed with unknown error")
}
