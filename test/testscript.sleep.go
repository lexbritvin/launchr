package test

import (
	"strconv"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
)

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
