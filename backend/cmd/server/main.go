package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/motah-fard/ai-agent/backend/internal/api/handlers"
	"github.com/motah-fard/ai-agent/backend/internal/api/routes"
	"github.com/motah-fard/ai-agent/backend/internal/integrations/jira"
	"github.com/motah-fard/ai-agent/backend/internal/llm"
	planningservice "github.com/motah-fard/ai-agent/backend/internal/services/planning"
	"github.com/motah-fard/ai-agent/backend/internal/storage/postgres"
)

func main() {
	ctx := context.Background()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	dbCfg, err := postgres.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.NewDB(ctx, dbCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := postgres.NewRepository(db)

	jiraCfg, err := jira.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	jiraClient := jira.NewClient(jiraCfg)

	model := "gpt-4o-mini"

	client := llm.NewClient(apiKey, model)
	planner := llm.NewPlanner(client)
	planningService := planningservice.NewService(planner, repo, jiraClient)
	planningHandler := handlers.NewPlanningHandler(planningService)

	router := routes.NewRouter(planningHandler)

	addr := ":8080"
	log.Printf("server listening on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
