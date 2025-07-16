//go:build windows

package action

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/launchrctl/launchr/internal/launchr"
)

func getCurrentUser() userInfo {
	// Use neutral 1000 when we can't get UID on Windows.
	const defaultUID = "1000"
	const defaultGID = "1000"
	return userInfo{
		UID: defaultUID,
		GID: defaultGID,
	}
}

func normalizeContainerMountPath(path string) string {
	path = launchr.MustAbs(path)
	// Convert windows paths C:\my\path -> /c/my/path for docker daemon.
	return "/mnt" + launchr.ConvertWindowsPath(path)
}

func isWSLShell(ctx context.Context, shell string) bool {
	checkWslCmd := exec.CommandContext(ctx, shell, "-c", "uname -r")
	wslOut := &strings.Builder{}
	checkWslCmd.Stdout = wslOut
	err := checkWslCmd.Run()
	if err != nil {
		return false
	}
	return strings.Contains(wslOut.String(), "WSL")
}

func getShellContext() (*shellContext, error) {
	path, err := findWindowsBash()
	if err != nil {
		return nil, err
	}
	cbin, env := getExecutableWindowsUnix(path)

	return &shellContext{
		Shell: path,
		Exec:  cbin,
		Env:   env,
	}, nil
}

func findWindowsBash() (string, error) {
	candidates := []string{
		"C:\\cygwin64\\bin\\bash.exe",
		"C:\\msys64\\usr\\bin\\bash.exe",
		"C:\\Program Files\\Git\\bin\\bash.exe",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	// Fallback to LookPath, it may give a bash of WSL
	// which is not exactly host but will run bash just fine.
	path, err := exec.LookPath("bash")
	if err != nil {
		return nil, err
	}
	return path, nil
}

func getExecutableWindowsUnix(shell string) (string, []string) {
	ctx := context.Background()
	// Get the path of the executable on the host.
	currentBin, err := os.Executable()
	env := os.Environ()
	if err != nil {
		return launchr.Version().Name, env
	}

	if isWSLShell(ctx, shell) {
		return normalizeContainerMountPath(currentBin), env
	}

	return launchr.ConvertWindowsPath(currentBin), env
}
