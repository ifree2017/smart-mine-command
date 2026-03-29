package ai

import (
	"context"
	"testing"
)

// mockLLMClient is a test double for LLMClient
type mockLLMClient struct {
	resp string
	err  error
}

func (m *mockLLMClient) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	return m.resp, m.err
}

func TestNewClient(t *testing.T) {
	c := NewClient("https://api.openai.com", "test-token")
	if c.baseURL != "https://api.openai.com" {
		t.Errorf("baseURL: got %s", c.baseURL)
	}
	if c.token != "test-token" {
		t.Errorf("token: got %s", c.token)
	}
}

func TestChatMessage_Fields(t *testing.T) {
	msg := ChatMessage{Role: "user", Content: "hello"}
	if msg.Role != "user" {
		t.Errorf("Role: got %s", msg.Role)
	}
	if msg.Content != "hello" {
		t.Errorf("Content: got %s", msg.Content)
	}
}

func TestChatResponse_Choices(t *testing.T) {
	resp := ChatResponse{
		Choices: []Choice{
			{Message: ChatMessage{Role: "assistant", Content: "hi"}},
		},
	}
	if len(resp.Choices) != 1 {
		t.Errorf("Choices: got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "hi" {
		t.Errorf("Content: got %s", resp.Choices[0].Message.Content)
	}
}

func TestChatRequest_JSON(t *testing.T) {
	req := ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMessage{
			{Role: "user", Content: "hello"},
		},
	}
	if req.Model != "gpt-4o-mini" {
		t.Errorf("Model: got %s", req.Model)
	}
	if len(req.Messages) != 1 {
		t.Errorf("Messages len: got %d", len(req.Messages))
	}
}
