package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewMCPServer(t *testing.T) {
	server := NewMCPServer("test-server", "1.0.0")
	if server == nil {
		t.Fatal("NewMCPServer returned nil")
	}

	if server.name != "test-server" {
		t.Errorf("Expected name 'test-server', got '%s'", server.name)
	}

	if server.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", server.version)
	}

	if server.tools == nil {
		t.Error("Tools map is nil")
	}

	if server.resources == nil {
		t.Error("Resources map is nil")
	}

	if server.prompts == nil {
		t.Error("Prompts map is nil")
	}
}

func TestRegisterTool(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	tool := &Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	}

	server.RegisterTool(tool)

	server.mu.RLock()
	registered, exists := server.tools["test_tool"]
	server.mu.RUnlock()

	if !exists {
		t.Fatal("Tool was not registered")
	}

	if registered.Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", registered.Name)
	}
}

func TestRegisterResource(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	resource := &Resource{
		URI:         "file:///test.txt",
		Name:        "test resource",
		Description: "A test resource",
		MimeType:    "text/plain",
		Handler: func(ctx context.Context) (interface{}, error) {
			return "test content", nil
		},
	}

	server.RegisterResource(resource)

	server.mu.RLock()
	registered, exists := server.resources["file:///test.txt"]
	server.mu.RUnlock()

	if !exists {
		t.Fatal("Resource was not registered")
	}

	if registered.Name != "test resource" {
		t.Errorf("Expected resource name 'test resource', got '%s'", registered.Name)
	}
}

func TestRegisterPrompt(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	prompt := &Prompt{
		Name:        "test_prompt",
		Description: "A test prompt",
		Arguments: []PromptArgument{
			{Name: "arg1", Description: "First argument", Required: true},
		},
		Handler: func(ctx context.Context, args map[string]string) (string, error) {
			return "test prompt text", nil
		},
	}

	server.RegisterPrompt(prompt)

	server.mu.RLock()
	registered, exists := server.prompts["test_prompt"]
	server.mu.RUnlock()

	if !exists {
		t.Fatal("Prompt was not registered")
	}

	if registered.Name != "test_prompt" {
		t.Errorf("Expected prompt name 'test_prompt', got '%s'", registered.Name)
	}
}

