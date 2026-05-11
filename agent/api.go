package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAI-compatible request/response types

type CompletionRequest struct {
	Model    string        `json:"model"`
	Messages []Message     `json:"messages"`
	Tools    []ToolRequest `json:"tools,omitempty"`
}

type Message struct {
	Role        string          `json:"role"`
	Content     string          `json:"content"`
	ToolCalls   []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID  string          `json:"tool_call_id,omitempty"`
}

// ToolRequest represents the tool definition sent to the API
type ToolRequest struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents the function details of a tool
type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// ToolDefinition represents a locally registered tool
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
	Function    func(input json.RawMessage) (string, error) `json:"-"`
}

type ToolCall struct {
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type CompletionResponse struct {
	Choices []struct {
		Message struct {
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// sendCompletion sends a chat completion request and returns the assistant's response.
func (app *appEnv) sendCompletion(messages []Message) (string, error) {
	// Convert registered tools to API format
	var toolRequests []ToolRequest
	for _, tool := range app.tools {
		toolRequests = append(toolRequests, ToolRequest{
			Type: "function",
			Function: ToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		})
	}

	reqBody := CompletionRequest{
		Model:    app.model,
		Messages: messages,
		Tools:    toolRequests,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", app.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+app.apiKey)

	resp, err := app.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var response CompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	assistantMsg := response.Choices[0].Message
	// Check if the assistant's message includes tool calls
	if len(assistantMsg.ToolCalls) > 0 {
		// Build new message history including the assistant message (with tool calls)
		// and add tool result messages
		newMessages := make([]Message, len(messages))
		copy(newMessages, messages)
		newMessages = append(newMessages, Message{
			Role:    "assistant",
			Content: assistantMsg.Content,
			ToolCalls: assistantMsg.ToolCalls,
		})

		// Execute each tool call and add tool result messages
		for _, toolCall := range assistantMsg.ToolCalls {
			toolResult := app.executeTool(toolCall)
			if toolResult.Error != nil {
				return "", fmt.Errorf("tool execution failed: %w", toolResult.Error)
			}
			newMessages = append(newMessages, Message{
				Role:       "tool",
				Content:    toolResult.Content,
				ToolCallID: toolCall.Function.Name,
			})
		}

		// Send a follow-up request with the tool results
		return app.sendCompletion(newMessages)
	}

	return assistantMsg.Content, nil
}

// executeTool finds and executes a tool based on the tool call
func (app *appEnv) executeTool(call ToolCall) (result struct {
	Content string
	Error   error
}) {
	// Find the tool definition by name
	for _, tool := range app.tools {
		if tool.Name == call.Function.Name {
			// Execute the tool function with the arguments
			content, err := tool.Function(json.RawMessage(call.Function.Arguments))
			if err != nil {
				return struct {
					Content string
					Error   error
				}{Error: fmt.Errorf("tool %s execution failed: %w", tool.Name, err)}
			}
			return struct {
				Content string
				Error   error
			}{Content: content}
		}
	}
	return struct {
		Content string
		Error   error
	}{Error: fmt.Errorf("unknown tool: %s", call.Function.Name)}
}
