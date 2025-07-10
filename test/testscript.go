// Package test contains functionality to test the application with testscript.
package test

import (
	"os"

	"github.com/rogpeppe/go-internal/testscript"

	// Include test plugins.
	_ "github.com/launchrctl/launchr/test/plugins"
)

// TestScriptCmds provides custom commands for testscript execution
var TestScriptCmds = map[string]func(ts *testscript.TestScript, neg bool, args []string){
	"txtproc": txtprocCmd,
	"sleep":   sleepCmd,
	"dlv":     dlvCmd,
}

// SetupDockerEnv configures docker backend in the test environment.
func SetupDockerEnv(env *testscript.Env) error {
	env.Vars = append(
		env.Vars,
		// Passthrough Docker env variables if set.
		"DOCKER_HOST="+os.Getenv("DOCKER_HOST"),
		"DOCKER_TLS_VERIFY="+os.Getenv("DOCKER_TLS_VERIFY"),
		"DOCKER_CERT_PATH="+os.Getenv("DOCKER_CERT_PATH"),
	)
	return nil
}
