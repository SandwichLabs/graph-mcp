package server

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Run(memoryPath string, serverName string) {
	// Initialize the MCP server with the provided memory path and server name
	// Create a new MCP server instance
	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		fmt.Fprintf(os.Stderr, "beforeAny: %s, %v, %v\n", method, id, message)
	})
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		fmt.Fprintf(os.Stderr, "onSuccess: %s, %v, %v, %v\n", method, id, message, result)
	})
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		fmt.Fprintf(os.Stderr, "onError: %s, %v, %v, %v\n", method, id, message, err)
	})
	hooks.AddBeforeInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest) {
		fmt.Fprintf(os.Stderr, "beforeInitialize: %v, %v\n", id, message)
	})
	hooks.AddOnRequestInitialization(func(ctx context.Context, id any, message any) error {
		fmt.Fprintf(os.Stderr, "AddOnRequestInitialization: %v, %v\n", id, message)
		// authorization verification and other preprocessing tasks are performed.
		return nil
	})
	hooks.AddAfterInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest, result *mcp.InitializeResult) {
		fmt.Fprintf(os.Stderr, "afterInitialize: %v, %v, %v\n", id, message, result)
	})
	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		fmt.Fprintf(os.Stderr, "afterCallTool: %v, %v, %v\n", id, message, result)
	})
	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		fmt.Fprintf(os.Stderr, "beforeCallTool: %v, %v\n", id, message)
	})
	// Define the task handler
	// Define the tools
	tools := 
	handler := createTaskHandler(memoryPath)

	s := server.NewMCPServer(serverName, "1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithHooks(hooks),
	)
	for _, tool := range tools {
		s.AddTool(*tool, handler) // Dereference tool
	}

	err = server.ServeStdio(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serving MCP: %v\n", err)
	}
}
