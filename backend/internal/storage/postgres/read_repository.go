package postgres

import (
	"context"
	"fmt"

	"github.com/motah-fard/ai-agent/backend/internal/models"
)

func (r *PostgresRepository) ListProjects(ctx context.Context) ([]ProjectSummary, error) {
	query := `
		SELECT id, app_name, summary, created_at::text
		FROM projects
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list projects query: %w", err)
	}
	defer rows.Close()

	var projects []ProjectSummary
	for rows.Next() {
		var p ProjectSummary
		if err := rows.Scan(&p.ID, &p.AppName, &p.Summary, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan project summary: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list projects rows error: %w", err)
	}

	return projects, nil
}

func (r *PostgresRepository) GetProjectPlan(ctx context.Context, projectID string) (*models.GeneratePlanResponse, error) {
	projectQuery := `
		SELECT id, app_name, summary, mvp_scope, assumptions, risks
		FROM projects
		WHERE id = $1
	`

	var (
		idRaw          string
		appNameRaw     string
		summaryRaw     string
		mvpScopeRaw    []byte
		assumptionsRaw []byte
		risksRaw       []byte
	)

	err := r.db.QueryRow(ctx, projectQuery, projectID).Scan(
		&idRaw,
		&appNameRaw,
		&summaryRaw,
		&mvpScopeRaw,
		&assumptionsRaw,
		&risksRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}

	mvpScope, err := unmarshalStringSlice(mvpScopeRaw)
	if err != nil {
		return nil, err
	}
	assumptions, err := unmarshalStringSlice(assumptionsRaw)
	if err != nil {
		return nil, err
	}
	risks, err := unmarshalStringSlice(risksRaw)
	if err != nil {
		return nil, err
	}

	plan := &models.GeneratePlanResponse{
		AppName:        appNameRaw,
		ProjectID:      idRaw,
		ProjectSummary: summaryRaw,
		MVPScope:       mvpScope,
		Assumptions:    assumptions,
		Risks:          risks,
		Epics:          []models.Epic{},
	}

	epicsByID := map[string]*models.Epic{}
	storiesByID := map[string]*models.Story{}

	epicsQuery := `
		SELECT id, project_id, title, description, priority,
		       acceptance_criteria, estimate_value, estimate_unit,
		       dependencies
		FROM epics
		WHERE project_id = $1
		ORDER BY order_index ASC
	`

	epicRows, err := r.db.Query(ctx, epicsQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("query epics: %w", err)
	}
	defer epicRows.Close()

	for epicRows.Next() {
		var (
			epic               models.Epic
			projectIDRaw       string
			acceptanceCriteria []byte
			dependencies       []byte
		)

		if err := epicRows.Scan(
			&epic.ID,
			&projectIDRaw,
			&epic.Title,
			&epic.Description,
			&epic.Priority,
			&acceptanceCriteria,
			&epic.Estimate.Value,
			&epic.Estimate.Unit,
			&dependencies,
		); err != nil {
			return nil, fmt.Errorf("scan epic: %w", err)
		}

		epic.ProjectID = projectIDRaw
		epic.Stories = []models.Story{}

		epic.AcceptanceCriteria, err = unmarshalStringSlice(acceptanceCriteria)
		if err != nil {
			return nil, err
		}
		epic.Dependencies, err = unmarshalStringSlice(dependencies)
		if err != nil {
			return nil, err
		}

		plan.Epics = append(plan.Epics, epic)
		epicsByID[epic.ID] = &plan.Epics[len(plan.Epics)-1]
	}

	if err := epicRows.Err(); err != nil {
		return nil, fmt.Errorf("epic rows error: %w", err)
	}

	storiesQuery := `
		SELECT id, epic_id, title, description, priority,
		       acceptance_criteria, estimate_value, estimate_unit,
		       dependencies
		FROM stories
		WHERE epic_id IN (
			SELECT id FROM epics WHERE project_id = $1
		)
		ORDER BY order_index ASC
	`

	storyRows, err := r.db.Query(ctx, storiesQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("query stories: %w", err)
	}
	defer storyRows.Close()

	for storyRows.Next() {
		var (
			story              models.Story
			epicIDRaw          string
			acceptanceCriteria []byte
			dependencies       []byte
		)

		if err := storyRows.Scan(
			&story.ID,
			&epicIDRaw,
			&story.Title,
			&story.Description,
			&story.Priority,
			&acceptanceCriteria,
			&story.Estimate.Value,
			&story.Estimate.Unit,
			&dependencies,
		); err != nil {
			return nil, fmt.Errorf("scan story: %w", err)
		}

		story.EpicID = epicIDRaw
		story.Tasks = []models.Task{}

		story.AcceptanceCriteria, err = unmarshalStringSlice(acceptanceCriteria)
		if err != nil {
			return nil, err
		}
		story.Dependencies, err = unmarshalStringSlice(dependencies)
		if err != nil {
			return nil, err
		}

		parentEpic, ok := epicsByID[epicIDRaw]
		if !ok {
			return nil, fmt.Errorf("parent epic not found for story %s", story.ID)
		}

		parentEpic.Stories = append(parentEpic.Stories, story)
		storiesByID[story.ID] = &parentEpic.Stories[len(parentEpic.Stories)-1]
	}

	if err := storyRows.Err(); err != nil {
		return nil, fmt.Errorf("story rows error: %w", err)
	}

	tasksQuery := `
		SELECT id, story_id, title, description, priority,
		       acceptance_criteria, estimate_value, estimate_unit,
		       dependencies
		FROM tasks
		WHERE story_id IN (
			SELECT s.id
			FROM stories s
			JOIN epics e ON s.epic_id = e.id
			WHERE e.project_id = $1
		)
		ORDER BY order_index ASC
	`

	taskRows, err := r.db.Query(ctx, tasksQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer taskRows.Close()

	for taskRows.Next() {
		var (
			task               models.Task
			storyIDRaw         string
			acceptanceCriteria []byte
			dependencies       []byte
		)

		if err := taskRows.Scan(
			&task.ID,
			&storyIDRaw,
			&task.Title,
			&task.Description,
			&task.Priority,
			&acceptanceCriteria,
			&task.Estimate.Value,
			&task.Estimate.Unit,
			&dependencies,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		task.StoryID = storyIDRaw

		task.AcceptanceCriteria, err = unmarshalStringSlice(acceptanceCriteria)
		if err != nil {
			return nil, err
		}
		task.Dependencies, err = unmarshalStringSlice(dependencies)
		if err != nil {
			return nil, err
		}

		parentStory, ok := storiesByID[storyIDRaw]
		if !ok {
			return nil, fmt.Errorf("parent story not found for task %s", task.ID)
		}

		parentStory.Tasks = append(parentStory.Tasks, task)
	}

	if err := taskRows.Err(); err != nil {
		return nil, fmt.Errorf("task rows error: %w", err)
	}

	return plan, nil
}
