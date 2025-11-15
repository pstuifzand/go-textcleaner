# TextCleaner REPL Skill

A Claude Code skill for interacting with the TextCleaner REPL interface via the socket server.

## Files in This Skill

- **SKILL.md** - The main skill definition with instructions and command reference
- **run-repl-command.sh** - Helper script for running REPL commands programmatically
- **README.md** - This file

## Quick Start

1. **Build the project:**
   ```bash
   go build -o go-textcleaner
   ```

2. **Start the socket server** (in a separate terminal):
   ```bash
   ./go-textcleaner --headless --socket /tmp/textcleaner.sock
   ```

3. **Start the REPL** (in another terminal):
   ```bash
   ./go-textcleaner --repl --socket /tmp/textcleaner.sock
   ```

4. **Use Claude** to help with REPL commands:
   - "Use the textcleaner REPL to create an Uppercase node"
   - "Set the input text to 'hello world' and show me the output"
   - "Create a pipeline with Uppercase → Lowercase nodes"

## How Claude Uses This Skill

Claude automatically discovers and uses this skill when:
- You ask for help managing TextCleaner pipelines via the REPL
- You need to test or configure pipeline operations
- You want to interact with the socket server using natural language commands

Claude will:
1. Recognize the REPL interface and available commands
2. Translate your requests into appropriate REPL commands
3. Execute commands and interpret the responses
4. Help you understand and troubleshoot any issues

## Example Interactions

### Create a simple pipeline
```
User: "Use the REPL to create an Uppercase node, then set input to 'hello' and get the output"

Claude uses the skill to:
1. create node Uppercase operation Uppercase
2. set input hello
3. get output
4. Reports: "Output is 'HELLO'"
```

### Build a complex pipeline
```
User: "Create a pipeline that converts text to uppercase, then replaces 'L' with '1' "

Claude:
1. create node Uppercase operation Uppercase
2. create child node_0 Replace operation Replace L 1
3. set input hello
4. get output
5. Reports the result
```

### Debug a pipeline
```
User: "Show me the current pipeline structure and explain what it does"

Claude:
1. show tree
2. get output (to verify it works)
3. Explains the pipeline configuration
```

## Helper Script

The `run-repl-command.sh` script makes it easier to run single REPL commands:

```bash
# Run a single command
./run-repl-command.sh /tmp/textcleaner.sock "create node Uppercase operation Uppercase"

# Using default socket path
./run-repl-command.sh "list nodes"
```

## Requirements

- TextCleaner built and available
- Socket server running with `--headless --socket` flags
- REPL feature enabled in the build

## Troubleshooting

### Socket server not running
```
Error: Socket server not found at /tmp/textcleaner.sock
```
Solution: Start the server first:
```bash
./go-textcleaner --headless --socket /tmp/textcleaner.sock
```

### REPL command not recognized
Check the SKILL.md file for the complete list of available commands and syntax.

### Changes not appearing
Make sure you're:
1. Getting the output with `get output`
2. Waiting for the REPL to respond before checking changes
3. Using correct node IDs (use `list nodes` to verify)

## Development Notes

This skill is configured to use only `Bash` and `Read` tools, which are safe for executing REPL commands and reading documentation.

To extend this skill:
1. Update SKILL.md with new command examples
2. Enhance run-repl-command.sh with additional features if needed
3. Commit changes to git for team access

## See Also

- Main project CLAUDE.md for full socket server documentation
- TextCleaner REPL command reference in SKILL.md
- TextCleaner documentation for operation types and parameters
