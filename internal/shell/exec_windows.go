//go:build windows

package shell

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"time"

	"mvdan.cc/sh/v3/interp"
)

// defaultKillTimeout matches mvdan's DefaultExecHandler default. Extracted
// so the coupling to upstream is explicit rather than a buried literal.
const defaultKillTimeout = 2 * time.Second

// isolateProcess is a no-op on Windows. Session isolation via Setsid is a
// Unix-only concept; Windows uses CREATE_NEW_PROCESS_GROUP which mvdan's
// default handler already handles adequately.
func isolateProcess(_ *exec.Cmd) {}

// killProcessTree terminates the entire process tree rooted at pid using
// taskkill /T /F. On Windows, terminating the direct child alone leaves
// grandchildren running (npm -> node -> chrome, python spawning helpers),
// which keeps pipes open and makes cancelled commands look hung.
func killProcessTree(pid int) {
	kill := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
	_ = kill.Run()
}

// processGroupExecHandler returns an ExecHandlerFunc that kills the entire
// child process tree when the context is cancelled, instead of only the
// direct child as interp.DefaultExecHandler does. The implementation mirrors
// the Unix variant: start the command, and on cancellation first give the
// tree a brief grace period via taskkill's graceful path before forcing.
func processGroupExecHandler(killTimeout time.Duration) interp.ExecHandlerFunc {
	return func(ctx context.Context, args []string) error {
		hc := interp.HandlerCtx(ctx)
		path, err := interp.LookPathDir(hc.Dir, hc.Env, args[0])
		if err != nil {
			fmt.Fprintln(hc.Stderr, err)
			return interp.ExitStatus(127)
		}

		cmd := exec.Cmd{
			Path:   path,
			Args:   args,
			Env:    execEnvList(hc.Env),
			Dir:    hc.Dir,
			Stdin:  hc.Stdin,
			Stdout: hc.Stdout,
			Stderr: hc.Stderr,
		}

		err = cmd.Start()
		if err == nil {
			stopf := context.AfterFunc(ctx, func() {
				if killTimeout > 0 {
					// Grace period before the force kill: a well-behaved
					// command may notice its cancelled context on its own.
					time.Sleep(killTimeout)
				}
				killProcessTree(cmd.Process.Pid)
			})
			defer stopf()

			err = cmd.Wait()
		}

		return exitStatusFromError(ctx, hc.Stderr, err)
	}
}

// exitStatusFromError translates an exec error into an interp exit status,
// matching the conventions of interp.DefaultExecHandler. Windows has no
// syscall.WaitStatus signal reporting, so exit codes pass through directly.
func exitStatusFromError(_ context.Context, stderr io.Writer, err error) error {
	if err == nil {
		return nil
	}
	switch err := err.(type) {
	case *exec.ExitError:
		return interp.ExitStatus(uint8(err.ExitCode()))
	case *exec.Error:
		fmt.Fprintf(stderr, "%v\n", err)
		return interp.ExitStatus(127)
	default:
		return err
	}
}
