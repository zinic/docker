package client

import (
	"fmt"
	"io"
	"net"
)

const (
	CONTINUE = 0
	RETURN   = 1
)

type CallDetails struct {
	Method       string
	Path         string
	Data         interface{}
	PassAuthInfo bool
}

type CliRemote interface {
	Call(cli *DockerCli, callDetails *CallDetails) (io.ReadCloser, int, error)
	Dial(cli *DockerCli) (net.Conn, error)
}
type CliRemoteInit func() (CliRemote, error)

type CliPlugin interface {
	Before(cli *DockerCli, cmd string, args []string) (action int, err error)
	After(cli *DockerCli, callError error, cmd string, args []string) (err error)
}
type CliPluginInit func() (CliPlugin, error)

var (
	// All registered remotes
	remotes map[string]CliRemoteInit

	// All registered plugins
	plugins map[string]CliPluginInit
)

func init() {
	remotes = make(map[string]CliRemoteInit)
	plugins = make(map[string]CliPluginInit)
}

func RegisterRemote(name string, initFunc CliRemoteInit) error {
	if _, exists := remotes[name]; exists {
		return fmt.Errorf("Name already registered %s", name)
	}

	remotes[name] = initFunc
	return nil
}

func RegisterPlugin(name string, initFunc CliPluginInit) error {
	if _, exists := plugins[name]; exists {
		return fmt.Errorf("Name already registered %s", name)
	}

	plugins[name] = initFunc
	return nil
}

func findRemote(name string) (CliRemoteInit, error) {
	if initFunc, exists := remotes[name]; exists {
		return initFunc, nil
	}

	return nil, fmt.Errorf("No such remote: %s", name)
}

func findPlugin(name string) (CliPluginInit, error) {
	if initFunc, exists := plugins[name]; exists {
		return initFunc, nil
	}

	return nil, fmt.Errorf("No such plugin: %s", name)
}

func NewCliRemote(name string) (CliRemote, error) {
	if initFunc, err := findRemote(name); err == nil {
		return initFunc()
	} else {
		return nil, err
	}
}

func NewCliPlugin(name string) (CliPlugin, error) {
	if initFunc, err := findPlugin(name); err == nil {
		return initFunc()
	} else {
		return nil, err
	}
}
