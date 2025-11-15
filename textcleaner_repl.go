package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

// REPLCommand represents a parsed command
type REPLCommand struct {
	Verb   string
	Object string
	Args   []string
}

// REPLFormatter handles output formatting
type REPLFormatter struct {
	useColor bool
}

// NewREPLFormatter creates a new formatter
func NewREPLFormatter(useColor bool) *REPLFormatter {
	return &REPLFormatter{useColor: useColor}
}

// PrintSuccess prints a success message
func (f *REPLFormatter) PrintSuccess(message string) {
	if f.useColor {
		color.Green("✓ %s\n", message)
	} else {
		fmt.Printf("✓ %s\n", message)
	}
}

// PrintError prints an error message
func (f *REPLFormatter) PrintError(message string) {
	if f.useColor {
		color.Red("✗ Error: %s\n", message)
	} else {
		fmt.Printf("✗ Error: %s\n", message)
	}
}

// PrintInfo prints an info message
func (f *REPLFormatter) PrintInfo(message string) {
	if f.useColor {
		color.Cyan("ℹ %s\n", message)
	} else {
		fmt.Printf("ℹ %s\n", message)
	}
}

// PrintTable prints a formatted ASCII table
func (f *REPLFormatter) PrintTable(headers []string, rows [][]string) {
	// Simple table printing without tablewriter
	if len(headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, h := range headers {
		fmt.Printf("%-*s", widths[i], h)
		if i < len(headers)-1 {
			fmt.Print("  ")
		}
	}
	fmt.Println()

	// Print separator
	for i := range headers {
		for j := 0; j < widths[i]; j++ {
			fmt.Print("-")
		}
		if i < len(headers)-1 {
			fmt.Print("  ")
		}
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			fmt.Printf("%-*s", widths[i], cell)
			if i < len(headers)-1 {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
}

// PrintJSON prints formatted JSON
func (f *REPLFormatter) PrintJSON(data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		f.PrintError("Failed to format JSON: " + err.Error())
		return
	}
	fmt.Println(string(jsonBytes))
}

// PrintPipeline prints the pipeline as a tree
func (f *REPLFormatter) PrintPipeline(pipeline []interface{}) {
	f.printTreeNodes(pipeline, 0)
}

// printTreeNodes recursively prints nodes as a tree
func (f *REPLFormatter) printTreeNodes(nodes []interface{}, indent int) {
	for i, nodeInterface := range nodes {
		if node, ok := nodeInterface.(map[string]interface{}); ok {
			prefix := strings.Repeat("  ", indent)
			isLast := i == len(nodes)-1

			if indent == 0 {
				if isLast {
					fmt.Print("└─ ")
				} else {
					fmt.Print("├─ ")
				}
			} else {
				if isLast {
					fmt.Print(prefix + "└─ ")
				} else {
					fmt.Print(prefix + "├─ ")
				}
			}

			// Print node info
			name := "Unknown"
			if n, ok := node["name"].(string); ok {
				name = n
			}
			nodeID := ""
			if id, ok := node["id"].(string); ok {
				nodeID = id
			}

			operation := ""
			if op, ok := node["operation"].(string); ok && op != "" {
				operation = fmt.Sprintf(" [%s]", op)
			}

			if f.useColor {
				fmt.Printf(color.CyanString(name) + color.YellowString(operation) + " " + color.BlackString("(%s)\n", nodeID))
			} else {
				fmt.Printf("%s%s (%s)\n", name, operation, nodeID)
			}

			// Print children
			if children, ok := node["children"].([]interface{}); ok && len(children) > 0 {
				f.printTreeNodes(children, indent+1)
			}
		}
	}
}

// ParseCommand parses a verb-first command string
func ParseCommand(input string) (*REPLCommand, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty command")
	}

	// Split by whitespace, but handle quoted strings
	parts := splitArgs(input)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmd := &REPLCommand{
		Verb: strings.ToLower(parts[0]),
	}

	if len(parts) > 1 {
		cmd.Object = strings.ToLower(parts[1])
		cmd.Args = parts[2:]
	}

	return cmd, nil
}

// splitArgs splits a command string into arguments, respecting quotes
func splitArgs(input string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	escaped := false

	for _, ch := range input {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if (ch == '"' || ch == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = ch
			continue
		}

		if ch == quoteChar && inQuotes {
			inQuotes = false
			quoteChar = 0
			continue
		}

		if ch == ' ' && !inQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(ch)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// ExecuteREPLCommand executes a REPL command and returns the result
func ExecuteREPLCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter, rl *readline.Instance) error {
	switch cmd.Verb {
	// Node management commands
	case "create":
		return handleCreateCommand(cmd, client, formatter)
	case "update":
		return handleUpdateCommand(cmd, client, formatter)
	case "delete":
		return handleDeleteCommand(cmd, client, formatter)
	case "select":
		return handleSelectCommand(cmd, client, formatter)

	// Tree operations
	case "indent":
		return handleIndentCommand(cmd, client, formatter)
	case "unindent":
		return handleUnindentCommand(cmd, client, formatter)
	case "move":
		return handleMoveCommand(cmd, client, formatter)

	// Query commands
	case "show":
		return handleShowCommand(cmd, client, formatter)
	case "list":
		return handleListCommand(cmd, client, formatter)
	case "get":
		return handleGetCommand(cmd, client, formatter)

	// Meta commands
	case "info":
		return handleInfoCommand(cmd, client, formatter)

	// Text processing
	case "set":
		return handleSetCommand(cmd, client, formatter, rl)

	// Pipeline commands
	case "export":
		return handleExportCommand(cmd, client, formatter)
	case "import":
		return handleImportCommand(cmd, client, formatter, rl)

	// Utility commands
	case "help":
		return handleHelpCommand(cmd)
	case "quit", "exit":
		return fmt.Errorf("exit")
	case "clear":
		fmt.Print("\033[2J\033[H") // Clear screen
		return nil

	default:
		formatter.PrintError(fmt.Sprintf("Unknown command: %s", cmd.Verb))
		formatter.PrintInfo("Type 'help' for available commands")
		return nil
	}
}

// Command handlers

func handleCreateCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	if cmd.Object == "node" {
		// create node <name> [type <node_type>] [operation <op_name>] [arg1 <value>] [arg2 <value>] [parent <parent_id>]
		// Support both: "create node Name OpName" and "create node Name operation OpName"
		// Node types: operation (default), foreach, if, group
		if len(cmd.Args) < 1 {
			formatter.PrintError("create node requires a name")
			return nil
		}
		name := cmd.Args[0]
		nodeType := "operation" // Default to operation type
		operation := ""
		arg1 := ""
		arg2 := ""
		condition := ""
		parentID := ""

		// Parse remaining arguments
		i := 1
		for i < len(cmd.Args) {
			arg := strings.ToLower(cmd.Args[i])

			// Check if this is a keyword
			if arg == "type" {
				if i+1 < len(cmd.Args) {
					nodeType = cmd.Args[i+1]
					i += 2
					continue
				}
			} else if arg == "operation" {
				if i+1 < len(cmd.Args) {
					operation = cmd.Args[i+1]
					i += 2
					continue
				}
			} else if arg == "arg1" {
				if i+1 < len(cmd.Args) {
					arg1 = cmd.Args[i+1]
					i += 2
					continue
				}
			} else if arg == "arg2" {
				if i+1 < len(cmd.Args) {
					arg2 = cmd.Args[i+1]
					i += 2
					continue
				}
			} else if arg == "condition" {
				if i+1 < len(cmd.Args) {
					condition = cmd.Args[i+1]
					i += 2
					continue
				}
			} else if arg == "parent" {
				if i+1 < len(cmd.Args) {
					parentID = cmd.Args[i+1]
					i += 2
					continue
				}
			} else if !isKeyword(arg) {
				// If not a keyword, treat as positional: first extra arg is operation
				if operation == "" {
					operation = cmd.Args[i]
				} else if arg1 == "" {
					arg1 = cmd.Args[i]
				} else if arg2 == "" {
					arg2 = cmd.Args[i]
				} else if condition == "" {
					condition = cmd.Args[i]
				}
				i++
				continue
			}
			i++
		}

		// Build JSON command with parent_id and node type
		var jsonCmd string
		if parentID != "" {
			jsonCmd = fmt.Sprintf(
				`{"action":"create_node","params":{"type":"%s","name":"%s","operation":"%s","arg1":"%s","arg2":"%s","condition":"%s","parent_id":"%s"}}`,
				escapeJSON(nodeType), escapeJSON(name), escapeJSON(operation), escapeJSON(arg1), escapeJSON(arg2), escapeJSON(condition), escapeJSON(parentID),
			)
		} else {
			jsonCmd = fmt.Sprintf(
				`{"action":"create_node","params":{"type":"%s","name":"%s","operation":"%s","arg1":"%s","arg2":"%s","condition":"%s"}}`,
				escapeJSON(nodeType), escapeJSON(name), escapeJSON(operation), escapeJSON(arg1), escapeJSON(arg2), escapeJSON(condition),
			)
		}

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			if result, ok := response["result"].(map[string]interface{}); ok {
				if nodeID, ok := result["node_id"].(string); ok {
					formatter.PrintSuccess(fmt.Sprintf("Created node: %s", nodeID))
					return nil
				}
			}
		}

		if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}
	} else if cmd.Object == "child" {
		// create child <parent_id> <name> [operation] [arg1] [arg2]
		if len(cmd.Args) < 2 {
			formatter.PrintError("create child requires parent_id and name")
			return nil
		}
		parentID := cmd.Args[0]
		name := cmd.Args[1]
		operation := ""
		arg1 := ""
		arg2 := ""

		if len(cmd.Args) > 2 {
			operation = cmd.Args[2]
		}
		if len(cmd.Args) > 3 {
			arg1 = cmd.Args[3]
		}
		if len(cmd.Args) > 4 {
			arg2 = cmd.Args[4]
		}

		jsonCmd := fmt.Sprintf(
			`{"action":"add_child_node","params":{"parent_id":"%s","type":"operation","name":"%s","operation":"%s","arg1":"%s","arg2":"%s","condition":""}}`,
			escapeJSON(parentID), escapeJSON(name), escapeJSON(operation), escapeJSON(arg1), escapeJSON(arg2),
		)

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			if result, ok := response["result"].(map[string]interface{}); ok {
				if nodeID, ok := result["node_id"].(string); ok {
					formatter.PrintSuccess(fmt.Sprintf("Created child node: %s", nodeID))
					return nil
				}
			}
		}

		if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}
	} else {
		formatter.PrintError("create requires 'node' or 'child' argument")
	}
	return nil
}

func handleUpdateCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	if cmd.Object == "node" {
		// update node <node_id> <name> [operation <op_name>] [arg1 <value>] [arg2 <value>]
		if len(cmd.Args) < 2 {
			formatter.PrintError("update node requires node_id and name")
			return nil
		}
		nodeID := cmd.Args[0]
		name := cmd.Args[1]
		operation := ""
		arg1 := ""
		arg2 := ""

		// Parse remaining arguments with keyword support
		i := 2
		for i < len(cmd.Args) {
			arg := strings.ToLower(cmd.Args[i])

			if arg == "operation" {
				if i+1 < len(cmd.Args) {
					operation = cmd.Args[i+1]
					i += 2
					continue
				}
			} else if arg == "arg1" {
				if i+1 < len(cmd.Args) {
					arg1 = cmd.Args[i+1]
					i += 2
					continue
				}
			} else if arg == "arg2" {
				if i+1 < len(cmd.Args) {
					arg2 = cmd.Args[i+1]
					i += 2
					continue
				}
			} else if !isKeyword(arg) {
				// Positional arguments
				if operation == "" {
					operation = cmd.Args[i]
				} else if arg1 == "" {
					arg1 = cmd.Args[i]
				} else if arg2 == "" {
					arg2 = cmd.Args[i]
				}
				i++
				continue
			}
			i++
		}

		jsonCmd := fmt.Sprintf(
			`{"action":"update_node","params":{"node_id":"%s","type":"operation","name":"%s","operation":"%s","arg1":"%s","arg2":"%s","condition":""}}`,
			escapeJSON(nodeID), escapeJSON(name), escapeJSON(operation), escapeJSON(arg1), escapeJSON(arg2),
		)

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			formatter.PrintSuccess(fmt.Sprintf("Updated node: %s", nodeID))
			return nil
		}

		if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}
	} else {
		formatter.PrintError("update requires 'node' argument")
	}
	return nil
}

func handleDeleteCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	if cmd.Object == "node" {
		// delete node <node_id>
		if len(cmd.Args) < 1 {
			formatter.PrintError("delete node requires node_id")
			return nil
		}
		nodeID := cmd.Args[0]

		jsonCmd := fmt.Sprintf(
			`{"action":"delete_node","params":{"node_id":"%s"}}`,
			escapeJSON(nodeID),
		)

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			formatter.PrintSuccess(fmt.Sprintf("Deleted node: %s", nodeID))
			return nil
		}

		if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}
	} else {
		formatter.PrintError("delete requires 'node' argument")
	}
	return nil
}

func handleSelectCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	if cmd.Object == "node" {
		// select node <node_id>
		if len(cmd.Args) < 1 {
			formatter.PrintError("select node requires node_id")
			return nil
		}
		nodeID := cmd.Args[0]

		jsonCmd := fmt.Sprintf(
			`{"action":"select_node","params":{"node_id":"%s"}}`,
			escapeJSON(nodeID),
		)

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			formatter.PrintSuccess(fmt.Sprintf("Selected node: %s", nodeID))
			return nil
		}

		if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}
	} else {
		formatter.PrintError("select requires 'node' argument")
	}
	return nil
}

func handleAddChildCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	// This is for future "add child" command implementation
	formatter.PrintError("add child command not yet implemented via REPL, use create child instead")
	return nil
}

func handleIndentCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	// indent <node_id>
	if len(cmd.Args) < 1 {
		formatter.PrintError("indent requires node_id")
		return nil
	}
	nodeID := cmd.Args[0]

	jsonCmd := fmt.Sprintf(
		`{"action":"indent_node","params":{"node_id":"%s"}}`,
		escapeJSON(nodeID),
	)

	response, err := client.Execute(jsonCmd)
	if err != nil {
		formatter.PrintError(err.Error())
		return nil
	}

	if success, ok := response["success"].(bool); ok && success {
		formatter.PrintSuccess(fmt.Sprintf("Indented node: %s", nodeID))
		return nil
	}

	if errMsg, ok := response["error"].(string); ok {
		formatter.PrintError(errMsg)
	}
	return nil
}

func handleUnindentCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	// unindent <node_id>
	if len(cmd.Args) < 1 {
		formatter.PrintError("unindent requires node_id")
		return nil
	}
	nodeID := cmd.Args[0]

	jsonCmd := fmt.Sprintf(
		`{"action":"unindent_node","params":{"node_id":"%s"}}`,
		escapeJSON(nodeID),
	)

	response, err := client.Execute(jsonCmd)
	if err != nil {
		formatter.PrintError(err.Error())
		return nil
	}

	if success, ok := response["success"].(bool); ok && success {
		formatter.PrintSuccess(fmt.Sprintf("Unindented node: %s", nodeID))
		return nil
	}

	if errMsg, ok := response["error"].(string); ok {
		formatter.PrintError(errMsg)
	}
	return nil
}

func handleMoveCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	// move up/down <node_id>
	if len(cmd.Args) < 1 {
		formatter.PrintError("move requires direction (up/down) and node_id")
		return nil
	}

	direction := strings.ToLower(cmd.Object)
	nodeID := cmd.Args[0]

	var action string
	if direction == "up" {
		action = "move_node_up"
	} else if direction == "down" {
		action = "move_node_down"
	} else {
		formatter.PrintError("move requires 'up' or 'down' direction")
		return nil
	}

	jsonCmd := fmt.Sprintf(
		`{"action":"%s","params":{"node_id":"%s"}}`,
		action, escapeJSON(nodeID),
	)

	response, err := client.Execute(jsonCmd)
	if err != nil {
		formatter.PrintError(err.Error())
		return nil
	}

	if success, ok := response["success"].(bool); ok && success {
		formatter.PrintSuccess(fmt.Sprintf("Moved node %s: %s", direction, nodeID))
		return nil
	}

	if errMsg, ok := response["error"].(string); ok {
		formatter.PrintError(errMsg)
	}
	return nil
}

func handleShowCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	switch cmd.Object {
	case "node":
		// show node <node_id>
		if len(cmd.Args) < 1 {
			formatter.PrintError("show node requires node_id")
			return nil
		}
		nodeID := cmd.Args[0]

		jsonCmd := fmt.Sprintf(
			`{"action":"get_node","params":{"node_id":"%s"}}`,
			escapeJSON(nodeID),
		)

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			if result, ok := response["result"].(map[string]interface{}); ok {
				formatter.PrintJSON(result)
			}
		} else if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}

	case "pipeline", "tree":
		// show pipeline/tree
		jsonCmd := `{"action":"export_pipeline","params":{}}`

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			if result, ok := response["result"].(map[string]interface{}); ok {
				if pipeline, ok := result["pipeline"].([]interface{}); ok {
					if cmd.Object == "tree" {
						formatter.PrintPipeline(pipeline)
					} else {
						formatter.PrintJSON(pipeline)
					}
				}
			}
		} else if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}

	default:
		formatter.PrintError("show requires 'node' or 'pipeline' or 'tree' argument")
	}
	return nil
}

func handleListCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	if cmd.Object == "nodes" {
		// list nodes
		jsonCmd := `{"action":"list_nodes","params":{}}`

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			if result, ok := response["result"].(map[string]interface{}); ok {
				if nodes, ok := result["nodes"].([]interface{}); ok {
					// Format as table
					headers := []string{"ID", "Name", "Type", "Operation"}
					var rows [][]string

					for _, nodeInterface := range nodes {
						if node, ok := nodeInterface.(map[string]interface{}); ok {
							id := shortenString(fmt.Sprintf("%v", node["id"]), 20)
							name := shortenString(fmt.Sprintf("%v", node["name"]), 30)
							nodeType := shortenString(fmt.Sprintf("%v", node["type"]), 15)
							operation := shortenString(fmt.Sprintf("%v", node["operation"]), 30)

							rows = append(rows, []string{id, name, nodeType, operation})
						}
					}

					formatter.PrintTable(headers, rows)
				}
			}
		} else if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}
	} else {
		formatter.PrintError("list requires 'nodes' argument")
	}
	return nil
}

func handleGetCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	switch cmd.Object {
	case "input":
		// get input
		jsonCmd := `{"action":"get_input_text","params":{}}`

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			if result, ok := response["result"].(map[string]interface{}); ok {
				if text, ok := result["text"].(string); ok {
					fmt.Println(text)
				}
			}
		} else if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}

	case "output":
		// get output
		jsonCmd := `{"action":"get_output_text","params":{}}`

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			if result, ok := response["result"].(map[string]interface{}); ok {
				if output, ok := result["output"].(string); ok {
					fmt.Println(output)
				}
			}
		} else if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}

	case "selected":
		// get selected
		jsonCmd := `{"action":"get_selected_node_id","params":{}}`

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			if result, ok := response["result"].(map[string]interface{}); ok {
				if nodeID, ok := result["node_id"].(string); ok {
					if nodeID == "" {
						formatter.PrintInfo("No node selected")
					} else {
						fmt.Println(nodeID)
					}
				}
			}
		} else if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}

	default:
		formatter.PrintError("get requires 'input', 'output', or 'selected' argument")
	}
	return nil
}

func handleSetCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter, rl *readline.Instance) error {
	if cmd.Object == "input" {
		// set input <text or empty for multiline>
		var text string
		if len(cmd.Args) > 0 {
			text = strings.Join(cmd.Args, " ")
		} else {
			// Prompt for multiline input
			formatter.PrintInfo("Enter text (end with blank line):")

			// Use readline to read multiple lines
			var lines []string
			rl.SetPrompt("")
			for {
				line, err := rl.Readline()
				if err == readline.ErrInterrupt {
					continue
				} else if err != nil {
					if err.Error() == "EOF" {
						break
					}
					break
				}

				// Empty line ends input
				if strings.TrimSpace(line) == "" {
					break
				}
				lines = append(lines, line)
			}
			rl.SetPrompt("textcleaner> ")
			text = strings.Join(lines, "\n")
		}

		jsonCmd := fmt.Sprintf(
			`{"action":"set_input_text","params":{"text":"%s"}}`,
			escapeJSON(text),
		)

		response, err := client.Execute(jsonCmd)
		if err != nil {
			formatter.PrintError(err.Error())
			return nil
		}

		if success, ok := response["success"].(bool); ok && success {
			formatter.PrintSuccess("Input text set")
			return nil
		}

		if errMsg, ok := response["error"].(string); ok {
			formatter.PrintError(errMsg)
		}
	} else {
		formatter.PrintError("set requires 'input' argument")
	}
	return nil
}

func handleExportCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	// export
	jsonCmd := `{"action":"export_pipeline","params":{}}`

	response, err := client.Execute(jsonCmd)
	if err != nil {
		formatter.PrintError(err.Error())
		return nil
	}

	if success, ok := response["success"].(bool); ok && success {
		if result, ok := response["result"].(map[string]interface{}); ok {
			formatter.PrintJSON(result["pipeline"])
		}
	} else if errMsg, ok := response["error"].(string); ok {
		formatter.PrintError(errMsg)
	}
	return nil
}

func handleImportCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter, rl *readline.Instance) error {
	// import <json or empty for multiline>
	var jsonStr string
	if len(cmd.Args) > 0 {
		jsonStr = strings.Join(cmd.Args, " ")
	} else {
		formatter.PrintInfo("Enter JSON pipeline (end with blank line):")

		// Use readline to read multiple lines
		var lines []string
		rl.SetPrompt("")
		for {
			line, err := rl.Readline()
			if err == readline.ErrInterrupt {
				continue
			} else if err != nil {
				if err.Error() == "EOF" {
					break
				}
				break
			}

			// Empty line ends input
			if strings.TrimSpace(line) == "" {
				break
			}
			lines = append(lines, line)
		}
		rl.SetPrompt("textcleaner> ")
		jsonStr = strings.Join(lines, "\n")
	}

	if jsonStr == "" {
		formatter.PrintError("import requires JSON data")
		return nil
	}

	// The JSON is already in the correct format, don't escape it
	jsonCmd := fmt.Sprintf(
		`{"action":"import_pipeline","params":{"json":%s}}`,
		jsonStr,
	)

	response, err := client.Execute(jsonCmd)
	if err != nil {
		formatter.PrintError(err.Error())
		return nil
	}

	if success, ok := response["success"].(bool); ok && success {
		formatter.PrintSuccess("Pipeline imported")
		return nil
	}

	if errMsg, ok := response["error"].(string); ok {
		formatter.PrintError(errMsg)
	}
	return nil
}

func handleHelpCommand(cmd *REPLCommand) error {
	if len(cmd.Args) > 0 {
		// Help for specific command
		showSpecificHelp(cmd.Args[0])
	} else {
		showMainHelp()
	}
	return nil
}

func handleInfoCommand(cmd *REPLCommand, client *SocketClient, formatter *REPLFormatter) error {
	// info types - list available node types and operations
	// info versions - show version info
	if len(cmd.Args) == 0 || cmd.Args[0] == "types" {
		return showAvailableTypes(client, formatter)
	}

	formatter.PrintError("Unknown info subcommand: " + strings.Join(cmd.Args, " "))
	formatter.PrintInfo("Use 'info types' to list available node types and operations")
	return nil
}

func showAvailableTypes(client *SocketClient, formatter *REPLFormatter) error {
	jsonCmd := `{"action":"list_node_types","params":{}}`

	response, err := client.Execute(jsonCmd)
	if err != nil {
		formatter.PrintError(err.Error())
		return nil
	}

	if success, ok := response["success"].(bool); ok && success {
		if result, ok := response["result"].(map[string]interface{}); ok {
			// Display node types
			if nodeTypes, ok := result["node_types"].([]interface{}); ok {
				formatter.PrintInfo("Available Node Types:")
				for _, typeInterface := range nodeTypes {
					if nodeType, ok := typeInterface.(string); ok {
						fmt.Printf("  • %s\n", nodeType)
					}
				}
				fmt.Println()
			}

			// Display operations
			if operations, ok := result["operations"].([]interface{}); ok {
				formatter.PrintInfo("Available Operations:")
				headers := []string{"Operation Name"}
				var rows [][]string

				for _, opInterface := range operations {
					if opName, ok := opInterface.(string); ok {
						rows = append(rows, []string{opName})
					}
				}

				formatter.PrintTable(headers, rows)
			}
			return nil
		}
	}

	if errMsg, ok := response["error"].(string); ok {
		formatter.PrintError(errMsg)
	}
	return nil
}

