package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/motah-fard/ai-agent/backend/internal/models"
)

type Planner struct {
	client *Client
}

func NewPlanner(client *Client) *Planner {
	return &Planner{
		client: client,
	}
}

func (p *Planner) GeneratePlan(ctx context.Context, req models.GeneratePlanRequest) (*models.GeneratePlanResponse, error) {
	prompt := BuildPlanningPrompt(req)

	raw, err := p.client.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	cleaned := strings.TrimSpace(raw)

	var result models.GeneratePlanResponse
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse model response: %w; raw response: %s", err, cleaned)
	}

	if err := validatePlan(result); err != nil {
		return nil, err
	}

	return &result, nil
}
func validatePlan(plan models.GeneratePlanResponse) error {
	if strings.TrimSpace(plan.ProjectSummary) == "" {
		return errors.New("missing project_summary")
	}
	if len(plan.MVPScope) == 0 {
		return errors.New("missing mvp_scope")
	}
	if len(plan.Epics) == 0 {
		return errors.New("missing epics")
	}

	for _, epic := range plan.Epics {
		if strings.TrimSpace(epic.Title) == "" {
			return errors.New("epic missing title")
		}
		if len(epic.Stories) == 0 {
			return errors.New("epic missing stories")
		}
		for _, story := range epic.Stories {
			if strings.TrimSpace(story.Title) == "" {
				return errors.New("story missing title")
			}
			if len(story.Tasks) == 0 {
				return errors.New("story missing tasks")
			}
			for _, task := range story.Tasks {
				if strings.TrimSpace(task.Title) == "" {
					return errors.New("task missing title")
				}
			}
		}
	}

	return nil
}
func (p *Planner) RegeneratePlan(ctx context.Context, existing *models.GeneratePlanResponse) (*models.GeneratePlanResponse, error) {
	prompt := BuildRegeneratePrompt(existing)

	raw, err := p.client.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	cleaned := stripCodeFences(raw)

	var result models.GeneratePlanResponse
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse regenerated model response: %w; raw response: %s", err, cleaned)
	}

	if err := validatePlan(result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *Planner) RefinePlan(ctx context.Context, existing *models.GeneratePlanResponse, instruction string) (*models.GeneratePlanResponse, error) {
	prompt := BuildRefinePrompt(existing, instruction)

	raw, err := p.client.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	cleaned := stripCodeFences(raw)

	var result models.GeneratePlanResponse
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse refined model response: %w; raw response: %s", err, cleaned)
	}

	if err := validatePlan(result); err != nil {
		return nil, err
	}

	return &result, nil
}
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	}
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}

	return strings.TrimSpace(s)
}
