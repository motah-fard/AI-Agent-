package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/motah-fard/ai-agent/backend/internal/models"
	planningservice "github.com/motah-fard/ai-agent/backend/internal/services/planning"
)

type PlanningHandler struct {
	service *planningservice.Service
}

func NewPlanningHandler(service *planningservice.Service) *PlanningHandler {
	return &PlanningHandler{
		service: service,
	}
}

func (h *PlanningHandler) GeneratePlan(w http.ResponseWriter, r *http.Request) {
	log.Println("GeneratePlan handler hit")

	var req models.GeneratePlanRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("decode error: %v\n", err)
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	log.Printf("request: %+v\n", req)

	if strings.TrimSpace(req.Idea) == "" {
		writeJSONError(w, http.StatusBadRequest, "idea is required")
		return
	}

	if strings.TrimSpace(req.AppName) == "" {
		writeJSONError(w, http.StatusBadRequest, "app_name is required")
		return
	}

	resp, err := h.service.GeneratePlan(r.Context(), req)
	if err != nil {
		log.Printf("service error: %v\n", err)
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("response: %+v\n", resp)
	writeJSON(w, http.StatusOK, resp)
}

func (h *PlanningHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(r.Context())
	if err != nil {
		log.Printf("list projects error: %v\n", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}

	writeJSON(w, http.StatusOK, projects)
}

func (h *PlanningHandler) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if strings.TrimSpace(projectID) == "" {
		writeJSONError(w, http.StatusBadRequest, "project id is required")
		return
	}

	project, err := h.service.GetProjectPlan(r.Context(), projectID)
	if err != nil {
		log.Printf("get project error: %v\n", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to get project")
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func (h *PlanningHandler) RegenerateProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if strings.TrimSpace(projectID) == "" {
		writeJSONError(w, http.StatusBadRequest, "project id is required")
		return
	}

	project, err := h.service.RegenerateProject(r.Context(), projectID)
	if err != nil {
		log.Printf("regenerate project error: %v\n", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to regenerate project")
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func (h *PlanningHandler) RefineProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if strings.TrimSpace(projectID) == "" {
		writeJSONError(w, http.StatusBadRequest, "project id is required")
		return
	}

	var req models.RefineProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Instruction) == "" {
		writeJSONError(w, http.StatusBadRequest, "instruction is required")
		return
	}

	project, err := h.service.RefineProject(r.Context(), projectID, req.Instruction)
	if err != nil {
		log.Printf("refine project error: %v\n", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to refine project")
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func (h *PlanningHandler) PreviewJiraExport(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if strings.TrimSpace(projectID) == "" {
		writeJSONError(w, http.StatusBadRequest, "project id is required")
		return
	}

	var req models.JiraPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.ProjectKey) == "" {
		writeJSONError(w, http.StatusBadRequest, "project_key is required")
		return
	}

	preview, err := h.service.PreviewJiraExport(r.Context(), projectID, req.ProjectKey)
	if err != nil {
		log.Printf("jira preview error: %v\n", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to preview jira export")
		return
	}

	writeJSON(w, http.StatusOK, preview)
}

func (h *PlanningHandler) ExportToJira(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if strings.TrimSpace(projectID) == "" {
		writeJSONError(w, http.StatusBadRequest, "project id is required")
		return
	}

	var req models.JiraPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.ProjectKey) == "" {
		writeJSONError(w, http.StatusBadRequest, "project_key is required")
		return
	}

	exported, err := h.service.ExportToJira(r.Context(), projectID, req.ProjectKey)
	if err != nil {
		log.Printf("jira export error: %v\n", err)
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, exported)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("encode error: %v\n", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
