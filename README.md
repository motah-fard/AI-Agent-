# AI PM Agent

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](#tech-stack)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-v0.1.0--alpha-blue)](#roadmap)
[![API](https://img.shields.io/badge/API-REST-6DB33F)](#api-overview)
[![Jira](https://img.shields.io/badge/Jira-Cloud-0052CC?logo=jira)](#jira-integration)

A Go-based AI backend that turns an application idea into a structured delivery plan, stores it in Postgres, lets you refine or regenerate it, and exports the resulting epics, stories, and tasks into Jira.

This project is built for founder-style product planning and early execution workflows, where a rough idea needs to become something operational instead of living forever in notes, chat threads, and vague ambition.

---

## What this project does

Given an app idea, the service can:

- generate a structured MVP plan using an LLM
- break work into **epics, stories, and tasks**
- attach:
  - descriptions
  - priorities
  - acceptance criteria
  - estimates
  - dependencies
- persist the plan in Postgres
- fetch previously saved plans
- regenerate a plan from saved state
- refine an existing plan with a follow-up instruction
- preview how that plan would map into Jira
- export the mapped issues into a Jira Cloud project

---

## Current feature set

### AI planning
- `POST /api/v1/plans/generate`
- Input: app name, idea, target users, constraints
- Output: nested plan with epics, stories, tasks

### Retrieval
- `GET /api/v1/projects`
- `GET /api/v1/projects/{id}`

### Plan evolution
- `PUT /api/v1/projects/{id}/regenerate`
- `POST /api/v1/projects/{id}/refine`

### Jira integration
- `POST /api/v1/projects/{id}/jira-preview`
- `POST /api/v1/projects/{id}/jira-export`

### Persistence
- Postgres-backed storage for:
  - projects
  - epics
  - stories
  - tasks

---

## Example flow

1. Send an app idea to the generate endpoint
2. Save the returned plan in Postgres
3. Read the stored plan later
4. Refine or regenerate it
5. Preview its Jira mapping
6. Export it to Jira Cloud

That gives you a working planning pipeline from raw idea to actionable backlog. Civilization advances one endpoint at a time.

---

## Architecture overview

```text
Client / UI / curl
       |
       v
HTTP Handlers (chi)
       |
       v
Planning Service
       |
       +--> LLM Planner
       |      |
       |      +--> Prompt builders
       |      +--> OpenAI client
       |
       +--> Postgres Repository
       |
       +--> Jira Integration Layer
              |
              +--> Preview mapper
              +--> Jira Cloud client
```

### Main backend layers

- `internal/api/handlers`
  - HTTP request/response layer
- `internal/api/routes`
  - route registration
- `internal/services/planning`
  - orchestration and business logic
- `internal/llm`
  - prompt builders, planner logic, OpenAI API client
- `internal/storage/postgres`
  - database access and repository logic
- `internal/integrations/jira`
  - Jira preview mapping and export client
- `internal/models`
  - request/response and domain models
- `migrations`
  - SQL schema setup

---

## Tech stack

- **Go**
- **Chi** for routing
- **Postgres**
- **pgx / pgxpool**
- **OpenAI API**
- **Jira Cloud REST API**
- JSON-based REST endpoints

---

## Repository structure

```text
backend/
  cmd/
    server/
      main.go
  internal/
    api/
      handlers/
      routes/
    integrations/
      jira/
    llm/
    models/
    services/
      planning/
    storage/
      postgres/
  migrations/
  .env.example
  .gitignore
  go.mod
  README.md
  LICENSE
```

---

## Setup

### Prerequisites

- Go 1.22+
- Postgres running locally
- OpenAI API key
- Jira Cloud account and API token if you want Jira export

---

## Environment variables

Create a `.env` file locally or export variables in your shell.

### Required for core planning

```bash
OPENAI_API_KEY=your_openai_api_key
DATABASE_URL=postgres://localhost:5432/ai_agent?sslmode=disable
```

### Required for Jira export

```bash
JIRA_BASE_URL=https://your-domain.atlassian.net
JIRA_EMAIL=you@example.com
JIRA_API_TOKEN=your_jira_api_token
JIRA_PROJECT_KEY=KAN
```

### Optional for richer Jira hierarchy linking

```bash
JIRA_EPIC_LINK_FIELD_ID=customfield_10014
```

---

## Local development

### 1. Clone the repository

```bash
git clone <your-repo-url>
cd backend
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Create the database

```bash
createdb ai_agent
```

### 4. Run migrations

```bash
psql "$DATABASE_URL" -f migrations/0001_init.sql
```

### 5. Start the server

```bash
go run ./cmd/server
```

By default, the server listens on:

```text
:8080
```

---

## API overview

### Health check

```http
GET /health
```

### Generate plan

```http
POST /api/v1/plans/generate
```

Example request:

```json
{
  "app_name": "StockPilot",
  "idea": "An inventory management app for small fashion brands",
  "target_users": ["small fashion brand owners", "operations managers"],
  "constraints": ["MVP in 4 weeks", "web app only"]
}
```

### List projects

```http
GET /api/v1/projects
```

### Get one project

```http
GET /api/v1/projects/{id}
```

### Regenerate a project

```http
PUT /api/v1/projects/{id}/regenerate
```

### Refine a project

```http
POST /api/v1/projects/{id}/refine
```

Example request:

```json
{
  "instruction": "Split stories into clearer backend and frontend tasks and tighten acceptance criteria."
}
```

### Preview Jira export

```http
POST /api/v1/projects/{id}/jira-preview
```

Example request:

```json
{
  "project_key": "KAN"
}
```

### Export to Jira

```http
POST /api/v1/projects/{id}/jira-export
```

Example request:

```json
{
  "project_key": "KAN"
}
```

---

## Example curl commands

### Generate

```bash
curl -X POST http://localhost:8080/api/v1/plans/generate \
  -H "Content-Type: application/json" \
  -d '{
    "app_name": "StockPilot",
    "idea": "An inventory management app for small fashion brands",
    "target_users": ["small fashion brand owners", "operations managers"],
    "constraints": ["MVP in 4 weeks", "web app only"]
  }'
```

### List projects

```bash
curl http://localhost:8080/api/v1/projects
```

### Get one project

```bash
curl http://localhost:8080/api/v1/projects/project_1
```

### Refine

```bash
curl -X POST http://localhost:8080/api/v1/projects/project_1/refine \
  -H "Content-Type: application/json" \
  -d '{
    "instruction": "Separate backend and frontend work more clearly and improve estimates."
  }'
```

### Jira preview

```bash
curl -X POST http://localhost:8080/api/v1/projects/project_1/jira-preview \
  -H "Content-Type: application/json" \
  -d '{
    "project_key": "KAN"
  }'
```

### Jira export

```bash
curl -X POST http://localhost:8080/api/v1/projects/project_1/jira-export \
  -H "Content-Type: application/json" \
  -d '{
    "project_key": "KAN"
  }'
```

---

## Data model summary

A generated plan is stored as:

- **Project**
  - app name
  - summary
  - MVP scope
  - assumptions
  - risks
- **Epics**
- **Stories**
- **Tasks**

Each level can include:

- title
- description
- priority
- acceptance criteria
- estimate
- dependencies

Local IDs are assigned to preserve hierarchy and make persistence/export easier.

---

## Jira integration notes

Current Jira flow supports:

- previewing the mapped Jira payloads
- exporting issues into a Jira Cloud project
- mapping:
  - local epics -> Jira Epics
  - local stories -> Jira Stories
  - local tasks -> Jira Tasks

Why tasks are exported as `Task` instead of `Sub-task` in the current version:
- Jira project configurations vary
- some projects reject `Sub-task` as an allowed issue type
- `Task` is a safer first-pass mapping for broader compatibility

A future version can support:
- configurable issue-type mapping
- real sub-task export when available
- issue linking and sync-back
- Jira key persistence in the database

---

## Current limitations

This is still an early version. A few notable gaps remain:

- no authentication for the backend itself
- no frontend dashboard yet
- Jira issue keys are not yet persisted locally
- duplicate Jira export prevention is not yet implemented
- no update-in-place sync with Jira issues yet
- LLM output quality still depends on prompt discipline and model behavior

In other words, it works, but it still has room to become less trusting and more professional.

---

## Security and secrets

### Where your secrets usually live

Your secrets are most likely in one of these places:

1. **shell environment variables**
   - `echo $OPENAI_API_KEY`
   - `echo $JIRA_BASE_URL`
   - `echo $JIRA_EMAIL`
   - `echo $DATABASE_URL`

2. **a local `.env` file**
   - if you created one for development

3. **shell profile files**
   - `~/.zshrc`
   - `~/.bashrc`
   - `~/.bash_profile`

4. **hardcoded values in code**
   - bad idea, but worth searching for anyway

5. **terminal history**
   - especially if you pasted tokens directly into `export ...` commands

### What should never be committed

Do **not** commit:

- `.env`
- real API tokens
- real database credentials
- screenshots that expose tokens
- copied curl commands with live secrets
- shell config files with personal credentials

### Quick secret check before pushing

Run these before publishing:

```bash
git status
git diff
git diff --cached
```

And also search for suspicious strings:

```bash
grep -R "sk-" .
grep -R "JIRA_API_TOKEN" .
grep -R "OPENAI_API_KEY" .
grep -R "atlassian.net" .
```

Also check whether `.env` is ignored.

### Recommended `.gitignore`

```gitignore
.env
.env.*
.DS_Store
bin/
coverage.out
```

### Recommended `.env.example`

```bash
OPENAI_API_KEY=your_openai_api_key
DATABASE_URL=postgres://localhost:5432/ai_agent?sslmode=disable

JIRA_BASE_URL=https://your-domain.atlassian.net
JIRA_EMAIL=you@example.com
JIRA_API_TOKEN=your_jira_api_token
JIRA_PROJECT_KEY=KAN
JIRA_EPIC_LINK_FIELD_ID=customfield_10014
```

---

## Versioning

This project currently fits a **v0.1.0 alpha** label.

Suggested progression:

- `v0.1.0`
  - AI planning backend
  - Postgres persistence
  - retrieval
  - refine/regenerate
  - Jira preview/export

- `v0.2.0`
  - Jira mapping persistence
  - duplicate export protection
  - sync status endpoints

- `v0.3.0`
  - frontend UI
  - safer Jira update flow
  - improved issue-type configuration

- `v1.0.0`
  - stable setup
  - export safety
  - sync persistence
  - documentation cleanup
  - production-friendly workflows

---

## Roadmap

### Near-term
- persist Jira issue mappings
- prevent duplicate Jira exports
- add Jira status endpoint
- improve issue hierarchy syncing

### Mid-term
- add frontend dashboard
- add GitHub integration
- add better plan editing flows
- add more project-level metadata

### Longer-term
- Jira update-in-place sync
- richer issue-link support
- team workflows
- approval layer before export
- multi-project management UI

---

## Why this project exists

A lot of project planning still happens in:
- notes
- docs
- chats
- half-finished spreadsheets
- vague optimism

This project tries to turn that into:
- structured planning
- persistent state
- iterative refinement
- operational execution in Jira

Which is, frankly, a better use of software than yet another AI toy with a shiny landing page and no backbone.

---

## Contributing

This repository is still early-stage, but contributions, suggestions, and improvements are welcome.

Good areas for contribution:
- Jira sync reliability
- better issue-type configuration
- API validation improvements
- testing
- frontend/dashboard work
- plan editing flows

---

## License

This project is licensed under the **MIT License**. See the [`LICENSE`](LICENSE) file for details.


