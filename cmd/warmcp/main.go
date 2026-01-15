package main

import (
	"fmt"
	"warmcp/pkg/common"
	"warmcp/pkg/taskwarrior"
	"warmcp/pkg/timewarrior"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"warmcp",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	taskwarrior.RegisterHandlers(s)
	timewarrior.RegisterHandlers(s)
	common.RegisterMCPFeatures(s)

	fmt.Println("warmcp server starting on stdio...")
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Fatal: %v\n", err)
	}
}
