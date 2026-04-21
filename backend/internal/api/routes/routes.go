package routes

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/motah-fard/ai-agent/backend/internal/api/handlers"
)

func NewRouter(planningHandler *handlers.PlanningHandler) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/plans/generate", planningHandler.GeneratePlan)
		r.Get("/projects", planningHandler.ListProjects)
		r.Get("/projects/{id}", planningHandler.GetProjectByID)
		r.Put("/projects/{id}/regenerate", planningHandler.RegenerateProject)
		r.Post("/projects/{id}/refine", planningHandler.RefineProject)
		r.Post("/projects/{id}/jira-preview", planningHandler.PreviewJiraExport)
		r.Post("/projects/{id}/jira-export", planningHandler.ExportToJira)
	})

	return r
}
