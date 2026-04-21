package jira

type Config struct {
	BaseURL         string
	Email           string
	APIToken        string
	ProjectKey      string
	EpicLinkFieldID string
}

type IssueType string

const (
	IssueTypeEpic    IssueType = "Epic"
	IssueTypeStory   IssueType = "Story"
	IssueTypeSubTask IssueType = "Task"
)

type PreviewRequest struct {
	ProjectID  string `json:"project_id"`
	ProjectKey string `json:"project_key"`
}

type PreviewIssue struct {
	LocalID     string            `json:"local_id"`
	IssueType   IssueType         `json:"issue_type"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	ParentRef   string            `json:"parent_ref,omitempty"`
	Priority    string            `json:"priority"`
	Labels      []string          `json:"labels,omitempty"`
	Fields      map[string]any    `json:"fields,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
}

type PreviewResponse struct {
	ProjectID      string         `json:"project_id"`
	AppName        string         `json:"app_name"`
	JiraProjectKey string         `json:"jira_project_key"`
	Issues         []PreviewIssue `json:"issues"`
}
type ExportRequest struct {
	ProjectKey string `json:"project_key"`
}

type ExportedIssue struct {
	LocalID   string    `json:"local_id"`
	IssueType IssueType `json:"issue_type"`
	Summary   string    `json:"summary"`
	JiraKey   string    `json:"jira_key"`
	JiraID    string    `json:"jira_id"`
	ParentRef string    `json:"parent_ref,omitempty"`
}

type ExportResponse struct {
	ProjectID      string          `json:"project_id"`
	AppName        string          `json:"app_name"`
	JiraProjectKey string          `json:"jira_project_key"`
	Created        []ExportedIssue `json:"created"`
}
