// Package printinput has a test plugin that outputs action input.
package printinput

import (
	"context"
	"fmt"
	"strings"

	"github.com/launchrctl/launchr/internal/launchr"
	"github.com/launchrctl/launchr/pkg/action"
)

func init() {
	launchr.RegisterPlugin(&Plugin{})
}

// Plugin is a test plugin declaration.
type Plugin struct{}

// PluginInfo implements [launchr.Plugin] interface.
func (p *Plugin) PluginInfo() launchr.PluginInfo {
	return launchr.PluginInfo{}
}

// OnAppInit implements [launchr.OnAppInitPlugin] interface.
func (p Plugin) OnAppInit(app launchr.App) error {
	var am action.Manager
	app.GetService(&am)
	am.AddDecorators(pluginPrintInput)
	return nil
}

func pluginPrintInput(_ action.Manager, a *action.Action) {
	if !strings.HasPrefix(a.ID, "test-print-input:") {
		return
	}
	a.SetRuntime(action.NewFnRuntime(func(_ context.Context, a *action.Action) error {
		def := a.ActionDef()
		for _, p := range def.Arguments {
			printParam(p.Name, a.Input().Arg(p.Name), a.Input().IsArgChanged(p.Name))
		}
		for _, p := range def.Options {
			printParam(p.Name, a.Input().Opt(p.Name), a.Input().IsOptChanged(p.Name))
		}
		return nil
	}))
}

func printParam(name string, val any, isChanged bool) {
	fmt.Printf("%s: %v %T %t\n", name, val, val, isChanged)
}
