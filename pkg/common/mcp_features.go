package common

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMCPFeatures(s *server.MCPServer) {
	// --- PROMPTS ---
	s.AddPrompt(mcp.NewPrompt("daily_planner",
		mcp.WithPromptDescription("Prepares a summary of pending tasks and time spent for review."),
	), dailyPlannerPromptHandler)

	s.AddPrompt(mcp.NewPrompt("setup_check",
		mcp.WithPromptDescription("Check if Taskwarrior and Timewarrior are correctly installed and configured."),
	), setupCheckPromptHandler)

	// --- RESOURCES ---
	s.AddResource(mcp.Resource{
		URI:         "task://config",
		Name:        "Taskwarrior Configuration",
		Description: "The content of the .taskrc file",
		MIMEType:    "text/plain",
	}, taskConfigResourceHandler)

	s.AddResource(mcp.Resource{
		URI:         "task://summary",
		Name:        "Taskwarrior Summary",
		Description: "A high-level summary of active projects and pending tasks",
		MIMEType:    "text/plain",
	}, taskSummaryResourceHandler)

	s.AddResource(mcp.Resource{
		URI:         "task://tags",
		Name:        "Taskwarrior Tags",
		Description: "List of all unique tags used in Taskwarrior",
		MIMEType:    "text/plain",
	}, taskTagsResourceHandler)

	s.AddResource(mcp.Resource{
		URI:         "task://projects",
		Name:        "Taskwarrior Projects",
		Description: "List of all unique projects in Taskwarrior",
		MIMEType:    "text/plain",
	}, taskProjectsResourceHandler)

	s.AddResource(mcp.Resource{
		URI:         "task://udas",
		Name:        "Taskwarrior UDAs",
		Description: "List of all User Defined Attributes configured",
		MIMEType:    "text/plain",
	}, taskUDAsResourceHandler)

	s.AddResource(mcp.Resource{
		URI:         "task://diagnostics",
		Name:        "Taskwarrior Diagnostics",
		Description: "Taskwarrior diagnostic information (config, version, environment)",
		MIMEType:    "text/plain",
	}, taskDiagnosticsResourceHandler)
}

func dailyPlannerPromptHandler(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Review tasks and plan",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.NewTextContent(
					"Please review my pending tasks using `task_list` and my recent activity using `timew_summary`. " +
						"Then, suggest a plan for today and ask me to confirm which tasks I should start.",
				),
			},
		},
	}, nil
}

func setupCheckPromptHandler(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Check environment",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.NewTextContent(
					"Please run `task_raw command:--version` and `timew_raw command:--version` to verify the installation.",
				),
			},
		},
	}, nil
}

func taskConfigResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	path := GetTaskrcPath()
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read taskrc at %s: %v", path, err)
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "text/plain",
			Text:     string(content),
		},
	}, nil
}

func taskSummaryResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Simple summary via CLI
	out, err := RunCommand("task", nil, []string{"rc.verbose=nothing", "rc.confirmation=off"}, "summary")
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "text/plain",
			Text:     out,
		},
	}, nil
}

func taskTagsResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	out, err := RunCommand("task", nil, []string{"rc.verbose=nothing", "rc.confirmation=off"}, "tags")
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "text/plain",
			Text:     out,
		},
	}, nil
}

func taskProjectsResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	out, err := RunCommand("task", nil, []string{"rc.verbose=nothing", "rc.confirmation=off"}, "projects")
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "text/plain",
			Text:     out,
		},
	}, nil
}

func taskUDAsResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	out, err := RunCommand("task", nil, []string{"rc.verbose=nothing", "rc.confirmation=off"}, "udas")
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "text/plain",
			Text:     out,
		},
	}, nil
}

func taskDiagnosticsResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	out, err := RunCommand("task", nil, []string{"rc.verbose=nothing", "rc.confirmation=off"}, "diagnostics")
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "text/plain",
			Text:     out,
		},
	}, nil
}
