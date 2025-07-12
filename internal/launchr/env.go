package launchr

import (
	"os"
	"regexp"
	"strings"
	"syscall"
)

// Application environment variables.
const (
	// EnvVarRootParentPID defines parent process id. May be used by forked processes.
	EnvVarRootParentPID = EnvVar("root_ppid")
	// EnvVarActionsPath defines path where to search for actions.
	EnvVarActionsPath = EnvVar("actions_path")
	// EnvVarLogLevel defines currently set log level, see --log-level or -v flag.
	EnvVarLogLevel = EnvVar("log_level")
	// EnvVarLogFormat defines currently set log format, see --log-format flag.
	EnvVarLogFormat = EnvVar("log_format")
	// EnvVarQuietMode defines if the application should output anything, see --quiet flag.
	EnvVarQuietMode = EnvVar("quiet_mode")
)

// EnvVar defines an environment variable and provides an interface to interact with it
// by prefixing the current app name.
// For example, if "my_var" is given as the variable name and the app name is "launchr",
// the accessed environment variable will be "LAUNCHR_MY_VAR".
type EnvVar string

// String implements [fmt.Stringer] interface.
func (key EnvVar) String() string {
	return strings.ToUpper(name + "_" + string(key))
}

// EnvString returns an os string of env variable with a value val.
func (key EnvVar) EnvString(val string) string {
	return key.String() + "=" + val
}

// Get returns env variable value.
func (key EnvVar) Get() string {
	return os.Getenv(key.String())
}

// Set sets env variable.
func (key EnvVar) Set(val string) error {
	return os.Setenv(key.String(), val)
}

// Unset unsets env variable.
func (key EnvVar) Unset() error {
	return os.Unsetenv(key.String())
}

// Global regex patterns for parameter expansion
var (
	// Matches ${var-default}, ${var:-default}, ${var+default}, ${var:+default}
	paramExpansionRe = regexp.MustCompile(`\$\{([^}]+)}`)

	// Specific patterns for each expansion type
	defaultValueRe = regexp.MustCompile(`^([^:-]+)(-|:-)(.*)$`)   // ${var-default} or ${var:-default}
	alternativeRe  = regexp.MustCompile(`^([^:+]+)(\+|:\+)(.*)$`) // ${var+alternative} or ${var:+alternative}
)

// Getenv is an environment variable expand callback.
func Getenv(key string) string {
	if key == "$" {
		return "$"
	}

	// Handle parameter expansion patterns first
	if paramExpansionRe.MatchString(key) {
		return expandParameters(key)
	}

	// Replace all subexpressions for simple $VAR patterns
	if strings.Contains(key, "$") {
		key = os.Expand(key, Getenv)
	}

	// Get the actual environment variable
	v, _ := syscall.Getenv(key)
	return v
}

func expandParameters(input string) string {
	return paramExpansionRe.ReplaceAllStringFunc(input, func(match string) string {
		// Remove ${ and } to get the parameter expression
		expr := match[2 : len(match)-1]

		// Check for default value patterns: ${var-default} or ${var:-default}
		if matches := defaultValueRe.FindStringSubmatch(expr); matches != nil {
			varName := matches[1]
			operator := matches[2]
			defaultValue := matches[3]

			value, exists := syscall.Getenv(varName)

			switch operator {
			case "-":
				// ${var-default}: use default if variable is unset
				if !exists {
					return expandIfNeeded(defaultValue)
				}
				return value
			case ":-":
				// ${var:-default}: use default if a variable is unset OR empty
				if !exists || value == "" {
					return expandIfNeeded(defaultValue)
				}
				return value
			}
		}

		// Check for alternative value patterns: ${var+alternative} or ${var:+alternative}
		if matches := alternativeRe.FindStringSubmatch(expr); matches != nil {
			varName := matches[1]
			operator := matches[2]
			alternative := matches[3]

			value, exists := syscall.Getenv(varName)

			switch operator {
			case "+":
				// ${var+alternative}: use alternative if a variable is set (even if empty)
				if exists {
					return expandIfNeeded(alternative)
				}
				return ""
			case ":+":
				// ${var:+alternative}: use alternative if a variable is set AND non-empty
				if exists && value != "" {
					return expandIfNeeded(alternative)
				}
				return ""
			}
		}

		// If no pattern matches, treat as simple variable: ${var}
		value, _ := syscall.Getenv(expr)
		return value
	})
}

// Helper function to recursively expand variables in default/alternative values
func expandIfNeeded(value string) string {
	if strings.Contains(value, "$") {
		return Getenv(value)
	}
	return value
}
