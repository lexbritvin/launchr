package launchr

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	coretest "github.com/launchrctl/launchr/test"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"launchr": func() {
			// Set testscript version.
			RunAndExit()
		},
		"testapp": func() {
			// Set global application name.
			name = "testapp"
			RunAndExit()
		},
	})
}

// TestScriptBuild tests how binary builds and outputs version.
func TestScriptBuild(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:                 "test/testdata/build",
		RequireExplicitExec: true,
		RequireUniqueNames:  true,
		Setup: func(env *testscript.Env) error {
			repoPath := MustAbs("./")
			env.Vars = append(
				env.Vars,
				"REPO_PATH="+repoPath,
				"CORE_PKG="+PkgPath,
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
