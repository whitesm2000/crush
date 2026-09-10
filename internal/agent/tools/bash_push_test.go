package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/shell"
	"github.com/stretchr/testify/require"
)

// TestBashTool_PushToBackground verifies that PushSessionBashBackground
// converts a running foreground command into a background job immediately,
// long before the auto-background threshold fires.
func TestBashTool_PushToBackground(t *testing.T) {
	workingDir := t.TempDir()
	tool := newBashToolForTest(workingDir)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "push-test-session")

	done := make(chan fantasy.ToolResponse, 1)
	start := time.Now()

	go func() {
		done <- runBashTool(t, tool, ctx, BashParams{
			Description: "long running command",
			Command:     "ping -n 60 127.0.0.1 > NUL",
			// Long threshold so only the push can trigger backgrounding.
			AutoBackgroundAfter: 300,
		})
	}()

	// Give the command time to start, then push it to the background.
	time.Sleep(2 * time.Second)
	require.True(t, PushSessionBashBackground("push-test-session"))
	require.False(t, PushSessionBashBackground("no-such-session"))

	resp := <-done
	elapsed := time.Since(start)

	require.Less(t, elapsed, 30*time.Second, "push should return long before the command finishes")
	var meta BashResponseMetadata
	require.NoError(t, json.Unmarshal([]byte(resp.Metadata), &meta))
	require.True(t, meta.Background)
	require.NotEmpty(t, meta.ShellID)

	bgManager := shell.GetBackgroundShellManager()
	require.NoError(t, bgManager.Kill(meta.ShellID))
}
