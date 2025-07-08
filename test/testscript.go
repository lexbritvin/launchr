// Package test contains functionality to test
package test

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	_ "unsafe" // Include an internal method of the testscript module.

	"github.com/rogpeppe/go-internal/testscript"

	// Include test plugins.
	_ "github.com/launchrctl/launchr/test/plugins"
)

// Constants for repeated string values
const (
	opReplace      = "replace"
	opReplaceRegex = "replace-regex"
	opRemoveLines  = "remove-lines"
	opRemoveRegex  = "remove-regex"
	opExtractLines = "extract-lines"
	opExtractRegex = "extract-regex"
)

// TestScriptCmds provides custom commands for testscript execution
var TestScriptCmds = map[string]func(ts *testscript.TestScript, neg bool, args []string){
	"txtproc": txtprocCmd,
	"dlv":     dlvCmd,
	"sleep":   sleepCmd,
}

// sleepCmd pauses execution for a specified duration
// Usage: sleep <duration>
// Examples:
//
//	sleep 1s
//	sleep 500ms
//	sleep 2m
func sleepCmd(ts *testscript.TestScript, neg bool, args []string) {
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

// txtprocCmd provides flexible text processing capabilities
// Usage examples:
//
//	txtproc replace 'old' 'new' input.txt output.txt
//	txtproc replace-regex 'pattern' 'replacement' input.txt output.txt
//	txtproc remove-lines 'pattern' input.txt output.txt
//	txtproc remove-regex 'pattern' input.txt output.txt
//	txtproc extract-lines 'pattern' input.txt output.txt
//	txtproc extract-regex 'pattern' input.txt output.txt
func txtprocCmd(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("txtproc does not support negation")
	}

	if len(args) < 3 {
		ts.Fatalf("txtproc: usage: txtproc <operation> [args...] <input> <output>")
	}

	operation := args[0]
	var inputFile, outputFile string
	var pattern, replacement string

	switch operation {
	case opReplace:
		if len(args) != 5 {
			ts.Fatalf("txtproc replace: usage: txtproc replace <pattern> <replacement> <input> <output>")
		}
		pattern = args[1]
		replacement = args[2]
		inputFile = args[3]
		outputFile = args[4]

	case opReplaceRegex:
		if len(args) != 5 {
			ts.Fatalf("txtproc replace-regex: usage: txtproc replace-regex <regex> <replacement> <input> <output>")
		}
		pattern = args[1]
		replacement = args[2]
		inputFile = args[3]
		outputFile = args[4]

	case opRemoveLines, opRemoveRegex, opExtractLines, opExtractRegex:
		if len(args) != 4 {
			ts.Fatalf("txtproc %s: usage: txtproc %s <pattern> <input> <output>", operation, operation)
		}
		pattern = args[1]
		inputFile = args[2]
		outputFile = args[3]

	default:
		ts.Fatalf("txtproc: unknown operation %q. Available: replace, replace-regex, remove-lines, remove-regex, extract-lines, extract-regex", operation)
	}

	// Read input content
	var content string
	var err error

	if inputFile == "stdout" {
		// Special case: read from testscript's stdout buffer
		content = ts.Getenv("stdout")
		if content == "" {
			// Try to read stdout content using testscript's internal mechanism
			// This is a workaround since testscript doesn't expose stdout directly
			ts.Fatalf("txtproc: no stdout content available. Make sure to run 'exec' command before using txtproc with stdout")
		}
	} else if inputFile == "stderr" {
		// Special case: read from testscript's stderr buffer
		content = ts.Getenv("stderr")
		if content == "" {
			ts.Fatalf("txtproc: no stderr content available. Make sure to run 'exec' command before using txtproc with stderr")
		}
	} else {
		// Regular file
		inputPath := ts.MkAbs(inputFile)
		// #nosec G304 - File path is validated by testscript framework
		contentBytes, readErr := os.ReadFile(inputPath)
		if readErr != nil {
			ts.Fatalf("txtproc: failed to read %s: %v", inputFile, readErr)
		}
		content = string(contentBytes)
	}

	// Process content
	result, err := processText(content, operation, pattern, replacement)
	if err != nil {
		ts.Fatalf("txtproc: %v", err)
	}

	// Write output file
	outputPath := ts.MkAbs(outputFile)
	// Use more restrictive file permissions for security
	err = os.WriteFile(outputPath, []byte(result), 0600)
	if err != nil {
		ts.Fatalf("txtproc: failed to write %s: %v", outputFile, err)
	}
}

