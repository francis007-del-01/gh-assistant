package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client provides Jira API operations
type Client struct {
	baseURL  string
	email    string
	apiToken string
	project  string
}

// Config holds Jira client configuration
type Config struct {
	BaseURL  string // e.g., https://yourcompany.atlassian.net
	Email    string
	APIToken string
	Project  string // Project key, e.g., "PROJ"
}

// Issue represents a Jira issue
type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields struct {
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
	} `json:"fields"`
}

// createIssueRequest represents the request body for creating an issue
type createIssueRequest struct {
	Fields createIssueFields `json:"fields"`
}

type createIssueFields struct {
	Project   projectField   `json:"project"`
	Summary   string         `json:"summary"`
	IssueType issueTypeField `json:"issuetype"`
}

type projectField struct {
	Key string `json:"key"`
}

type issueTypeField struct {
	Name string `json:"name"`
}

// transitionRequest represents a transition request
type transitionRequest struct {
	Transition transitionField `json:"transition"`
}

type transitionField struct {
	ID string `json:"id"`
}

// transitionsResponse represents the response from getting transitions
type transitionsResponse struct {
	Transitions []transition `json:"transitions"`
}

type transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   struct {
		Name string `json:"name"`
	} `json:"to"`
}

// New creates a new Jira client
func New(cfg Config) *Client {
	return &Client{
		baseURL:  cfg.BaseURL,
		email:    cfg.Email,
		apiToken: cfg.APIToken,
		project:  cfg.Project,
	}
}

// IsConfigured returns true if Jira is properly configured
func (c *Client) IsConfigured() bool {
	return c.baseURL != "" && c.email != "" && c.apiToken != "" && c.project != ""
}

// CreateIssue creates a new Jira issue and returns the created issue
func (c *Client) CreateIssue(summary string) (*Issue, error) {
	reqBody := createIssueRequest{
		Fields: createIssueFields{
			Project:   projectField{Key: c.project},
			Summary:   summary,
			IssueType: issueTypeField{Name: "Task"},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/rest/api/3/issue", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	var issue Issue
	if err := json.Unmarshal(body, &issue); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &issue, nil
}

// TransitionToInProgress moves the issue to "In Progress" status
func (c *Client) TransitionToInProgress(issueKey string) error {
	// First, get available transitions
	transitions, err := c.getTransitions(issueKey)
	if err != nil {
		return err
	}

	// Find the "In Progress" transition
	var inProgressID string
	for _, t := range transitions {
		// Check both transition name and target status name (case-insensitive matching)
		if t.Name == "In Progress" || t.To.Name == "In Progress" ||
			t.Name == "Start Progress" || t.Name == "Start" {
			inProgressID = t.ID
			break
		}
	}

	if inProgressID == "" {
		// If no specific transition found, try common variations
		for _, t := range transitions {
			if t.To.Name == "In Progress" {
				inProgressID = t.ID
				break
			}
		}
	}

	if inProgressID == "" {
		return fmt.Errorf("no 'In Progress' transition available for issue %s", issueKey)
	}

	// Execute the transition
	return c.doTransition(issueKey, inProgressID)
}

func (c *Client) getTransitions(issueKey string) ([]transition, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/rest/api/3/issue/"+issueKey+"/transitions", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	var transResp transitionsResponse
	if err := json.Unmarshal(body, &transResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return transResp.Transitions, nil
}

func (c *Client) doTransition(issueKey, transitionID string) error {
	reqBody := transitionRequest{
		Transition: transitionField{ID: transitionID},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/rest/api/3/issue/"+issueKey+"/transitions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// CreateIssueWithTitle creates a Jira issue with title format "JIRA-ID - message"
// and transitions it to In Progress. Returns the formatted title.
func (c *Client) CreateIssueWithTitle(commitMessage string) (string, error) {
	// Create the issue first (with just the commit message as summary)
	issue, err := c.CreateIssue(commitMessage)
	if err != nil {
		return "", fmt.Errorf("failed to create issue: %w", err)
	}

	// Transition to In Progress
	if err := c.TransitionToInProgress(issue.Key); err != nil {
		// Don't fail completely, just warn - the issue was created
		fmt.Printf("⚠️  Warning: Could not transition to In Progress: %v\n", err)
	}

	// Return the formatted title
	return fmt.Sprintf("%s - %s", issue.Key, commitMessage), nil
}

// GetIssueURL returns the browser URL for an issue
func (c *Client) GetIssueURL(issueKey string) string {
	return fmt.Sprintf("%s/browse/%s", c.baseURL, issueKey)
}

