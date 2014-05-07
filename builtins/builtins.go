package builtins

import (
	api "github.com/dotcloud/docker/api/server"
	"github.com/dotcloud/docker/daemon/networkdriver/bridge"
	"github.com/dotcloud/docker/engine"
	"github.com/dotcloud/docker/registry"
	"github.com/dotcloud/docker/server"
)

func Register(eng *engine.Engine) error {
	if err := daemon(eng); err != nil {
		return err
	}
	if err := remote(eng); err != nil {
		return err
	}
	return registry.NewService().Install(eng)
}

// remote: a RESTful api for cross-docker communication
func remote(eng *engine.Engine) error {
	return eng.Register("serveapi", api.ServeApi)
}

// daemon: a default execution and storage backend for Docker on Linux,
// with the following underlying components:
//
// * Pluggable storage drivers including aufs, vfs, lvm and btrfs.
// * Pluggable execution drivers including lxc and chroot.
//
// In practice `daemon` still includes most core Docker components, including:
//
// * The reference registry client implementation
// * Image management
// * The build facility
// * Logging
//
// These components should be broken off into plugins of their own.
//
func daemon(eng *engine.Engine) error {
	if err := eng.Register("initserver", server.InitServer); err != nil {
		return err
	}
	return eng.Register("init_networkdriver", bridge.InitDriver)
}
