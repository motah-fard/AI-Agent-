package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/motah-fard/ai-agent/backend/internal/models"
)

func (r *PostgresRepository) SaveProjectPlan(ctx context.Context, appName string, plan *models.GeneratePlanResponse) error {
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}
	if plan.ProjectID == "" {
		return fmt.Errorf("plan.ProjectID is required")
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	mvpScopeJSON, err := marshalStringSlice(plan.MVPScope)
	if err != nil {
		return err
	}
	assumptionsJSON, err := marshalStringSlice(plan.Assumptions)
	if err != nil {
		return err
	}
	risksJSON, err := marshalStringSlice(plan.Risks)
	if err != nil {
		return err
	}

	projectQuery := `
		INSERT INTO projects (id, app_name, summary, mvp_scope, assumptions, risks)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			app_name = EXCLUDED.app_name,
			summary = EXCLUDED.summary,
			mvp_scope = EXCLUDED.mvp_scope,
			assumptions = EXCLUDED.assumptions,
			risks = EXCLUDED.risks,
			updated_at = NOW()
	`

	_, err = tx.Exec(
		ctx,
		projectQuery,
		plan.ProjectID,
		appName,
		plan.ProjectSummary,
		mvpScopeJSON,
		assumptionsJSON,
		risksJSON,
	)
	if err != nil {
		return fmt.Errorf("insert project: %w", err)
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM tasks
		WHERE story_id IN (
			SELECT s.id
			FROM stories s
			JOIN epics e ON s.epic_id = e.id
			WHERE e.project_id = $1
		)
	`, plan.ProjectID)
	if err != nil {
		return fmt.Errorf("delete existing tasks: %w", err)
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM stories
		WHERE epic_id IN (
			SELECT id FROM epics WHERE project_id = $1
		)
	`, plan.ProjectID)
	if err != nil {
		return fmt.Errorf("delete existing stories: %w", err)
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM epics
		WHERE project_id = $1
	`, plan.ProjectID)
	if err != nil {
		return fmt.Errorf("delete existing epics: %w", err)
	}

	for epicIndex, epic := range plan.Epics {
		epicAcceptanceJSON, err := marshalStringSlice(epic.AcceptanceCriteria)
		if err != nil {
			return err
		}
		epicDependenciesJSON, err := marshalStringSlice(epic.Dependencies)
		if err != nil {
			return err
		}

		epicQuery := `
			INSERT INTO epics (
				id, project_id, title, description, priority,
				acceptance_criteria, estimate_value, estimate_unit,
				dependencies, order_index
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`

		_, err = tx.Exec(
			ctx,
			epicQuery,
			epic.ID,
			plan.ProjectID,
			epic.Title,
			epic.Description,
			epic.Priority,
			epicAcceptanceJSON,
			epic.Estimate.Value,
			epic.Estimate.Unit,
			epicDependenciesJSON,
			epicIndex,
		)
		if err != nil {
			return fmt.Errorf("insert epic %s: %w", epic.ID, err)
		}

		for storyIndex, story := range epic.Stories {
			storyAcceptanceJSON, err := marshalStringSlice(story.AcceptanceCriteria)
			if err != nil {
				return err
			}
			storyDependenciesJSON, err := marshalStringSlice(story.Dependencies)
			if err != nil {
				return err
			}

			storyQuery := `
				INSERT INTO stories (
					id, epic_id, title, description, priority,
					acceptance_criteria, estimate_value, estimate_unit,
					dependencies, order_index
				)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			`

			_, err = tx.Exec(
				ctx,
				storyQuery,
				story.ID,
				epic.ID,
				story.Title,
				story.Description,
				story.Priority,
				storyAcceptanceJSON,
				story.Estimate.Value,
				story.Estimate.Unit,
				storyDependenciesJSON,
				storyIndex,
			)
			if err != nil {
				return fmt.Errorf("insert story %s: %w", story.ID, err)
			}

			for taskIndex, task := range story.Tasks {
				taskAcceptanceJSON, err := marshalStringSlice(task.AcceptanceCriteria)
				if err != nil {
					return err
				}
				taskDependenciesJSON, err := marshalStringSlice(task.Dependencies)
				if err != nil {
					return err
				}

				taskQuery := `
					INSERT INTO tasks (
						id, story_id, title, description, priority,
						acceptance_criteria, estimate_value, estimate_unit,
						dependencies, order_index
					)
					VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
				`

				_, err = tx.Exec(
					ctx,
					taskQuery,
					task.ID,
					story.ID,
					task.Title,
					task.Description,
					task.Priority,
					taskAcceptanceJSON,
					task.Estimate.Value,
					task.Estimate.Unit,
					taskDependenciesJSON,
					taskIndex,
				)
				if err != nil {
					return fmt.Errorf("insert task %s: %w", task.ID, err)
				}
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
