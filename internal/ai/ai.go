package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Provider represents an AI provider
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
)

// Client handles AI API interactions
type Client struct {
	provider   Provider
	apiKey     string
	model      string
	httpClient *http.Client
}

// Config holds AI client configuration
type Config struct {
	Provider Provider
	APIKey   string
	Model    string
}

// New creates a new AI client
func New(cfg Config) *Client {
	if cfg.Model == "" {
		switch cfg.Provider {
		case ProviderOpenAI:
			cfg.Model = "gpt-4o-mini"
		case ProviderAnthropic:
			cfg.Model = "claude-3-5-sonnet-20241022"
		}
	}

	return &Client{
		provider: cfg.Provider,
		apiKey:   cfg.APIKey,
		model:    cfg.Model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateCommitMessage generates a commit message from a git diff
func (c *Client) GenerateCommitMessage(diff string, changedFiles []string) (string, error) {
	if diff == "" {
		return "", errors.New("no diff provided")
	}

	prompt := buildCommitPrompt(diff, changedFiles)

	switch c.provider {
	case ProviderOpenAI:
		return c.callOpenAI(prompt)
	case ProviderAnthropic:
		return c.callAnthropic(prompt)
	default:
		return "", fmt.Errorf("unsupported provider: %s", c.provider)
	}
}

func buildCommitPrompt(diff string, changedFiles []string) string {
	// Truncate diff if too long
	maxDiffLen := 12000
	truncatedDiff := diff
	if len(diff) > maxDiffLen {
		truncatedDiff = diff[:maxDiffLen] + "\n... [diff truncated]"
	}

	filesContext := ""
	if len(changedFiles) > 0 {
		filesContext = fmt.Sprintf("\nChanged files:\n- %s\n", strings.Join(changedFiles, "\n- "))
	}

	return fmt.Sprintf(`You are an expert at writing clear, concise git commit messages following conventional commits format.

Analyze the following git diff and generate a meaningful commit message.
%s
Git Diff:
%s

Rules for the commit message:
1. Use conventional commits format: type(scope): description
2. Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore
3. Keep the first line under 72 characters
4. Be specific about what changed and why
5. If there are multiple unrelated changes, focus on the main one
6. Do NOT include any explanation, just the commit message
7. Do NOT wrap in quotes or code blocks

Respond with ONLY the commit message, nothing else.`, filesContext, truncatedDiff)
}

// OpenAI API types
type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) callOpenAI(prompt string) (string, error) {
	reqBody := openAIRequest{
		Model: c.model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result openAIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", errors.New("no response from API")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

// Anthropic API types
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) callAnthropic(prompt string) (string, error) {
	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: 256,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result anthropicResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Content) == 0 {
		return "", errors.New("no response from API")
	}

	return strings.TrimSpace(result.Content[0].Text), nil
}

