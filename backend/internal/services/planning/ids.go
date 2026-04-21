package planning

import (
	"fmt"

	"github.com/motah-fard/ai-agent/backend/internal/models"
)

func AssignLocalIDs(plan *models.GeneratePlanResponse) {
	if plan.ProjectID == "" {
		plan.ProjectID = "project_1"
	}

	for i := range plan.Epics {
		if plan.Epics[i].ID == "" {
			plan.Epics[i].ID = fmt.Sprintf("epic_%d", i+1)
		}
		plan.Epics[i].ProjectID = plan.ProjectID

		for j := range plan.Epics[i].Stories {
			if plan.Epics[i].Stories[j].ID == "" {
				plan.Epics[i].Stories[j].ID = fmt.Sprintf("%s_story_%d", plan.Epics[i].ID, j+1)
			}
			plan.Epics[i].Stories[j].EpicID = plan.Epics[i].ID

			for k := range plan.Epics[i].Stories[j].Tasks {
				if plan.Epics[i].Stories[j].Tasks[k].ID == "" {
					plan.Epics[i].Stories[j].Tasks[k].ID = fmt.Sprintf("%s_task_%d", plan.Epics[i].Stories[j].ID, k+1)
				}
				plan.Epics[i].Stories[j].Tasks[k].StoryID = plan.Epics[i].Stories[j].ID
			}
		}
	}
}
