package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Taskwarrior utilities
func runTask(args ...string) (string, error) {
	// Global Overrides based on task.pdf and automation requirements
	// rc.confirmation=off: prevents hanging on destructive or recurring tasks
	// rc.verbose=nothing: ensures we get clean data/confirmation without headers
	// rc.hooks=on: ensures Timewarrior hooks trigger during start/stop
	baseArgs := []string{"rc.confirmation=off", "rc.verbose=nothing", "rc.hooks=on"}
	finalArgs := append(baseArgs, args...)

	cmd := exec.Command("task", finalArgs...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		return "", fmt.Errorf("task error: %v\nOutput: %s", err, output)
	}
	return output, nil
}

func main() {
	s := server.NewMCPServer(
		"TaskMaster-Pro-Go",
		"3.0.0",
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	// --- 1. CORE TASK TOOLS ---

	s.AddTool(mcp.NewTool("task_list",
		mcp.WithDescription("Export tasks as JSON. Use filters like 'project:Work' or '+PENDING'."),
		mcp.WithString("filter", mcp.Description("Filter string. Default: status:pending")),
	), listHandler)

	s.AddTool(mcp.NewTool("task_add",
		mcp.WithDescription("Create a new task with metadata."),
		mcp.WithString("description", mcp.Required(), mcp.Description("Task description")),
		mcp.WithString("metadata", mcp.Description("Attributes like 'project:Home due:2pm +next'")),
	), addHandler)

	s.AddTool(mcp.NewTool("task_modify",
		mcp.WithDescription("Update a task by UUID."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
		mcp.WithString("modifications", mcp.Required(), mcp.Description("Mods like 'project:New /old/new/ +tag'")),
	), modifyHandler)

	// --- 2. THE "MISSING" COMMANDS (Agentic Control) ---

	s.AddTool(mcp.NewTool("task_start_stop",
		mcp.WithDescription("Start or stop a task (triggers Timewarrior hooks)."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Either 'start' or 'stop'")),
	), startStopHandler)

	s.AddTool(mcp.NewTool("task_annotate",
		mcp.WithDescription("Add a note (annotation) to a task."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
		mcp.WithString("text", mcp.Required(), mcp.Description("The note to add")),
	), annotateHandler)

	s.AddTool(mcp.NewTool("task_log",
		mcp.WithDescription("Record a task that is already completed (useful for history gap filling)."),
		mcp.WithString("description", mcp.Required(), mcp.Description("What you did")),
		mcp.WithString("metadata", mcp.Description("Attributes like 'project:Work completed:yesterday'")),
	), logHandler)

	// --- 3. SYSTEM & BULK TOOLS ---

	s.AddTool(mcp.NewTool("task_calc",
		mcp.WithDescription("Evaluate Taskwarrior date math (e.g., 'now + 4d')."),
		mcp.WithString("expression", mcp.Required(), mcp.Description("Math string")),
	), calcHandler)

	s.AddTool(mcp.NewTool("task_import_bulk",
		mcp.WithDescription("Import JSON tasks for massive state updates."),
		mcp.WithString("json_data", mcp.Required(), mcp.Description("JSON array of tasks")),
	), importHandler)

	s.AddTool(mcp.NewTool("timew_context",
		mcp.WithDescription("Get raw Timewarrior data for scheduling analysis."),
		mcp.WithString("range", mcp.Description("Range like ':week'")),
	), timewHandler)

	// --- 4. SDK PROMPTS (The AI Strategy) ---

	s.AddPrompt(mcp.NewPrompt("daily_planner",
		mcp.WithPromptDescription("Prepares a summary of time spent and pending tasks for scheduling."),
		mcp.WithPromptArgument("focus", mcp.Description("Primary goal for the day"), mcp.RequiredPromptArgument()),
	), plannerPromptHandler)

	// --- SERVER START ---
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Fatal: %v\n", err)
	}
}

// --- HANDLERS ---

func listHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filter, _ := req.Params.Arguments["filter"].(string)
	if filter == "" {
		filter = "status:pending"
	}
	out, err := runTask(append(strings.Fields(filter), "export")...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func addHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	desc, _ := req.Params.RequireString("description")
	meta, _ := req.Params.Arguments["metadata"].(string)
	args := append([]string{"add", desc}, strings.Fields(meta)...)
	out, err := runTask(args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func modifyHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	uuid, _ := req.Params.RequireString("uuid")
	mods, _ := req.Params.RequireString("modifications")
	args := append([]string{uuid, "modify"}, strings.Fields(mods)...)
	out, err := runTask(args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func startStopHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	uuid, _ := req.Params.RequireString("uuid")
	action, _ := req.Params.RequireString("action")
	if action != "start" && action != "stop" {
		return mcp.NewToolResultError("Action must be 'start' or 'stop'"), nil
	}
	out, err := runTask(uuid, action)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func annotateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	uuid, _ := req.Params.RequireString("uuid")
	text, _ := req.Params.RequireString("text")
	out, err := runTask(uuid, "annotate", text)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func logHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	desc, _ := req.Params.RequireString("description")
	meta, _ := req.Params.Arguments["metadata"].(string)
	args := append([]string{"log", desc}, strings.Fields(meta)...)
	out, err := runTask(args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func calcHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	expr, _ := req.Params.RequireString("expression")
	out, err := runTask("calc", expr)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func importHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, _ := req.Params.RequireString("json_data")
	cmd := exec.Command("task", "rc.confirmation=off", "import")
	cmd.Stdin = strings.NewReader(data)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(string(out)), nil
	}
	return mcp.NewToolResultText("Import Successful"), nil
}

func timewHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	trange, _ := req.Params.Arguments["range"].(string)
	if trange == "" {
		trange = ":week"
	}
	out, err := exec.Command("timew", "export", trange).Output()
	if err != nil {
		return mcp.NewToolResultError("Timewarrior failed"), nil
	}
	return mcp.NewToolResultText(string(out)), nil
}

// Prompt Handler: Instructs the LLM on how to use the above tools for a "Daily Review"
func plannerPromptHandler(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	focus := req.Params.Arguments["focus"]
	
	return &mcp.GetPromptResult{
		Description: "Analyze current state and plan day",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.NewTextContent(fmt.Sprintf(
					"My focus today is: %s. Please do the following:\n"+
					"1. Call task_list to see pending tasks.\n"+
					"2. Call timew_context to see my recent activity.\n"+
					"3. Suggest which tasks I should 'start' now based on my focus.", 
					focus,
				)),
			},
		},
	}, nil
}
