package timewarrior

import (
	"context"
	"strings"
	"warmcp/pkg/common"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func runTimew(args ...string) (string, error) {
	return common.RunCommand("timew", nil, args...)
}

func RegisterHandlers(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("timew_start",
		mcp.WithDescription("Start tracking time. PROMPT FOR CONFIRMATION."),
		mcp.WithString("tags", mcp.Description("Tags for the time entry")),
	), startHandler)

	s.AddTool(mcp.NewTool("timew_stop",
		mcp.WithDescription("Stop tracking time. PROMPT FOR CONFIRMATION."),
		mcp.WithString("tags", mcp.Description("Optional tags for the entry being stopped")),
	), stopHandler)

	s.AddTool(mcp.NewTool("timew_continue",
		mcp.WithDescription("Continue tracking the most recent activity. PROMPT FOR CONFIRMATION."),
	), continueHandler)

	s.AddTool(mcp.NewTool("timew_summary",
		mcp.WithDescription("Get time tracking summary. NO CONFIRMATION NEEDED."),
		mcp.WithString("range", mcp.Description("Time range like ':week', ':day'")),
	), summaryHandler)

	s.AddTool(mcp.NewTool("timew_export",
		mcp.WithDescription("Export time data as JSON. NO CONFIRMATION NEEDED."),
		mcp.WithString("range", mcp.Description("Time range like ':week', ':day'")),
	), exportHandler)

	s.AddTool(mcp.NewTool("timew_raw",
		mcp.WithDescription("Run raw timew command. PROMPT FOR CONFIRMATION."),
		mcp.WithString("command", mcp.Required(), mcp.Description("Full timew command arguments")),
	), rawHandler)
}

func startHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	tags, _ := argsMap["tags"].(string)
	args := append([]string{"start"}, strings.Fields(tags)...)
	out, err := runTimew(args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func stopHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	tags, _ := argsMap["tags"].(string)
	args := append([]string{"stop"}, strings.Fields(tags)...)
	out, err := runTimew(args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func continueHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	out, err := runTimew("continue")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func summaryHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	trange, _ := argsMap["range"].(string)
	args := []string{"summary"}
	if trange != "" {
		args = append(args, trange)
	}
	out, err := runTimew(args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func exportHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	trange, _ := argsMap["range"].(string)
	args := []string{"export"}
	if trange != "" {
		args = append(args, trange)
	}
	out, err := runTimew(args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}

func rawHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := req.Params.Arguments.(map[string]any)
	cmd, _ := argsMap["command"].(string)
	out, err := runTimew(strings.Fields(cmd)...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(out), nil
}
