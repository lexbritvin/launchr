package launchr

import (
	"os"
	"runtime"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

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

// TestScriptBuild tests how binary builds and outputs version.
func TestScriptBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows")
	}
	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir:                 "test/testdata/build",
		RequireExplicitExec: true,
		RequireUniqueNames:  true,
		Setup: func(env *testscript.Env) error {
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
		},
	})
}

func TestScriptAll(t *testing.T) {
	t.Parallel()
	type testcase struct {
		name      string
		dir       string
		skipShort bool
	}
	testcases := []testcase{
		{name: "common", dir: "test/testdata/common"},
		{name: "action/discovery", dir: "test/testdata/action/discovery"},
		{name: "action/input", dir: "test/testdata/action/input"},
		{name: "runtime/container", dir: "test/testdata/runtime/container", skipShort: true},
		{name: "runtime/shell", dir: "test/testdata/runtime/shell"},
	}
	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipShort && testing.Short() {
				t.Skip()
			}
			t.Parallel()
			testscript.Run(t, testscript.Params{
				Dir:                 tt.dir,
				RequireExplicitExec: true,
				RequireUniqueNames:  true,
				ContinueOnError:     true,
				Cmds:                coretest.TestScriptCmds,
			})
		})
	}
}
