package taskwarrior

import (
	"context"
	"fmt"
	"testing"
	"warmcp/pkg/common"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

type MockRunner struct {
	LastCmd  string
	LastEnv  []string
	LastArgs []string
	Output   string
	Err      error
}

func (m *MockRunner) Run(name string, env []string, baseArgs []string, args ...string) (string, error) {
	m.LastCmd = name
	m.LastEnv = env
	m.LastArgs = append(baseArgs, args...)
	return m.Output, m.Err
}

func TestTaskAdd(t *testing.T) {
	mock := &MockRunner{Output: "Created task 1."}
	common.Runner = mock

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"description": "Buy milk",
		"metadata":    "project:Home",
	}

	res, err := addHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "Created task 1.")
	assert.Equal(t, "task", mock.LastCmd)
	assert.Contains(t, mock.LastArgs, "Buy milk")
	assert.Contains(t, mock.LastArgs, "project:Home")
}

func TestTaskList(t *testing.T) {
	mock := &MockRunner{Output: "[{\"description\":\"Task 1\"}]"}
	common.Runner = mock

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"filter": "+PENDING",
	}

	res, err := listHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "Task 1")
	assert.Equal(t, "task", mock.LastCmd)
	assert.Contains(t, mock.LastArgs, "+PENDING")
	assert.Contains(t, mock.LastArgs, "export")
}

func TestTaskModifyComplex(t *testing.T) {
	mock := &MockRunner{Output: "Modified 1 task."}
	common.Runner = mock

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"filter":        "project:Work +PENDING",
		"modifications": "priority:H due:tomorrow",
	}

	res, err := modifyHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "Modified 1 task.")
	assert.Equal(t, "task", mock.LastCmd)
	// Check order: overrides, then filters, then command, then modifications
	assert.Equal(t, "rc.confirmation=off", mock.LastArgs[0])
	assert.Contains(t, mock.LastArgs, "project:Work")
	assert.Contains(t, mock.LastArgs, "+PENDING")
	assert.Contains(t, mock.LastArgs, "modify")
	assert.Contains(t, mock.LastArgs, "priority:H")
	assert.Contains(t, mock.LastArgs, "due:tomorrow")
}

func TestTaskErrorHandling(t *testing.T) {
	mock := &MockRunner{Err: fmt.Errorf("exit status 1"), Output: "Error: Task not found"}
	common.Runner = mock

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"uuid": "invalid-uuid"}

	res, err := doneHandler(context.Background(), req)
	assert.NoError(t, err) // Handlers return MCP error results, not Go errors
	assert.True(t, res.IsError)
	assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "Task not found")
}

func TestTaskPurge(t *testing.T) {
	mock := &MockRunner{Output: "Purged 1 task."}
	common.Runner = mock

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"filter": "status:deleted"}

	res, err := purgeHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "task", mock.LastCmd)
	assert.Contains(t, mock.LastArgs, "purge")
	assert.Contains(t, mock.LastArgs, "status:deleted")
	assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "Purged 1 task.")
}

func TestTaskAppend(t *testing.T) {
	mock := &MockRunner{Output: "Appended to task."}
	common.Runner = mock

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"uuid": "uuid1", "text": "more info"}

	res, err := appendHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "Appended to task.")
	assert.Contains(t, mock.LastArgs, "append")
	assert.Contains(t, mock.LastArgs, "more info")
}

func TestTaskImport(t *testing.T) {
	mock := &MockRunner{Output: "Imported 1 task."}
	common.Runner = mock

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"json_data": "[{\"description\":\"new task\"}]"}

	res, err := importHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "Imported 1 task.")
	assert.Contains(t, mock.LastArgs, "import")
	// Verify that the argument looks like a temp file path
	assert.Contains(t, mock.LastArgs[len(mock.LastArgs)-1], "task_import")
}
