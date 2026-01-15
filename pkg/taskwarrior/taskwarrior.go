package taskwarrior

import (
	"context"
	"fmt"
	"os"
	"strings"
	"warmcp/pkg/common"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var baseArgs = []string{"rc.confirmation=off", "rc.verbose=nothing", "rc.hooks=on"}

// TaskCommand represents a structured Taskwarrior command.
type TaskCommand struct {
	Overrides     []string
	Filters       []string
	Command       string
	Modifications []string
}

func (c *TaskCommand) Run() (string, error) {
	args := append([]string{}, c.Overrides...)
	args = append(args, c.Filters...)
	if c.Command != "" {
		args = append(args, c.Command)
	}
	args = append(args, c.Modifications...)

	env := []string{
		fmt.Sprintf("TASKRC=%s", common.GetTaskrcPath()),
	}
	return common.RunCommand("task", env, baseArgs, args...)
}

func RegisterHandlers(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("task_add",
		mcp.WithDescription("Create a new task. PROMPT FOR CONFIRMATION."),
		mcp.WithString("description", mcp.Required(), mcp.Description("Task description")),
		mcp.WithString("metadata", mcp.Description("Attributes like 'project:Home due:2pm +next'")),
	), addHandler)

	s.AddTool(mcp.NewTool("task_modify",
		mcp.WithDescription("Modify tasks. Can take filters and multiple modifications. PROMPT FOR CONFIRMATION."),
		mcp.WithString("filter", mcp.Description("Filter for tasks to modify (e.g., '+PENDING project:Work')")),
		mcp.WithString("modifications", mcp.Required(), mcp.Description("Modifications to apply (e.g., 'project:New /old/new/ +tag')")),
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
		mcp.WithDescription("Add annotation. PROMPT FOR CONFIRMATION."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Annotation text")),
	), annotateHandler)

	s.AddTool(mcp.NewTool("task_denote",
		mcp.WithDescription("Remove annotation. PROMPT FOR CONFIRMATION."),
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
		mcp.WithDescription("Evaluate Taskwarrior date math. NO CONFIRMATION NEEDED."),
		mcp.WithString("expression", mcp.Required(), mcp.Description("Math expression")),
	), calcHandler)

	s.AddTool(mcp.NewTool("task_raw",
		mcp.WithDescription("Run raw task command. PROMPT FOR CONFIRMATION."),
		mcp.WithString("command", mcp.Required(), mcp.Description("Full task command arguments")),
	), rawHandler)

	s.AddTool(mcp.NewTool("task_config",
		mcp.WithDescription("View or modify Taskwarrior configuration. PROMPT FOR CONFIRMATION for modifications."),
		mcp.WithString("name", mcp.Description("Config name to view or set")),
		mcp.WithString("value", mcp.Description("Value to set (if empty, views the config)")),
	), configHandler)

	s.AddTool(mcp.NewTool("task_purge",
		mcp.WithDescription("Permanently remove tasks from the database. PROMPT FOR CONFIRMATION."),
		mcp.WithString("filter", mcp.Required(), mcp.Description("Filter for tasks to purge")),
	), purgeHandler)

	s.AddTool(mcp.NewTool("task_append",
		mcp.WithDescription("Append text to a task's description. PROMPT FOR CONFIRMATION."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Text to append")),
	), appendHandler)

	s.AddTool(mcp.NewTool("task_prepend",
		mcp.WithDescription("Prepend text to a task's description. PROMPT FOR CONFIRMATION."),
		mcp.WithString("uuid", mcp.Required(), mcp.Description("UUID of the task")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Text to prepend")),
	), prependHandler)

	s.AddTool(mcp.NewTool("task_import",
		mcp.WithDescription("Import tasks from JSON format. PROMPT FOR CONFIRMATION."),
		mcp.WithString("json_data", mcp.Required(), mcp.Description("JSON string of tasks to import")),
	), importHandler)

	s.AddTool(mcp.NewTool("task_tags",
		mcp.WithDescription("List all unique tags. NO CONFIRMATION NEEDED."),
	), tagsHandler)

	s.AddTool(mcp.NewTool("task_projects",
		mcp.WithDescription("List all unique projects. NO CONFIRMATION NEEDED."),
	), projectsHandler)

	s.AddTool(mcp.NewTool("task_udas",
		mcp.WithDescription("List all User Defined Attributes. NO CONFIRMATION NEEDED."),
	), udasHandler)

	s.AddTool(mcp.NewTool("task_diagnostics",
		mcp.WithDescription("Show Taskwarrior diagnostic information (config, version, environment). NO CONFIRMATION NEEDED."),
	), diagnosticsHandler)

	s.AddTool(mcp.NewTool("task_stats",
		mcp.WithDescription("Show database statistics. NO CONFIRMATION NEEDED."),
	), statsHandler)
}

func listHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	filter, _ := argsMap["filter"].(string)
	if filter == "" {
		filter = "status:pending"
	}
	cmd := &TaskCommand{
		Filters: strings.Fields(filter),
		Command: "export",
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func addHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	desc, _ := argsMap["description"].(string)
	meta, _ := argsMap["metadata"].(string)

	cmd := &TaskCommand{
		Command:       "add",
		Modifications: append([]string{desc}, strings.Fields(meta)...),
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func modifyHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	filter, _ := argsMap["filter"].(string)
	mods, _ := argsMap["modifications"].(string)

	cmd := &TaskCommand{
		Filters:       strings.Fields(filter),
		Command:       "modify",
		Modifications: strings.Fields(mods),
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func doneHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	cmd := &TaskCommand{
		Filters: []string{uuid},
		Command: "done",
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func deleteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	cmd := &TaskCommand{
		Filters: []string{uuid},
		Command: "delete",
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func annotateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	text, _ := argsMap["text"].(string)
	cmd := &TaskCommand{
		Filters:       []string{uuid},
		Command:       "annotate",
		Modifications: []string{text},
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func denoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	text, _ := argsMap["text"].(string)
	cmd := &TaskCommand{
		Filters:       []string{uuid},
		Command:       "denote",
		Modifications: []string{text},
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func startHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	cmd := &TaskCommand{
		Filters: []string{uuid},
		Command: "start",
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func stopHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	cmd := &TaskCommand{
		Filters: []string{uuid},
		Command: "stop",
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func undoHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmd := &TaskCommand{
		Command: "undo",
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func calcHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	expr, _ := argsMap["expression"].(string)
	cmd := &TaskCommand{
		Command:       "calc",
		Modifications: []string{expr},
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func rawHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	cmdStr, _ := argsMap["command"].(string)
	fields := strings.Fields(cmdStr)

	cmd := &TaskCommand{}
	if len(fields) > 0 {
		cmd.Command = fields[0]
		cmd.Modifications = fields[1:]
	}

	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func configHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	name, _ := argsMap["name"].(string)
	val, ok := argsMap["value"].(string)

	cmd := &TaskCommand{Command: "config"}
	if name != "" {
		cmd.Modifications = []string{name}
		if ok && val != "" {
			cmd.Modifications = append(cmd.Modifications, val)
		}
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func purgeHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	filter, _ := argsMap["filter"].(string)
	cmd := &TaskCommand{
		Filters: strings.Fields(filter),
		Command: "purge",
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func appendHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	text, _ := argsMap["text"].(string)
	cmd := &TaskCommand{
		Filters:       []string{uuid},
		Command:       "append",
		Modifications: []string{text},
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func prependHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	uuid, _ := argsMap["uuid"].(string)
	text, _ := argsMap["text"].(string)
	cmd := &TaskCommand{
		Filters:       []string{uuid},
		Command:       "prepend",
		Modifications: []string{text},
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func importHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	data, _ := argsMap["json_data"].(string)

	tmpFile, err := os.CreateTemp("", "task_import_*.json")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.WriteString(data); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	tmpFile.Close()

	cmd := &TaskCommand{
		Command:       "import",
		Modifications: []string{tmpFile.Name()},
	}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func tagsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmd := &TaskCommand{Command: "tags"}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func projectsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmd := &TaskCommand{Command: "projects"}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func udasHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmd := &TaskCommand{Command: "udas"}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func diagnosticsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmd := &TaskCommand{Command: "diagnostics"}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func statsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmd := &TaskCommand{Command: "stats"}
	out, err := cmd.Run()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}
