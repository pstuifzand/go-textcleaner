#!/bin/bash

# TextCleaner REPL Command Runner
# Usage: ./run-repl-command.sh <socket-path> <command>
# Example: ./run-repl-command.sh /tmp/textcleaner.sock "create node Uppercase operation Uppercase"

set -e

SOCKET_PATH="${1:---socket}"
COMMAND="${2:-}"

if [[ -z "$SOCKET_PATH" ]] || [[ "$SOCKET_PATH" == "--socket" ]]; then
    echo "Usage: $0 <socket-path> <command>"
    echo "Example: $0 /tmp/textcleaner.sock 'create node Uppercase operation Uppercase'"
    echo ""
    echo "Default socket path: /tmp/textcleaner.sock"
    exit 1
fi

# If only one argument, treat it as the command and use default socket
if [[ -z "$COMMAND" ]]; then
    COMMAND="$SOCKET_PATH"
    SOCKET_PATH="/tmp/textcleaner.sock"
fi

# Check if socket server is running
if [[ ! -S "$SOCKET_PATH" ]]; then
    echo "Error: Socket server not found at $SOCKET_PATH"
    echo "Start the server first with:"
    echo "  ./go-textcleaner --headless --socket $SOCKET_PATH"
    exit 1
fi

# Start REPL in non-interactive mode with the command
echo "$COMMAND" | ./go-textcleaner --repl --socket "$SOCKET_PATH" 2>&1 | tail -n +4
