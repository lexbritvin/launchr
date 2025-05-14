package registeredfs

import (
	"embed"

	"github.com/launchrctl/launchr/internal/launchr"
	"github.com/launchrctl/launchr/pkg/action"
)

//go:embed test-embed-fs/*
var actionsfs embed.FS

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
func (p *Plugin) OnAppInit(app launchr.App) error {
	// Add custom fs to default discovery.
	app.RegisterFS(action.NewDiscoveryFS(actionsfs, app.GetWD()))
	return nil
}
