package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/motah-fard/ai-agent/backend/internal/models"
)

type ProjectSummary struct {
	ID        string `json:"id"`
	AppName   string `json:"app_name"`
	Summary   string `json:"summary"`
	CreatedAt string `json:"created_at"`
}

type ProjectWriter interface {
	SaveProjectPlan(ctx context.Context, appName string, plan *models.GeneratePlanResponse) error
}

type ProjectReader interface {
	GetProjectPlan(ctx context.Context, projectID string) (*models.GeneratePlanResponse, error)
	ListProjects(ctx context.Context) ([]ProjectSummary, error)
}

type Repository interface {
	ProjectWriter
	ProjectReader
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func marshalStringSlice(v []string) ([]byte, error) {
	if v == nil {
		v = []string{}
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal []string: %w", err)
	}
	return b, nil
}

func unmarshalStringSlice(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return []string{}, nil
	}

	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("unmarshal []string: %w", err)
	}
	return out, nil
}
