package planning

import (
	"fmt"
	"strings"

	"github.com/motah-fard/ai-agent/backend/internal/models"
)

type FlatProject struct {
	ID          string
	AppName     string
	Summary     string
	MVPScope    string
	Assumptions string
	Risks       string
}

type FlatEpic struct {
	ID                 string
	ProjectID          string
	Title              string
	Description        string
	Priority           string
	AcceptanceCriteria string
	EstimateValue      int
	EstimateUnit       string
	Dependencies       string
	OrderIndex         int
}

type FlatStory struct {
	ID                 string
	EpicID             string
	Title              string
	Description        string
	Priority           string
	AcceptanceCriteria string
	EstimateValue      int
	EstimateUnit       string
	Dependencies       string
	OrderIndex         int
}

type FlatTask struct {
	ID                 string
	StoryID            string
	Title              string
	Description        string
	Priority           string
	AcceptanceCriteria string
	EstimateValue      int
	EstimateUnit       string
	Dependencies       string
	OrderIndex         int
}

type FlattenedPlan struct {
	Project FlatProject
	Epics   []FlatEpic
	Stories []FlatStory
	Tasks   []FlatTask
}

func FlattenPlan(appName string, plan *models.GeneratePlanResponse) FlattenedPlan {
	projectID := chooseID(plan.ProjectID, "project_1")

	project := FlatProject{
		ID:          projectID,
		AppName:     appName,
		Summary:     plan.ProjectSummary,
		MVPScope:    strings.Join(plan.MVPScope, " || "),
		Assumptions: strings.Join(plan.Assumptions, " || "),
		Risks:       strings.Join(plan.Risks, " || "),
	}

	var flatEpics []FlatEpic
	var flatStories []FlatStory
	var flatTasks []FlatTask

	for i, epic := range plan.Epics {
		epicID := chooseID(epic.ID, fmt.Sprintf("epic_%d", i+1))

		flatEpics = append(flatEpics, FlatEpic{
			ID:                 epicID,
			ProjectID:          projectID,
			Title:              epic.Title,
			Description:        epic.Description,
			Priority:           epic.Priority,
			AcceptanceCriteria: strings.Join(epic.AcceptanceCriteria, " || "),
			EstimateValue:      epic.Estimate.Value,
			EstimateUnit:       epic.Estimate.Unit,
			Dependencies:       strings.Join(epic.Dependencies, " || "),
			OrderIndex:         i,
		})

		for j, story := range epic.Stories {
			storyID := chooseID(story.ID, fmt.Sprintf("%s_story_%d", epicID, j+1))

			flatStories = append(flatStories, FlatStory{
				ID:                 storyID,
				EpicID:             epicID,
				Title:              story.Title,
				Description:        story.Description,
				Priority:           story.Priority,
				AcceptanceCriteria: strings.Join(story.AcceptanceCriteria, " || "),
				EstimateValue:      story.Estimate.Value,
				EstimateUnit:       story.Estimate.Unit,
				Dependencies:       strings.Join(story.Dependencies, " || "),
				OrderIndex:         j,
			})

			for k, task := range story.Tasks {
				taskID := chooseID(task.ID, fmt.Sprintf("%s_task_%d", storyID, k+1))

				flatTasks = append(flatTasks, FlatTask{
					ID:                 taskID,
					StoryID:            storyID,
					Title:              task.Title,
					Description:        task.Description,
					Priority:           task.Priority,
					AcceptanceCriteria: strings.Join(task.AcceptanceCriteria, " || "),
					EstimateValue:      task.Estimate.Value,
					EstimateUnit:       task.Estimate.Unit,
					Dependencies:       strings.Join(task.Dependencies, " || "),
					OrderIndex:         k,
				})
			}
		}
	}

	return FlattenedPlan{
		Project: project,
		Epics:   flatEpics,
		Stories: flatStories,
		Tasks:   flatTasks,
	}
}

func chooseID(current string, fallback string) string {
	if strings.TrimSpace(current) != "" {
		return current
	}
	return fallback
}
