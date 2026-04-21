package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/motah-fard/ai-agent/backend/internal/models"
)

func BuildPlanningPrompt(req models.GeneratePlanRequest) string {
	return fmt.Sprintf(`
You are a senior product planning assistant.

Given the app idea below, generate a practical MVP delivery plan.

Return valid JSON only.
Do not include markdown.
Do not include explanations outside the JSON.

Return JSON with this exact shape:
{
  "project_summary": "string",
  "mvp_scope": ["string"],
  "assumptions": ["string"],
  "risks": ["string"],
  "epics": [
    {
      "title": "string",
      "description": "string",
      "priority": "high|medium|low",
      "acceptance_criteria": ["string"],
      "estimate": {
        "value": 0,
        "unit": "hours|days|points"
      },
      "dependencies": ["string"],
      "stories": [
        {
          "title": "string",
          "description": "string",
          "priority": "high|medium|low",
          "acceptance_criteria": ["string"],
          "estimate": {
            "value": 0,
            "unit": "hours|days|points"
          },
          "dependencies": ["string"],
          "tasks": [
            {
              "title": "string",
              "description": "string",
              "priority": "high|medium|low",
              "acceptance_criteria": ["string"],
              "estimate": {
                "value": 0,
                "unit": "hours|days|points"
              },
              "dependencies": ["string"]
            }
          ]
        }
      ]
    }
  ]
}

Rules:
- Focus only on MVP
- Create exactly 3 to 5 epics
- Each epic must contain 2 to 4 stories
- Each story must contain 2 to 5 implementation tasks
- Keep the scope realistic for a first release
- Descriptions must be practical and specific
- Acceptance criteria must be testable
- Estimates must be rough but realistic
- Dependencies must reference other item titles by exact title
- Avoid vague items like "do backend", "build UI", or "set up everything"
- Prefer implementation-ready work items
- Do not leave arrays empty unless truly necessary

App name: %s
Idea: %s
Target users: %s
Constraints: %s
`,
		req.AppName,
		req.Idea,
		strings.Join(req.TargetUsers, ", "),
		strings.Join(req.Constraints, ", "),
	)
}
func BuildRegeneratePrompt(existing *models.GeneratePlanResponse) string {
	return fmt.Sprintf(`
You are a senior product planning assistant.

Regenerate the following project plan from scratch, but keep it aligned with the existing product intent.

Return valid JSON only.
Do not include markdown.
Do not include explanations outside the JSON.

Return JSON with this exact shape:
{
  "app_name": "string",
  "project_summary": "string",
  "mvp_scope": ["string"],
  "assumptions": ["string"],
  "risks": ["string"],
  "epics": [
    {
      "title": "string",
      "description": "string",
      "priority": "high|medium|low",
      "acceptance_criteria": ["string"],
      "estimate": {
        "value": 0,
        "unit": "hours|days|points"
      },
      "dependencies": ["string"],
      "stories": [
        {
          "title": "string",
          "description": "string",
          "priority": "high|medium|low",
          "acceptance_criteria": ["string"],
          "estimate": {
            "value": 0,
            "unit": "hours|days|points"
          },
          "dependencies": ["string"],
          "tasks": [
            {
              "title": "string",
              "description": "string",
              "priority": "high|medium|low",
              "acceptance_criteria": ["string"],
              "estimate": {
                "value": 0,
                "unit": "hours|days|points"
              },
              "dependencies": ["string"]
            }
          ]
        }
      ]
    }
  ]
}

Rules:
- Focus only on MVP
- Create exactly 3 to 5 epics
- Each epic must contain 2 to 4 stories
- Each story must contain 2 to 5 implementation tasks
- Keep the scope realistic for a first release
- Preserve the product intent, but improve clarity and structure

Existing project plan:
%s
`, mustJSON(existing))
}

func BuildRefinePrompt(existing *models.GeneratePlanResponse, instruction string) string {
	return fmt.Sprintf(`
You are a senior product planning assistant.

Refine the existing project plan according to the user's instruction.

Return valid JSON only.
Do not include markdown.
Do not include explanations outside the JSON.

Return JSON with this exact shape:
{
  "app_name": "string",
  "project_summary": "string",
  "mvp_scope": ["string"],
  "assumptions": ["string"],
  "risks": ["string"],
  "epics": [
    {
      "title": "string",
      "description": "string",
      "priority": "high|medium|low",
      "acceptance_criteria": ["string"],
      "estimate": {
        "value": 0,
        "unit": "hours|days|points"
      },
      "dependencies": ["string"],
      "stories": [
        {
          "title": "string",
          "description": "string",
          "priority": "high|medium|low",
          "acceptance_criteria": ["string"],
          "estimate": {
            "value": 0,
            "unit": "hours|days|points"
          },
          "dependencies": ["string"],
          "tasks": [
            {
              "title": "string",
              "description": "string",
              "priority": "high|medium|low",
              "acceptance_criteria": ["string"],
              "estimate": {
                "value": 0,
                "unit": "hours|days|points"
              },
              "dependencies": ["string"]
            }
          ]
        }
      ]
    }
  ]
}

Rules:
- Focus only on MVP
- Apply the refinement instruction carefully
- Keep useful existing work unless the instruction implies changing it
- Keep the hierarchy practical and implementation-ready
- Keep dependencies meaningful and title-based
- Maintain realistic estimates

User instruction:
%s

Existing project plan:
%s
`, instruction, mustJSON(existing))
}
func mustJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
