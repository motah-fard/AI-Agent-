package models

type GeneratePlanRequest struct {
	AppName     string   `json:"app_name"`
	Idea        string   `json:"idea"`
	TargetUsers []string `json:"target_users"`
	Constraints []string `json:"constraints"`
}

type Estimate struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

type Task struct {
	ID                 string   `json:"id,omitempty"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Priority           string   `json:"priority"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Estimate           Estimate `json:"estimate"`
	Dependencies       []string `json:"dependencies"`
	StoryID            string   `json:"story_id,omitempty"`
}

type Story struct {
	ID                 string   `json:"id,omitempty"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Priority           string   `json:"priority"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Estimate           Estimate `json:"estimate"`
	Dependencies       []string `json:"dependencies"`
	Tasks              []Task   `json:"tasks"`
	EpicID             string   `json:"epic_id,omitempty"`
}

type Epic struct {
	ID                 string   `json:"id,omitempty"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Priority           string   `json:"priority"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Estimate           Estimate `json:"estimate"`
	Dependencies       []string `json:"dependencies"`
	Stories            []Story  `json:"stories"`
	ProjectID          string   `json:"project_id,omitempty"`
}

type GeneratePlanResponse struct {
	AppName        string   `json:"app_name"`
	ProjectID      string   `json:"project_id,omitempty"`
	ProjectSummary string   `json:"project_summary"`
	MVPScope       []string `json:"mvp_scope"`
	Assumptions    []string `json:"assumptions"`
	Risks          []string `json:"risks"`
	Epics          []Epic   `json:"epics"`
}
type RefineProjectRequest struct {
	Instruction string `json:"instruction"`
}
type JiraPreviewRequest struct {
	ProjectKey string `json:"project_key"`
}
