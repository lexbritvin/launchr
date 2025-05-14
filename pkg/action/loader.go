package action

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/launchrctl/launchr/internal/launchr"
)

// Loader is an interface for loading an action file.
type Loader interface {
	// Content returns the raw file content.
	Content() ([]byte, error)
	// Load parses Content to a Definition with substituted values.
	Load(LoadContext) (*Definition, error)
	// LoadRaw parses Content to a Definition raw values. Template strings are escaped.
	LoadRaw() (*Definition, error)
}

// LoadContext stores relevant and isolated data needed for processors.
type LoadContext struct {
	Action *Action
}

// LoadProcessor is an interface for processing input on load.
type LoadProcessor interface {
	// Process gets an input action file data and returns a processed result.
	Process(LoadContext, []byte) ([]byte, error)
}

type pipeProcessor struct {
	p []LoadProcessor
}

// NewPipeProcessor creates a new processor containing several processors that handles input consequentially.
func NewPipeProcessor(p ...LoadProcessor) LoadProcessor {
	return &pipeProcessor{p: p}
}

func (p *pipeProcessor) Process(ctx LoadContext, b []byte) ([]byte, error) {
	var err error
	for _, proc := range p.p {
		b, err = proc.Process(ctx, b)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}

type envProcessor struct{}

func (p envProcessor) Process(ctx LoadContext, b []byte) ([]byte, error) {
	pv := newPredefinedVars(ctx.Action)
	getenv := func(key string) string {
		v, ok := pv.getenv(key)
		if ok {
			return v
		}
		return launchr.Getenv(key)
	}
	s := os.Expand(string(b), getenv)
	return []byte(s), nil
}

type inputProcessor struct{}

var rgxTplVar = regexp.MustCompile(`{{.*?\.(\S+).*?}}`)

type errMissingVar struct {
	vars map[string]struct{}
}

// Error implements error interface.
func (err errMissingVar) Error() string {
	f := make([]string, 0, len(err.vars))
	for k := range err.vars {
		f = append(f, k)
	}
	return fmt.Sprintf("the following variables were used but never defined: %v", f)
}

func (p inputProcessor) Process(ctx LoadContext, b []byte) ([]byte, error) {
	if ctx.Action == nil {
		return b, nil
	}
	a := ctx.Action
	def := ctx.Action.ActionDef()
	// Collect template variables.
	data := ConvertInputToTplVars(a.Input(), def)
	addPredefinedVariables(data, a)

	// Parse action without variables to validate
	tpl := template.New(a.ID)
	_, err := tpl.Parse(string(b))
	// Check if variables have dashes to show the error properly.
	err = checkDashErr(err, data)
	if err != nil {
		return nil, err
	}

	// Execute template.
	buf := bytes.NewBuffer(make([]byte, 0, len(b)))
	err = tpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}

	// Find if some vars were used but not defined.
	res := buf.Bytes()
	err = findMissingVars(b, res, data)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ConvertInputToTplVars creates a map with input variables suitable for template engine.
func ConvertInputToTplVars(input *Input, ac *DefAction) map[string]any {
	args := input.Args()
	opts := input.Opts()
	values := make(map[string]any, len(args)+len(opts))

	// Collect arguments and options values.
	collectInputVars(values, args, ac.Arguments)
	collectInputVars(values, opts, ac.Options)

	// @todo consider boolean, it's strange in output - "true/false"
	// @todo handle array options

	return values
}

func collectInputVars(values map[string]any, params InputParams, def ParametersList) {
	for _, pdef := range def {
		key := pdef.Name
		// Set value: default or input parameter.
		dval := pdef.Default
		values[key] = dval
		values[replDashes.Replace(key)] = dval
		if v, ok := params[pdef.Name]; ok {
			// Allow usage of dashed variable names like "my-name" by replacing dashes to underscores.
			values[key] = v
			values[replDashes.Replace(key)] = v
		}
	}
}

func addPredefinedVariables(data map[string]any, a *Action) {
	// TODO: Deprecated, use env variables.
	pv := newPredefinedVars(a)
	for k, v := range pv.templateData() {
		data[k] = v
	}
}

func checkDashErr(err error, data map[string]any) error {
	if err == nil {
		return nil
	}
	// Check if variables have dashes to show the error properly.
	hasDash := false
	for k := range data {
		if strings.Contains(k, "-") {
			hasDash = true
			break
		}
	}
	if hasDash && strings.Contains(err.Error(), "bad character U+002D '-'") {
		return fmt.Errorf(`unexpected '-' symbol in a template variable. 
Action definition is correct, but dashes are not allowed in templates, replace "-" with "_" in {{ }} blocks`)
	}
	return err
}

func findMissingVars(orig, repl []byte, data map[string]any) error {
	miss := make(map[string]struct{})
	if !bytes.Contains(repl, []byte("<no value>")) {
		return nil
	}
	matches := rgxTplVar.FindAllSubmatch(orig, -1)
	for _, m := range matches {
		k := string(m[1])
		if _, ok := data[k]; !ok {
			miss[k] = struct{}{}
		}
	}
	// If we don't have parameter names, the values probably were nil.
	// It's ok, users will be able to identify missing parameters.
	if len(miss) != 0 {
		return errMissingVar{miss}
	}
	return nil
}
