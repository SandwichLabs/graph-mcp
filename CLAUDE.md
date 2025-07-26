# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the Agent Memory Graph MCP - an MCP (Model Context Protocol) server that exposes memory management and knowledge retrieval functions for LLMs. The project is in extremely early development and should not be used in production.

- Architecture Doc: [Architecture Doc](./docs/ARCHITECTURE.md)

## Architecture

The project follows a clean architecture pattern:

- **cmd/**: CLI commands using Cobra framework
- **internal/server/**: MCP server implementation using mark3labs/mcp-go
- **internal/tui/**: Terminal UI using Charmbracelet Bubble Tea (incomplete)
- **main.go**: Entry point that delegates to cmd package

The main binary is a CLI tool called `amg` that starts an MCP server with a specified memory graph directory.

## Development Commands

### Build and Run
```bash
go build -o amg .
./amg /path/to/memory/directory --name knowledge
```

### Testing
```bash
go test ./...                    # Run all tests
go test -v ./internal/...       # Run tests with verbose output
```

### Development Tools
```bash
go mod tidy                      # Clean up dependencies
go fmt ./...                     # Format code
go vet ./...                     # Static analysis
```

## Key Dependencies

- **github.com/mark3labs/mcp-go**: MCP protocol implementation
- **github.com/spf13/cobra**: CLI framework
- **github.com/charmbracelet/bubbletea**: TUI framework
- **github.com/tmc/langchaingo**: LangChain Go integration

## MCP Server Details

The server exposes memory management tools via MCP protocol. Key files:
- `internal/server/server.go:14`: Main server entry point
- `cmd/root.go:22`: Server startup in CLI command

The server includes extensive hooks for debugging and monitoring MCP interactions.

## Code Structure Notes

- Server has incomplete tool definitions (line 47 in server.go shows missing tools implementation)
- TUI implementation is skeletal with empty Update/View methods
- Project appears to be transitioning from a more complex design (see deleted files in git status)

## Current Development Status

Initial scaffolding in progress. The project is not yet functional and requires significant work on the server logic, tool definitions, and TUI implementation.