func TestHandleInitialize(t *testing.T) {
	server := NewMCPServer("test-server", "1.0.0")

	result, err := server.handleInitialize(context.Background(), nil)
	if err != nil {
		t.Fatalf("handleInitialize failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["protocolVersion"] != "2024-11-05" {
		t.Error("Protocol version mismatch")
	}

	serverInfo, ok := resultMap["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("serverInfo is not a map")
	}

	if serverInfo["name"] != "test-server" {
		t.Error("Server name mismatch")
	}

	if serverInfo["version"] != "1.0.0" {
		t.Error("Server version mismatch")
	}
}

func TestHandleToolsList(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	tool := &Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "result", nil
		},
	}

	server.RegisterTool(tool)

	result, err := server.handleToolsList(context.Background(), nil)
	if err != nil {
		t.Fatalf("handleToolsList failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	tools, ok := resultMap["tools"].([]map[string]interface{})
	if !ok {
		t.Fatal("tools is not an array")
	}

	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	if tools[0]["name"] != "test_tool" {
		t.Error("Tool name mismatch")
	}
}

func TestHandleToolsCall(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	called := false
	tool := &Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			called = true
			return "success", nil
		},
	}

	server.RegisterTool(tool)

	params := map[string]interface{}{
		"name":      "test_tool",
		"arguments": map[string]interface{}{},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handleToolsCall(context.Background(), paramsJSON)
	if err != nil {
		t.Fatalf("handleToolsCall failed: %v", err)
	}

	if !called {
		t.Error("Tool handler was not called")
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	content, ok := resultMap["content"].([]map[string]interface{})
	if !ok || len(content) == 0 {
		t.Fatal("Content is missing or invalid")
	}
}

func TestHandleToolsCallNotFound(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	params := map[string]interface{}{
		"name":      "nonexistent_tool",
		"arguments": map[string]interface{}{},
	}
	paramsJSON, _ := json.Marshal(params)

	_, err := server.handleToolsCall(context.Background(), paramsJSON)
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestHandleResourcesList(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	resource := &Resource{
		URI:         "file:///test.txt",
		Name:        "test",
		Description: "test resource",
		MimeType:    "text/plain",
		Handler: func(ctx context.Context) (interface{}, error) {
			return "content", nil
		},
	}

	server.RegisterResource(resource)

	result, err := server.handleResourcesList(context.Background(), nil)
	if err != nil {
		t.Fatalf("handleResourcesList failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	resources, ok := resultMap["resources"].([]map[string]interface{})
	if !ok {
		t.Fatal("resources is not an array")
	}

	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}
}

func TestHandleResourcesRead(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	resource := &Resource{
		URI:         "file:///test.txt",
		Name:        "test",
		Description: "test resource",
		MimeType:    "text/plain",
		Handler: func(ctx context.Context) (interface{}, error) {
			return "test content", nil
		},
	}

	server.RegisterResource(resource)

	params := map[string]interface{}{
		"uri": "file:///test.txt",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handleResourcesRead(context.Background(), paramsJSON)
	if err != nil {
		t.Fatalf("handleResourcesRead failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	contents, ok := resultMap["contents"].([]map[string]interface{})
	if !ok || len(contents) == 0 {
		t.Fatal("Contents is missing or invalid")
	}
}

func TestHandlePromptsList(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	prompt := &Prompt{
		Name:        "test_prompt",
		Description: "test prompt",
		Arguments: []PromptArgument{
			{Name: "arg1", Description: "test arg", Required: true},
		},
		Handler: func(ctx context.Context, args map[string]string) (string, error) {
			return "prompt text", nil
		},
	}

	server.RegisterPrompt(prompt)

	result, err := server.handlePromptsList(context.Background(), nil)
	if err != nil {
		t.Fatalf("handlePromptsList failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	prompts, ok := resultMap["prompts"].([]map[string]interface{})
	if !ok {
		t.Fatal("prompts is not an array")
	}

	if len(prompts) != 1 {
		t.Errorf("Expected 1 prompt, got %d", len(prompts))
	}
}

func TestHandlePromptsGet(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	prompt := &Prompt{
		Name:        "test_prompt",
		Description: "test prompt",
		Arguments:   []PromptArgument{},
		Handler: func(ctx context.Context, args map[string]string) (string, error) {
			return "generated prompt text", nil
		},
	}

	server.RegisterPrompt(prompt)

	params := map[string]interface{}{
		"name":      "test_prompt",
		"arguments": map[string]string{},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handlePromptsGet(context.Background(), paramsJSON)
	if err != nil {
		t.Fatalf("handlePromptsGet failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	messages, ok := resultMap["messages"].([]map[string]interface{})
	if !ok || len(messages) == 0 {
		t.Fatal("Messages is missing or invalid")
	}
}

func TestHandleRequest(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  nil,
	}

	response := server.handleRequest(req)

	if response.JSONRPC != "2.0" {
		t.Error("JSONRPC version mismatch")
	}

	if response.ID != 1 {
		t.Error("ID mismatch")
	}

	if response.Error != nil {
		t.Errorf("Unexpected error: %v", response.Error)
	}

	if response.Result == nil {
		t.Error("Result is nil")
	}
}

func TestHandleRequestMethodNotFound(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "nonexistent_method",
		Params:  nil,
	}

	response := server.handleRequest(req)

	if response.Error == nil {
		t.Error("Expected error for nonexistent method")
	}

	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", response.Error.Code)
	}
}

func TestMCPServerStartStop(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	// Test stop
	server.Stop()

	// Verify context is cancelled
	select {
	case <-server.ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context was not cancelled")
	}
}

func TestMCPRequestResponse(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      "test-id",
		Method:  "test_method",
		Params:  json.RawMessage(`{"key":"value"}`),
	}

	if req.JSONRPC != "2.0" {
		t.Error("JSONRPC version mismatch")
	}

	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      "test-id",
		Result:  map[string]string{"status": "ok"},
	}

	if resp.Error != nil {
		t.Error("Error should be nil")
	}
}

func TestMCPError(t *testing.T) {
	err := &MCPError{
		Code:    -32600,
		Message: "Invalid Request",
		Data:    map[string]string{"detail": "missing field"},
	}

	if err.Code != -32600 {
		t.Error("Error code mismatch")
	}

	if err.Message != "Invalid Request" {
		t.Error("Error message mismatch")
	}
}

func TestToolHandler(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		name, ok := args["name"].(string)
		if !ok {
			return nil, nil
		}
		return "Hello, " + name, nil
	}

	result, err := handler(context.Background(), map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	if result != "Hello, World" {
		t.Errorf("Expected 'Hello, World', got '%v'", result)
	}
}

func TestResourceHandler(t *testing.T) {
	handler := func(ctx context.Context) (interface{}, error) {
		return "resource content", nil
	}

	result, err := handler(context.Background())
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	if result != "resource content" {
		t.Errorf("Expected 'resource content', got '%v'", result)
	}
}

func TestPromptHandler(t *testing.T) {
	handler := func(ctx context.Context, args map[string]string) (string, error) {
		topic := args["topic"]
		return "Write about " + topic, nil
	}

	result, err := handler(context.Background(), map[string]string{"topic": "Go programming"})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	if result != "Write about Go programming" {
		t.Errorf("Expected 'Write about Go programming', got '%s'", result)
	}
}

func TestMCPServerConcurrency(t *testing.T) {
	server := NewMCPServer("test", "1.0")

	// Register multiple tools concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			tool := &Tool{
				Name:        string(rune('A' + id)),
				Description: "test",
				InputSchema: map[string]interface{}{},
				Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return "ok", nil
				},
			}
			server.RegisterTool(tool)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	server.mu.RLock()
	toolCount := len(server.tools)
	server.mu.RUnlock()

	if toolCount != 10 {
		t.Errorf("Expected 10 tools, got %d", toolCount)
	}
}

func TestMCPServerJSONEncoding(t *testing.T) {
	// Create a request
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	// Encode request
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(req); err != nil {
		t.Fatalf("Failed to encode request: %v", err)
	}

	// Decode request
	decoder := json.NewDecoder(&buf)
	var decoded MCPRequest
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode request: %v", err)
	}

	if decoded.Method != "initialize" {
		t.Error("Method mismatch after encoding/decoding")
	}
}

func TestPromptArgument(t *testing.T) {
	arg := PromptArgument{
		Name:        "test_arg",
		Description: "A test argument",
		Required:    true,
	}

	if arg.Name != "test_arg" {
		t.Error("Name mismatch")
	}

	if !arg.Required {
		t.Error("Required should be true")
	}
}
