//go:build unix

package action

import (
	"os"
	"os/exec"
	osuser "os/user"

	"github.com/launchrctl/launchr/internal/launchr"
)

func getCurrentUser() userInfo {
	// If running in a container native environment, run container as a current user.
	curuser := userInfo{}
	u, err := osuser.Current()
	if err == nil {
		curuser.UID = u.Uid
		curuser.GID = u.Gid
	}
	return curuser
}

func normalizeContainerMountPath(path string) string {
	return launchr.MustAbs(path)
}

func getShellContext() (*shellContext, error) {
	currentBin, err := os.Executable()
	if err != nil {
		currentBin = launchr.Version().Name
	}
	defaultShell := os.Getenv("SHELL")
	shctx := &shellContext{
		Shell: defaultShell,
		Exec:  currentBin,
		Env:   os.Environ(),
	}
	if defaultShell == "" {
		path, err := exec.LookPath("bash")
		if err != nil {
			return nil, err
		}
		shctx.Shell = path
	}
	return shctx, nil
}
