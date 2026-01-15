package taskwarrior

import (
	"context"
	"testing"
	"warmcp/pkg/common"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

type MockRunner struct {
	LastCmd  string
	LastArgs []string
	Output   string
	Err      error
}

func (m *MockRunner) Run(name string, baseArgs []string, args ...string) (string, error) {
	m.LastCmd = name
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
