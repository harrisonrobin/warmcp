package timewarrior

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

func TestTimewStart(t *testing.T) {
	mock := &MockRunner{Output: "Tracking Work"}
	common.Runner = mock

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"tags": "Work",
	}

	res, err := startHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "Tracking Work")
	assert.Equal(t, "timew", mock.LastCmd)
	assert.Contains(t, mock.LastArgs, "start")
	assert.Contains(t, mock.LastArgs, "Work")
}

func TestTimewSummary(t *testing.T) {
	mock := &MockRunner{Output: "Summary data"}
	common.Runner = mock

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"range": ":day",
	}

	res, err := summaryHandler(context.Background(), req)
	assert.NoError(t, err)
	assert.Contains(t, res.Content[0].(mcp.TextContent).Text, "Summary data")
	assert.Equal(t, "timew", mock.LastCmd)
	assert.Contains(t, mock.LastArgs, "summary")
	assert.Contains(t, mock.LastArgs, ":day")
}
