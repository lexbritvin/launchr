//go:build windows

package action

import (
	"context"
	"github.com/launchrctl/launchr/internal/launchr"
	"os"
	"os/exec"
	"strings"
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

func getShellAndExecutable() (string, string, error) {
	candidates := []string{
		"C:\\cygwin64\\bin\\bash.exe",
		"C:\\msys64\\usr\\bin\\bash.exe",
		"C:\\Program Files\\Git\\bin\\bash.exe",
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, getExecutableUnix(path), nil
		}
	}
	// Fallback to LookPath, it may give a bash of WSL
	// which is not exactly host but will run bash just fine.
	path, err := exec.LookPath("bash")
	if err != nil {
		return "", "", err
	}

	return path, getExecutableUnix(path), nil
}

func getExecutableUnix(shell string) string {
	ctx := context.Background()
	// Get the path of the executable on the host.
	currentBin, err := os.Executable()
	if err != nil {
		return launchr.Version().Name
	}

	if isWSLShell(ctx, shell) {
		return normalizeContainerMountPath(currentBin)
	}

	return launchr.ConvertWindowsPath(currentBin)
}