func showMainHelp() {
	help := `
TextCleaner REPL - Available Commands
=====================================

NODE MANAGEMENT:
  create node <name> [operation] [arg1] [arg2] [condition]
                              Create a new root-level node
  create child <parent_id> <name> [operation] [arg1] [arg2]
                              Create a child node
  update node <node_id> <name> [operation] [arg1] [arg2]
                              Update an existing node
  delete node <node_id>        Delete a node and its children
  select node <node_id>        Select a node

TREE OPERATIONS:
  indent <node_id>            Make node a child of previous sibling
  unindent <node_id>          Make node a sibling of its parent
  move up <node_id>           Move node earlier in sibling list
  move down <node_id>         Move node later in sibling list

QUERY COMMANDS:
  show node <node_id>         Show details of a specific node
  show pipeline               Show pipeline as JSON
  show tree                   Show pipeline as tree view
  list nodes                  List all root nodes as table
  get input                   Get current input text
  get output                  Get processed output text
  get selected                Get currently selected node ID

TEXT PROCESSING:
  set input <text>            Set input text
  set input                   Enter multiline input mode

PIPELINE MANAGEMENT:
  export                      Export pipeline as JSON
  import <json>               Import pipeline from JSON
  import                      Enter multiline JSON import mode

UTILITIES:
  help [command]              Show this help or help for specific command
  info [types]                Show available node types and operations
  clear                       Clear the screen
  quit, exit                  Exit the REPL

EXAMPLES:
  > info types                # Show all available node types and operations
  > create node Uppercase operation Uppercase
  > set input hello world
  > get output
  > show tree
  > list nodes
  > move down node_0
  > help create

Type 'help' followed by a command name for detailed help.
`
	fmt.Print(help)
}

func showSpecificHelp(command string) {
	helps := map[string]string{
		"create": `
create node <name> [operation] [arg1] [arg2] [condition]
  Creates a new root-level node.

  Examples:
    create node Uppercase operation Uppercase
    create node Replace operation Replace foo bar

create child <parent_id> <name> [operation] [arg1] [arg2]
  Creates a child node under a parent node.

  Example:
    create child node_0 Lowercase operation Lowercase
`,
		"update": `
update node <node_id> <name> [operation] [arg1] [arg2]
  Updates an existing node's properties.

  Example:
    update node node_0 NewName operation Uppercase
`,
		"delete": `
delete node <node_id>
  Deletes a node and all its children.

  Example:
    delete node node_0
`,
		"set": `
set input <text>
  Sets the input text to be processed by the pipeline.

  Examples:
    set input hello world
    set input
      (then enter multiline text)
`,
		"show": `
show node <node_id>     Show details of a specific node
show pipeline           Show the pipeline as JSON
show tree               Show the pipeline as a tree diagram
`,
		"move": `
move up <node_id>       Move a node earlier in its sibling list
move down <node_id>     Move a node later in its sibling list
`,
		"indent": `
indent <node_id>        Make a node a child of the previous sibling
unindent <node_id>      Make a node a sibling of its parent
`,
	}

	if help, ok := helps[command]; ok {
		fmt.Println(help)
	} else {
		fmt.Printf("No help available for '%s'\n", command)
		fmt.Println("Type 'help' for a list of all commands")
	}
}

// REPLSession manages the REPL interactive session
type REPLSession struct {
	client    *SocketClient
	formatter *REPLFormatter
	history   []string
}

// NewREPLSession creates a new REPL session
func NewREPLSession(socketPath string) (*REPLSession, error) {
	client, err := NewSocketClient(socketPath)
	if err != nil {
		return nil, err
	}

	session := &REPLSession{
		client:    client,
		formatter: NewREPLFormatter(true),
		history:   make([]string, 0),
	}

	return session, nil
}

// Run starts the interactive REPL loop
func (rs *REPLSession) Run() error {
	// Create readline instance
	rl, err := readline.New("textcleaner> ")
	if err != nil {
		return err
	}
	defer rl.Close()

	// Print banner
	color.Cyan("TextCleaner REPL v1.0\n")
	color.Cyan("Connected to socket server at %s\n", rs.client.conn.LocalAddr())
	color.Cyan("Type 'help' for available commands\n\n")

	// Main REPL loop
	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err != nil {
			// readline returns io.EOF as a simple string, not the io.EOF constant
			if err.Error() == "EOF" {
				fmt.Println()
				break
			}
			rs.formatter.PrintError(err.Error())
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Store in history
		rs.history = append(rs.history, line)

		// Parse and execute command
		cmd, err := ParseCommand(line)
		if err != nil {
			rs.formatter.PrintError(err.Error())
			continue
		}

		if err := ExecuteREPLCommand(cmd, rs.client, rs.formatter, rl); err != nil {
			if err.Error() == "exit" {
				break
			}
			rs.formatter.PrintError(err.Error())
		}
	}

	rs.formatter.PrintInfo("Goodbye!")
	rs.client.Close()
	return nil
}

// Helper functions

func isKeyword(arg string) bool {
	switch strings.ToLower(arg) {
	case "type", "operation", "arg1", "arg2", "condition", "parent":
		return true
	}
	return false
}

func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1]) // Remove quotes added by Marshal
}

func shortenString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
