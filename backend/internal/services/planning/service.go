package planning

import (
	"context"

	"github.com/motah-fard/ai-agent/backend/internal/integrations/jira"
	"github.com/motah-fard/ai-agent/backend/internal/llm"
	"github.com/motah-fard/ai-agent/backend/internal/models"
	"github.com/motah-fard/ai-agent/backend/internal/storage/postgres"
)

type Service struct {
	planner    *llm.Planner
	repo       postgres.Repository
	jiraClient *jira.Client
}

func NewService(planner *llm.Planner, repo postgres.Repository, jiraClient *jira.Client) *Service {
	return &Service{
		planner:    planner,
		repo:       repo,
		jiraClient: jiraClient,
	}
}

func (s *Service) GeneratePlan(ctx context.Context, req models.GeneratePlanRequest) (*models.GeneratePlanResponse, error) {
	plan, err := s.planner.GeneratePlan(ctx, req)
	if err != nil {
		return nil, err
	}

	AssignLocalIDs(plan)

	if plan.AppName == "" {
		plan.AppName = req.AppName
	}

	if err := s.repo.SaveProjectPlan(ctx, plan.AppName, plan); err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *Service) ListProjects(ctx context.Context) ([]postgres.ProjectSummary, error) {
	return s.repo.ListProjects(ctx)
}

func (s *Service) GetProjectPlan(ctx context.Context, projectID string) (*models.GeneratePlanResponse, error) {
	return s.repo.GetProjectPlan(ctx, projectID)
}

func (s *Service) RegenerateProject(ctx context.Context, projectID string) (*models.GeneratePlanResponse, error) {
	existing, err := s.repo.GetProjectPlan(ctx, projectID)
	if err != nil {
		return nil, err
	}

	regenerated, err := s.planner.RegeneratePlan(ctx, existing)
	if err != nil {
		return nil, err
	}

	regenerated.ProjectID = projectID
	if regenerated.AppName == "" {
		regenerated.AppName = existing.AppName
	}

	AssignLocalIDs(regenerated)

	if err := s.repo.SaveProjectPlan(ctx, regenerated.AppName, regenerated); err != nil {
		return nil, err
	}

	return regenerated, nil
}

func (s *Service) RefineProject(ctx context.Context, projectID string, instruction string) (*models.GeneratePlanResponse, error) {
	existing, err := s.repo.GetProjectPlan(ctx, projectID)
	if err != nil {
		return nil, err
	}

	refined, err := s.planner.RefinePlan(ctx, existing, instruction)
	if err != nil {
		return nil, err
	}

	refined.ProjectID = projectID
	if refined.AppName == "" {
		refined.AppName = existing.AppName
	}

	AssignLocalIDs(refined)

	if err := s.repo.SaveProjectPlan(ctx, refined.AppName, refined); err != nil {
		return nil, err
	}

	return refined, nil
}

func (s *Service) PreviewJiraExport(ctx context.Context, projectID string, jiraProjectKey string) (*jira.PreviewResponse, error) {
	plan, err := s.repo.GetProjectPlan(ctx, projectID)
	if err != nil {
		return nil, err
	}

	preview := jira.BuildPreview(plan, jiraProjectKey)
	return preview, nil
}

func (s *Service) ExportToJira(ctx context.Context, projectID string, jiraProjectKey string) (*jira.ExportResponse, error) {
	plan, err := s.repo.GetProjectPlan(ctx, projectID)
	if err != nil {
		return nil, err
	}

	preview := jira.BuildPreview(plan, jiraProjectKey)
	return s.jiraClient.CreateIssuesFromPreview(ctx, preview)
}
