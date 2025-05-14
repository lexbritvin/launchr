package embedaction

import (
	"context"
	"embed"

	"github.com/launchrctl/launchr/internal/launchr"
	"github.com/launchrctl/launchr/pkg/action"
)

//go:embed action/*
var actionfs embed.FS

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
	// Use subdirectory so the content is available in the root "./".
	a, err := action.NewYAMLFromFS("testplugin.embedaction:container", launchr.MustSubFS(actionfs, "action"))
	if err != nil {
		return nil, err
	}
	return []*action.Action{a}, nil
}
