package launchr

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"k8s.io/utils/strings/slices"

	coretest "github.com/launchrctl/launchr/test"
)

func TestMain(m *testing.M) {
	// Set testscript version.
	version = "v0.0.0-testscript"
	builtWith = "testscript v0.0.0"
	testscript.Main(m, map[string]func(){
		"launchr": RunAndExit,
		"testapp": func() {
			// Set global application name.
			name = "testapp"
			RunAndExit()
		},
	})
}

func TestScriptAll(t *testing.T) {
	t.Parallel()
	type tsSetupfn = func(*testscript.Env) error
	type testcase struct {
		name      string
		dir       string
		setup     []tsSetupfn
		skipShort bool
		skipOS    []string
	}
	testcases := []testcase{
		{
			name:      "build",
			dir:       "test/testdata/build",
			setup:     []tsSetupfn{setupBuildEnv},
			skipShort: true,
			skipOS:    []string{"windows"},
		},
		{name: "common", dir: "test/testdata/common"},
		{name: "action/discovery", dir: "test/testdata/action/discovery"},
		{name: "action/input", dir: "test/testdata/action/input"},

		{name: "runtime/shell", dir: "test/testdata/runtime/shell"},
		{
			name:      "runtime/container/docker",
			dir:       "test/testdata/runtime/container",
			setup:     []tsSetupfn{coretest.SetupDockerEnv},
			skipShort: true,
		},
	}
	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipShort && testing.Short() {
				t.Skip("skipping test in short mode")
			}
			if slices.Contains(tt.skipOS, runtime.GOOS) {
				t.Skip("skipping test on " + runtime.GOOS)
			}
			t.Parallel()
			testscript.Run(t, testscript.Params{
				Dir:      tt.dir,
				Cmds:     coretest.TestScriptCmds,
				Deadline: time.Now().Add(30 * time.Second),

				RequireExplicitExec: true,
				RequireUniqueNames:  true,
				ContinueOnError:     true,

				Setup: func(env *testscript.Env) error {
					for _, fn := range tt.setup {
						if err := fn(env); err != nil {
							return err
						}
					}
					return nil
				},
			})
		})
	}
}

func setupBuildEnv(env *testscript.Env) error {
	repoPath := MustAbs("./")
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	env.Vars = append(
		env.Vars,
		"REPO_PATH="+repoPath,
		"CORE_PKG="+PkgPath,
		"REAL_HOME="+home,
	)
	return nil
}
