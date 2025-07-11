// Package test contains functionality to test the application with testscript.
package test

import (
	"bytes"
	"os"
	"strconv"
	"time"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/launchrctl/launchr/internal/launchr"
	_ "github.com/launchrctl/launchr/test/plugins" // Include test plugins.
)

// CmdsTestScript provides custom commands for testscript execution.
func CmdsTestScript() map[string]func(ts *testscript.TestScript, neg bool, args []string) {
	return map[string]func(ts *testscript.TestScript, neg bool, args []string){
		// txtproc provides flexible text processing capabilities
		// Usage:
		//	txtproc replace 'old' 'new' input.txt output.txt
		//	txtproc replace-regex 'pattern' 'replacement' input.txt output.txt
		//	txtproc remove-lines 'pattern' input.txt output.txt
		//	txtproc remove-regex 'pattern' input.txt output.txt
		//	txtproc extract-lines 'pattern' input.txt output.txt
		//	txtproc extract-regex 'pattern' input.txt output.txt
		"txtproc":  CmdTxtProc,
		"dos2unix": CmdDos2unix,
		// sleep pauses execution for a specified duration
		// Usage:
		//  sleep <duration>
		// Examples:
		//	sleep 1s
		//	sleep 500ms
		//	sleep 2m
		"sleep": CmdSleep,
		// dlv runs the given binary with Delve for debugging.
		// Please, note that the test must be run with debug headers for it to work.
		// Usage:
		//   dlv <app_name>
		"dlv": CmdDlv,
	}
}

// SetupEnvDocker configures docker backend in the test environment.
func SetupEnvDocker(env *testscript.Env) error {
	env.Vars = append(
		env.Vars,
		// Passthrough Docker env variables if set.
		"DOCKER_HOST="+os.Getenv("DOCKER_HOST"),
		"DOCKER_TLS_VERIFY="+os.Getenv("DOCKER_TLS_VERIFY"),
		"DOCKER_CERT_PATH="+os.Getenv("DOCKER_CERT_PATH"),
	)
	return nil
}

// SetupEnvRandom sets up a random environment variable.
func SetupEnvRandom(env *testscript.Env) error {
	env.Vars = append(
		env.Vars,
		"RANDOM="+launchr.GetRandomString(8),
	)
	return nil
}

// CmdSleep pauses execution for a specified duration
func CmdSleep(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("sleep does not support negation")
	}

	if len(args) != 1 {
		ts.Fatalf("sleep: usage: sleep <duration>")
	}

	duration, err := time.ParseDuration(args[0])
	if err != nil {
		// Try parsing as seconds if it's just a number
		if seconds, numErr := strconv.ParseFloat(args[0], 64); numErr == nil {
			duration = time.Duration(seconds * float64(time.Second))
		} else {
			ts.Fatalf("sleep: invalid duration %q: %v", args[0], err)
		}
	}

	if duration < 0 {
		ts.Fatalf("sleep: duration cannot be negative")
	}

	time.Sleep(duration)
}

// CmdDos2unix converts CRLF line endings to LF in the specified files
func CmdDos2unix(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! dos2unix")
	}
	if len(args) < 1 {
		ts.Fatalf("usage: dos2unix paths...")
	}
	for _, file := range args {
		// Get absolute path relative to test directory
		absPath := ts.MkAbs(file)

		// Check if file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			ts.Fatalf("dos2unix: file %s does not exist", file)
		}

		// Read file content
		content, err := os.ReadFile(absPath)
		if err != nil {
			ts.Fatalf("dos2unix: failed to read %s: %v", file, err)
		}

		// Convert CRLF to LF
		content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
		// Also handle standalone CR (classic Mac line endings)
		content = bytes.ReplaceAll(content, []byte("\r"), []byte("\n"))

		// Write back to file, preserving original file permissions
		fileInfo, err := os.Stat(absPath)
		if err != nil {
			ts.Fatalf("dos2unix: failed to get file info for %s: %v", file, err)
		}

		err = os.WriteFile(absPath, content, fileInfo.Mode())
		if err != nil {
			ts.Fatalf("dos2unix: failed to write %s: %v", file, err)
		}
	}
}
