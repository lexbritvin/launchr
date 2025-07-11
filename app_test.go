package launchr

import (
	"os"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/launchrctl/launchr/internal/launchr"
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

func TestBinary(t *testing.T) {
	t.Parallel()
	type tsSetupfn = func(*testscript.Env) error
	type testcase struct {
		name      string
		dir       string
		files     []string
		setup     []tsSetupfn
		skipShort bool
		skipOS    []string
		timeout   time.Duration
		conseq    bool
	}
	testcases := []testcase{
		{name: "common", dir: "test/testdata/common", timeout: 30 * time.Second},
		{name: "action/discovery", dir: "test/testdata/action/discovery", timeout: 30 * time.Second},
		{name: "action/input", dir: "test/testdata/action/input", timeout: 30 * time.Second},

		// Runtime Shell.
		{name: "runtime/shell", dir: "test/testdata/runtime/shell", timeout: 30 * time.Second},
		// Runtime Docker.
		{
			name:      "runtime/container/docker",
			dir:       "test/testdata/runtime/container",
			setup:     []tsSetupfn{coretest.SetupEnvDocker, coretest.SetupEnvRandom},
			skipShort: true,
			timeout:   120 * time.Second, // Download and build of images may take time on cold run.
		},

		// Test binary build using self.
		// This test must run last and should not be parallelized,
		// so that the build cache is warm after it.
		// If it fails due to a timeout, try warming the cache manually with `make build`.
		{
			// Run the build once to warm up the build cache.
			name:      "build-warmup",
			files:     []string{"test/testdata/build/no-cache.txtar"},
			setup:     []tsSetupfn{setupBuildEnv},
			skipShort: true,
			skipOS:    []string{"windows"},
			timeout:   120 * time.Second,
			conseq:    true,
		},
		{
			name:      "build",
			dir:       "test/testdata/build",
			setup:     []tsSetupfn{setupBuildEnv},
			skipShort: true,
			skipOS:    []string{"windows"},
			timeout:   120 * time.Second,
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
			if !tt.conseq {
				t.Parallel()
			}
			var deadline time.Time
			if tt.timeout != 0 && !launchr.Version().Debug {
				deadline = time.Now().Add(tt.timeout)
			}

			testscript.Run(t, testscript.Params{
				Dir:      tt.dir,
				Files:    tt.files,
				Cmds:     coretest.CmdsTestScript(),
				Deadline: deadline,

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
