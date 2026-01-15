package taskwarrior

import (
	"context"
	"strings"
	"warmcp/pkg/common"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var baseArgs = []string{"rc.confirmation=off", "rc.verbose=nothing", "rc.hooks=on"}

func runTask(args ...string) (string, error) {
	return common.RunCommand("task", baseArgs, args...)
}

func RegisterHandlers(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("task_add",
		mcp.WithDescription("Create a new task. PROMPT FOR CONFIRMATION."),
		mcp.WithString("description", mcp.Required(), mcp.Description("Task description")),
		mcp.WithString("metadata", mcp.Description("Attributes like 'project:Home due:2pm +next'")),
	), addHandler)

	s.AddTool(mcp.NewTool("task_modify",
		mcp.WithDescription("Modify a task by UUID. PROMPT FOR CONFIRMATION."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
		mcp.WithString("modifications", mcp.Required(), mcp.Description("Modifications like 'project:New /old/new/ +tag'")),
	), modifyHandler)

	s.AddTool(mcp.NewTool("task_done",
		mcp.WithDescription("Mark a task as done. PROMPT FOR CONFIRMATION."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
	), doneHandler)

	s.AddTool(mcp.NewTool("task_delete",
		mcp.WithDescription("Delete a task. PROMPT FOR CONFIRMATION."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
	), deleteHandler)

	s.AddTool(mcp.NewTool("task_list",
		mcp.WithDescription("List tasks (export JSON). NO CONFIRMATION NEEDED."),
		mcp.WithString("filter", mcp.Description("Filter string. Default: status:pending")),
	), listHandler)

	s.AddTool(mcp.NewTool("task_annotate",
		mcp.WithDescription("Add annotation to a task. PROMPT FOR CONFIRMATION."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Annotation text")),
	), annotateHandler)

	s.AddTool(mcp.NewTool("task_denote",
		mcp.WithDescription("Remove annotation from a task. PROMPT FOR CONFIRMATION."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Annotation text to remove (substring match)")),
	), denoteHandler)

	s.AddTool(mcp.NewTool("task_start",
		mcp.WithDescription("Start a task. PROMPT FOR CONFIRMATION."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
	), startHandler)

	s.AddTool(mcp.NewTool("task_stop",
		mcp.WithDescription("Stop a task. PROMPT FOR CONFIRMATION."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
	), stopHandler)

	s.AddTool(mcp.NewTool("task_undo",
		mcp.WithDescription("Undo the last Taskwarrior operation. PROMPT FOR CONFIRMATION."),
	), undoHandler)

	s.AddTool(mcp.NewTool("task_calc",
		mcp.WithDescription("Evaluate Taskwarrior date math (e.g., 'now + 4d'). NO CONFIRMATION NEEDED."),
		mcp.WithString("expression", mcp.Required(), mcp.Description("Math expression")),
	), calcHandler)

	s.AddTool(mcp.NewTool("task_raw",
		mcp.WithDescription("Run raw task command. PROMPT FOR CONFIRMATION."),
		mcp.WithString("command", mcp.Required(), mcp.Description("Full task command arguments")),
	), rawHandler)
}

func listHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	filter, _ := argsMap["filter"].(string)
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
	argsMap, _ := req.Params.Arguments.(map[string]any)
	desc, _ := argsMap["description"].(string)
	meta, _ := argsMap["metadata"].(string)
	args := append([]string{"add", desc}, strings.Fields(meta)...)
	out, err := runTask(args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func modifyHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	mods, _ := argsMap["modifications"].(string)
	args := append([]string{uuid, "modify"}, strings.Fields(mods)...)
	out, err := runTask(args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func doneHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	out, err := runTask(uuid, "done")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func deleteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	out, err := runTask(uuid, "delete")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func annotateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	text, _ := argsMap["text"].(string)
	out, err := runTask(uuid, "annotate", text)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func denoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	text, _ := argsMap["text"].(string)
	out, err := runTask(uuid, "denote", text)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func startHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	out, err := runTask(uuid, "start")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func stopHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	out, err := runTask(uuid, "stop")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func undoHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	out, err := runTask("undo")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func calcHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	expr, _ := argsMap["expression"].(string)
	out, err := runTask("calc", expr)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func rawHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	cmd, _ := argsMap["command"].(string)
	out, err := runTask(strings.Fields(cmd)...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}
