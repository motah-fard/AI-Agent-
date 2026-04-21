package jira

import (
	"fmt"
	"strings"

	"github.com/motah-fard/ai-agent/backend/internal/models"
)

func BuildPreview(plan *models.GeneratePlanResponse, jiraProjectKey string) *PreviewResponse {
	resp := &PreviewResponse{
		ProjectID:      plan.ProjectID,
		AppName:        plan.AppName,
		JiraProjectKey: jiraProjectKey,
		Issues:         []PreviewIssue{},
	}

	for _, epic := range plan.Epics {
		epicIssue := PreviewIssue{
			LocalID:     epic.ID,
			IssueType:   IssueTypeEpic,
			Summary:     epic.Title,
			Description: buildEpicDescription(epic),
			Priority:    epic.Priority,
			Labels:      []string{"ai-agent", "generated", "epic"},
			Fields: map[string]any{
				"project_key": jiraProjectKey,
			},
			Meta: map[string]string{
				"level":       "epic",
				"description": epic.Description,
			},
		}
		resp.Issues = append(resp.Issues, epicIssue)

		for _, story := range epic.Stories {
			storyIssue := PreviewIssue{
				LocalID:     story.ID,
				IssueType:   IssueTypeStory,
				Summary:     story.Title,
				Description: buildStoryDescription(story),
				ParentRef:   epic.ID,
				Priority:    story.Priority,
				Labels:      []string{"ai-agent", "generated", "story"},
				Fields: map[string]any{
					"project_key": jiraProjectKey,
				},
				Meta: map[string]string{
					"level":       "story",
					"epic_ref":    epic.ID,
					"description": story.Description,
				},
			}
			resp.Issues = append(resp.Issues, storyIssue)

			for _, task := range story.Tasks {
				taskIssue := PreviewIssue{
					LocalID:     task.ID,
					IssueType:   IssueTypeSubTask,
					Summary:     task.Title,
					Description: buildTaskDescription(task),
					ParentRef:   story.ID,
					Priority:    task.Priority,
					Labels:      []string{"ai-agent", "generated", "task"},
					Fields: map[string]any{
						"project_key": jiraProjectKey,
					},
					Meta: map[string]string{
						"level":       "task",
						"story_ref":   story.ID,
						"description": task.Description,
					},
				}
				resp.Issues = append(resp.Issues, taskIssue)
			}
		}
	}

	return resp
}

func buildEpicDescription(epic models.Epic) string {
	return joinSections(
		fmt.Sprintf("**Description**\n%s", epic.Description),
		fmt.Sprintf("**Acceptance Criteria**\n%s", bulletList(epic.AcceptanceCriteria)),
		fmt.Sprintf("**Estimate**\n%d %s", epic.Estimate.Value, epic.Estimate.Unit),
		fmt.Sprintf("**Dependencies**\n%s", bulletListOrNone(epic.Dependencies)),
	)
}

func buildStoryDescription(story models.Story) string {
	return joinSections(
		fmt.Sprintf("**Description**\n%s", story.Description),
		fmt.Sprintf("**Acceptance Criteria**\n%s", bulletList(story.AcceptanceCriteria)),
		fmt.Sprintf("**Estimate**\n%d %s", story.Estimate.Value, story.Estimate.Unit),
		fmt.Sprintf("**Dependencies**\n%s", bulletListOrNone(story.Dependencies)),
	)
}

func buildTaskDescription(task models.Task) string {
	return joinSections(
		fmt.Sprintf("**Description**\n%s", task.Description),
		fmt.Sprintf("**Acceptance Criteria**\n%s", bulletList(task.AcceptanceCriteria)),
		fmt.Sprintf("**Estimate**\n%d %s", task.Estimate.Value, task.Estimate.Unit),
		fmt.Sprintf("**Dependencies**\n%s", bulletListOrNone(task.Dependencies)),
	)
}

func bulletList(items []string) string {
	if len(items) == 0 {
		return "- None"
	}

	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, "- "+item)
	}
	return strings.Join(lines, "\n")
}

func bulletListOrNone(items []string) string {
	if len(items) == 0 {
		return "- None"
	}
	return bulletList(items)
}

func joinSections(parts ...string) string {
	return strings.Join(parts, "\n\n")
}
