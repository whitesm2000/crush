package tools

import "sync"

// foregroundBash tracks a session's in-flight foreground bash command so it
// can be pushed to the background on demand (ctrl+b in the TUI). The push
// channel is buffered with a slot for one signal; sends are non-blocking so
// a push with no waiter is simply a no-op.
type foregroundBash struct {
	push chan struct{}
}

var foregroundBashes sync.Map // session ID -> *foregroundBash

// PushSessionBashBackground asks the session's currently running foreground
// bash command to convert itself into a background job immediately. It
// reports whether a foreground command was running to receive the push.
func PushSessionBashBackground(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	v, ok := foregroundBashes.Load(sessionID)
	if !ok {
		return false
	}
	select {
	case v.(*foregroundBash).push <- struct{}{}:
		return true
	default:
		// A push is already pending; treat as delivered.
		return true
	}
}

// registerForegroundBash records the handle for a session's foreground bash
// command and returns it. The caller must unregister when the command ends.
func registerForegroundBash(sessionID string) *foregroundBash {
	handle := &foregroundBash{push: make(chan struct{}, 1)}
	foregroundBashes.Store(sessionID, handle)
	return handle
}

// unregisterForegroundBash removes the session's foreground handle, keeping
// any entry another command already registered (compare-and-delete style).
func unregisterForegroundBash(sessionID string, handle *foregroundBash) {
	foregroundBashes.CompareAndDelete(sessionID, handle)
}