func processText(content, operation, pattern, replacement string) (string, error) {
	switch operation {
	case opReplace:
		return strings.ReplaceAll(content, pattern, replacement), nil

	case opReplaceRegex:
		re, err := regexp.Compile("(?m)" + pattern)
		if err != nil {
			return "", fmt.Errorf("invalid regex %q: %v", pattern, err)
		}
		return re.ReplaceAllString(content, replacement), nil

	case opRemoveLines:
		lines := strings.Split(content, "\n")
		var result []string
		for _, line := range lines {
			if !strings.Contains(line, pattern) {
				result = append(result, line)
			}
		}
		return strings.Join(result, "\n"), nil

	case opRemoveRegex:
		re, err := regexp.Compile(pattern)
		if err != nil {
			return "", fmt.Errorf("invalid regex %q: %v", pattern, err)
		}
		lines := strings.Split(content, "\n")
		var result []string
		for _, line := range lines {
			if !re.MatchString(line) {
				result = append(result, line)
			}
		}
		return strings.Join(result, "\n"), nil

	case opExtractLines:
		lines := strings.Split(content, "\n")
		var result []string
		for _, line := range lines {
			if strings.Contains(line, pattern) {
				result = append(result, line)
			}
		}
		return strings.Join(result, "\n"), nil

	case opExtractRegex:
		re, err := regexp.Compile(pattern)
		if err != nil {
			return "", fmt.Errorf("invalid regex %q: %v", pattern, err)
		}
		lines := strings.Split(content, "\n")
		var result []string
		for _, line := range lines {
			if re.MatchString(line) {
				result = append(result, line)
			}
		}
		return strings.Join(result, "\n"), nil

	default:
		return "", fmt.Errorf("unknown operation %q", operation)
	}
}

//go:linkname lookPath github.com/rogpeppe/go-internal/internal/os/execpath.Look
func lookPath(file string, getenv func(string) string) (string, error)

// isDebugMode checks if the current binary is running in debug mode
func isDebugMode() bool {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return false
	}

	// Check for debug-related build settings
	for _, setting := range buildInfo.Settings {
		if setting.Key == "-gcflags" {
			// Check if gcflags contains debug-related flags
			if strings.Contains(setting.Value, "-N") && strings.Contains(setting.Value, "-l") {
				return true
			}
		}
	}
	return false
}

// dlvCmd implements a custom testscript command for debugging with Delve
func dlvCmd(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("dlv command does not support negation")
	}

	if len(args) < 1 {
		ts.Fatalf("dlv: missing binary name\nUsage: dlv <binary> [args...]")
	}

	// Check if running in debug mode
	if !isDebugMode() {
		ts.Fatalf("dlv command requires the tests to be run with debug flags")
	}

	command := args[0]
	binaryArgs := args[1:]
	if filepath.Base(command) == command {
		if lp, err := lookPath(command, ts.Getenv); err != nil {
			ts.Fatalf("error when looking for %s: %v", command, err)
		} else {
			command = lp
		}
	}

	// Find an available port
	port := findAvailablePort()

	// Log connection information
	ts.Logf("=== Delve Debug Server ===")
	ts.Logf("Debugging binary: %s", command)
	ts.Logf("Port: %d", port)
	ts.Logf("Connect with: dlv connect 127.0.0.1:%d", port)
	ts.Logf("GoLand Remote Debug: 127.0.0.1:%d", port)
	ts.Logf("=========================")

	// Build dlv command arguments
	cmdArgs := []string{
		"exec", command,
		"--listen=127.0.0.1:" + strconv.Itoa(port),
		"--headless=true",
		"--api-version=2",
		"--accept-multiclient",
	}

	// Add binary arguments if any
	if len(binaryArgs) > 0 {
		cmdArgs = append(cmdArgs, "--")
		cmdArgs = append(cmdArgs, binaryArgs...)
	}

	// Execute dlv using testscript's exec method
	_ = ts.Exec("dlv", cmdArgs...)
}

// findAvailablePort finds an available port starting from 2345
func findAvailablePort() int {
	for port := 2345; port <= 2355; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return 2345 // fallback
}

// isPortAvailable checks if a port is available
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
