// Package loglevels has a test plugin that outputs logs.
package loglevels

import (
	"context"

	"github.com/launchrctl/launchr/internal/launchr"
	"github.com/launchrctl/launchr/pkg/action"
)

const actionyaml = `
runtime: plugin
action:
  title: Test Plugin - Log levels
`

func init() {
	launchr.RegisterPlugin(&Plugin{})
}

// Plugin is a test plugin declaration.
type Plugin struct{}

// PluginInfo implements [launchr.Plugin] interface.
func (p *Plugin) PluginInfo() launchr.PluginInfo {
	return launchr.PluginInfo{}
}

// DiscoverActions implements [launchr.ActionDiscoveryPlugin] interface.
func (p *Plugin) DiscoverActions(_ context.Context) ([]*action.Action, error) {
	a := action.NewFromYAML("testplugin:log-levels", []byte(actionyaml))
	a.SetRuntime(action.NewFnRuntime(func(_ context.Context, _ *action.Action) error {
		launchr.Log().Debug("this is DEBUG log")
		launchr.Log().Info("this is INFO log")
		launchr.Log().Warn("this is WARN log")
		launchr.Log().Error("this is ERROR log")
		return nil
	}))

	return []*action.Action{a}, nil
}
