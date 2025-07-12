package action

import (
	"context"
	"errors"
	"fmt"
	"github.com/launchrctl/launchr/internal/launchr"
	"os"
	"os/exec"
)

type runtimeShell struct {
	WithLogger
}

// NewShellRuntime creates a new action shell runtime.
func NewShellRuntime() Runtime {
	return &runtimeShell{}
}

func (r *runtimeShell) Clone() Runtime {
	return NewShellRuntime()
}

func (r *runtimeShell) Init(_ context.Context, _ *Action) (err error) {
	return nil
}

func (r *runtimeShell) Execute(ctx context.Context, a *Action) (err error) {
	log := r.LogWith("run_env", "shell", "action_id", a.ID)
	log.Debug("starting execution of the action")

	streams := a.Input().Streams()
	rt := a.RuntimeDef()

	shell, cbin, err := getShellAndExecutable()
	if err != nil {
		return err
	}
	log.Debug("using shell", "shell", shell)

	cmd := exec.CommandContext(ctx, shell, "-l", "-c", rt.Shell.Script) //nolint:gosec // G204 user script is expected.
	cmd.Dir = a.WorkDir()
	cmd.Env = append(getShellEnv(), rt.Shell.Env...)
	// TODO: Add ACTION_DIR and other
	cmd.Env = append(cmd.Env, "CBIN="+cbin)
	cmd.Stdout = streams.Out()
	cmd.Stderr = streams.Err()
	// Do no attach stdin, as it may not work as expected.

	err = cmd.Start()
	if err != nil {
		return err
	}

	// If we attached with TTY, all signals will be processed by a child process.
	sigc := launchr.NotifySignals()
	go launchr.HandleSignals(ctx, sigc, func(s os.Signal, _ string) error {
		log.Debug("forwarding signal for action", "sig", s, "pid", cmd.Process.Pid)
		return cmd.Process.Signal(s)
	})
	defer launchr.StopCatchSignals(sigc)

	cmdErr := cmd.Wait()
	var exitErr *exec.ExitError
	if errors.As(cmdErr, &exitErr) {
		exitCode := exitErr.ExitCode()
		msg := fmt.Sprintf("action %q finished with exit code %d", a.ID, exitCode)
		// Process was interrupted.
		if exitCode == -1 {
			exitCode = 130
			msg = fmt.Sprintf("action %q was interrupted, finished with exit code %d", a.ID, exitCode)
		}
		log.Info("action finished with the exit code", "exit_code", exitCode)
		return launchr.NewExitError(exitCode, msg)
	}
	return cmdErr
}

func (r *runtimeShell) Close() error {
	return nil
}

func getShellEnv() []string {
	// TODO: Filter PATH etc, it must come from login
	// TODO: Add CBIN
	return os.Environ()
}
