package jira

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL         string
	email           string
	apiToken        string
	defaultProjKey  string
	epicLinkFieldID string
	httpClient      *http.Client
}

func LoadConfigFromEnv() (Config, error) {
	baseURL := strings.TrimSpace(os.Getenv("JIRA_BASE_URL"))
	email := strings.TrimSpace(os.Getenv("JIRA_EMAIL"))
	apiToken := strings.TrimSpace(os.Getenv("JIRA_API_TOKEN"))
	projectKey := strings.TrimSpace(os.Getenv("JIRA_PROJECT_KEY"))
	epicLinkFieldID := strings.TrimSpace(os.Getenv("JIRA_EPIC_LINK_FIELD_ID"))

	if baseURL == "" {
		return Config{}, fmt.Errorf("JIRA_BASE_URL is required")
	}
	if email == "" {
		return Config{}, fmt.Errorf("JIRA_EMAIL is required")
	}
	if apiToken == "" {
		return Config{}, fmt.Errorf("JIRA_API_TOKEN is required")
	}

	return Config{
		BaseURL:         strings.TrimRight(baseURL, "/"),
		Email:           email,
		APIToken:        apiToken,
		ProjectKey:      projectKey,
		EpicLinkFieldID: epicLinkFieldID,
	}, nil
}

func NewClient(cfg Config) *Client {
	return &Client{
		baseURL:         strings.TrimRight(cfg.BaseURL, "/"),
		email:           cfg.Email,
		apiToken:        cfg.APIToken,
		defaultProjKey:  cfg.ProjectKey,
		epicLinkFieldID: cfg.EpicLinkFieldID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type createIssueRequest struct {
	Fields map[string]any `json:"fields"`
}

type createIssueResponse struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

func (c *Client) CreateIssuesFromPreview(ctx context.Context, preview *PreviewResponse) (*ExportResponse, error) {
	if preview == nil {
		return nil, fmt.Errorf("preview is nil")
	}
	if strings.TrimSpace(preview.JiraProjectKey) == "" {
		if c.defaultProjKey == "" {
			return nil, fmt.Errorf("jira project key is required")
		}
		preview.JiraProjectKey = c.defaultProjKey
	}

	result := &ExportResponse{
		ProjectID:      preview.ProjectID,
		AppName:        preview.AppName,
		JiraProjectKey: preview.JiraProjectKey,
		Created:        []ExportedIssue{},
	}

	createdKeysByLocalID := map[string]string{}

	// pass 1: create epics
	for _, issue := range preview.Issues {
		if issue.IssueType != IssueTypeEpic {
			continue
		}

		created, err := c.createOne(ctx, preview.JiraProjectKey, issue, createdKeysByLocalID)
		if err != nil {
			return nil, fmt.Errorf("create epic %s: %w", issue.LocalID, err)
		}

		createdKeysByLocalID[issue.LocalID] = created.JiraKey
		result.Created = append(result.Created, *created)
	}

	// pass 2: create stories
	for _, issue := range preview.Issues {
		if issue.IssueType != IssueTypeStory {
			continue
		}

		created, err := c.createOne(ctx, preview.JiraProjectKey, issue, createdKeysByLocalID)
		if err != nil {
			return nil, fmt.Errorf("create story %s: %w", issue.LocalID, err)
		}

		createdKeysByLocalID[issue.LocalID] = created.JiraKey
		result.Created = append(result.Created, *created)
	}

	// pass 3: create subtasks
	for _, issue := range preview.Issues {
		if issue.IssueType != IssueTypeSubTask {
			continue
		}

		created, err := c.createOne(ctx, preview.JiraProjectKey, issue, createdKeysByLocalID)
		if err != nil {
			return nil, fmt.Errorf("create sub-task %s: %w", issue.LocalID, err)
		}

		createdKeysByLocalID[issue.LocalID] = created.JiraKey
		result.Created = append(result.Created, *created)
	}

	return result, nil
}

func (c *Client) createOne(ctx context.Context, projectKey string, issue PreviewIssue, createdKeysByLocalID map[string]string) (*ExportedIssue, error) {
	fields := map[string]any{
		"project": map[string]any{
			"key": projectKey,
		},
		"summary": issue.Summary,
		"issuetype": map[string]any{
			"name": string(issue.IssueType),
		},
		"description": adfDocFromText(issue.Description),
	}

	if len(issue.Labels) > 0 {
		fields["labels"] = issue.Labels
	}

	if strings.TrimSpace(issue.Priority) != "" {
		fields["priority"] = map[string]any{
			"name": normalizePriority(issue.Priority),
		}
	}

	// For now, task-level items are exported as normal Jira Tasks,
	// so no parent field is sent. If in the future we want to use Sub-tasks, we would need to set the "parent" field here with the parent issue key.

	// Optional Epic Link for stories if configured
	if issue.IssueType == IssueTypeStory &&
		strings.TrimSpace(issue.ParentRef) != "" &&
		strings.TrimSpace(c.epicLinkFieldID) != "" {
		parentEpicKey, ok := createdKeysByLocalID[issue.ParentRef]
		if ok {
			fields[c.epicLinkFieldID] = parentEpicKey
		}
	}

	reqBody := createIssueRequest{
		Fields: fields,
	}

	created, err := c.createIssue(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	return &ExportedIssue{
		LocalID:   issue.LocalID,
		IssueType: issue.IssueType,
		Summary:   issue.Summary,
		JiraKey:   created.Key,
		JiraID:    created.ID,
		ParentRef: issue.ParentRef,
	}, nil
}

func (c *Client) createIssue(ctx context.Context, payload createIssueRequest) (*createIssueResponse, error) {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal jira create issue payload: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/rest/api/3/issue",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("build jira create issue request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+basicAuthToken(c.email, c.apiToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send jira create issue request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read jira create issue response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira create issue failed: status=%d body=%s", resp.StatusCode, string(respBytes))
	}

	var created createIssueResponse
	if err := json.Unmarshal(respBytes, &created); err != nil {
		return nil, fmt.Errorf("unmarshal jira create issue response: %w", err)
	}

	if created.Key == "" || created.ID == "" {
		return nil, fmt.Errorf("jira create issue response missing id/key: %s", string(respBytes))
	}

	return &created, nil
}

func basicAuthToken(email, token string) string {
	raw := email + ":" + token
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func normalizePriority(p string) string {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case "highest":
		return "Highest"
	case "high":
		return "High"
	case "medium":
		return "Medium"
	case "low":
		return "Low"
	case "lowest":
		return "Lowest"
	default:
		return "Medium"
	}
}

func adfDocFromText(text string) map[string]any {
	lines := splitNonEmptyLines(text)

	content := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		content = append(content, map[string]any{
			"type": "paragraph",
			"content": []map[string]any{
				{
					"type": "text",
					"text": line,
				},
			},
		})
	}

	if len(content) == 0 {
		content = []map[string]any{
			{
				"type": "paragraph",
				"content": []map[string]any{
					{
						"type": "text",
						"text": "",
					},
				},
			},
		}
	}

	return map[string]any{
		"type":    "doc",
		"version": 1,
		"content": content,
	}
}

func splitNonEmptyLines(s string) []string {
	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}
