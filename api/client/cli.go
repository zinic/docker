package client

import (
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "reflect"
    "strings"
    "text/template"

    flag "github.com/dotcloud/docker/pkg/mflag"
    "github.com/dotcloud/docker/pkg/term"
    "github.com/dotcloud/docker/registry"
    "github.com/dotcloud/docker/utils"
)

var funcMap = template.FuncMap{
    "json": func(v interface{}) string {
        a, _ := json.Marshal(v)
        return string(a)
    },
}

func (cli *DockerCli) getMethod(name string) (func(...string) error, bool) {
    methodName := "Cmd" + strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
    method := reflect.ValueOf(cli).MethodByName(methodName)

    if !method.IsValid() {
        return nil, false
    }

    return method.Interface().(func(...string) error), true
}

func (cli *DockerCli) ParseCommands(args ...string) error {
    if len(args) > 0 {
        method, exists := cli.getMethod(args[0])

        if !exists {
            fmt.Println("Error: Command not found:", args[0])
            return cli.CmdHelp(args[1:]...)
        }

        // Hook for calling the plugins before a command method is executed
        for name, plugin := range cli.Plugins {
            utils.Debugf("Calling Plugin(%s).Before(...)", name)

            action, err := plugin.Before(cli, args[0], args[1:])

            switch action {
            case PLUGIN_EXIT_CMD:
                return err
            }
        }

        callErr := method(args[1:]...)

        // Hook for calling the plugins after a command method is executed
        for name, plugin := range cli.Plugins {
            utils.Debugf("Calling Plugin(%s).After(...)", name)

            if err := plugin.After(cli, callErr, args[0], args[1:]); err != nil {
                utils.Errorf("Plugin(%s) failed execution: %s.", name, err)
            }
        }

        return callErr
    }

    return cli.CmdHelp(args...)
}

func (cli *DockerCli) Subcmd(name, signature, description string) *flag.FlagSet {
    flags := flag.NewFlagSet(name, flag.ContinueOnError)
    flags.Usage = func() {
        fmt.Fprintf(cli.Err, "\nUsage: docker %s %s\n\n%s\n\n", name, signature, description)
        flags.PrintDefaults()
        os.Exit(2)
    }
    return flags
}

func (cli *DockerCli) LoadConfigFile() (*registry.ConfigFile, error) {
    userHome := os.Getenv("HOME")

    if configFile, err := registry.LoadConfig(userHome); err == nil {
        return configFile, nil
    } else {
        return nil, err
    }
}

func NewDockerCli(in io.ReadCloser, out, err io.Writer, proto, addr string, tlsConfig *tls.Config) *DockerCli {
    var (
        isTerminal = false
        terminalFd uintptr
        cliRemote  CliRemote
        cliPlugins = make(map[string]CliPlugin)
        scheme     = "http"
    )

    if tlsConfig != nil {
        scheme = "https"
    }

    if in != nil {
        if file, ok := in.(*os.File); ok {
            terminalFd = file.Fd()
            isTerminal = term.IsTerminal(terminalFd)
        }
    }

    if err == nil {
        err = out
    }

    // Figure out which remote should be active, if any.
    if cliRemoteEnv := os.Getenv("DOCKER_CLI_REMOTE"); cliRemoteEnv != "" {
        cliRemoteName := strings.Trim(cliRemoteEnv, " \t\r\n")

        if remoteInstance, err := NewCliRemote(cliRemoteName); err == nil {
            cliRemote = remoteInstance
        } else {
            utils.Errorf("Unable to init CLI remote: %s", cliRemoteName)
        }
    } else {
        // Add the default remote if no remote was specified by the user.
        if remoteInstance, err := NewCliRemote("http"); err == nil {
            cliRemote = remoteInstance
        } else {
            utils.Errorf("Unable to init the default CLI remote: http.")
        }
    }

    if cliPluginEnv := os.Getenv("DOCKER_CLI_PLUGINS"); cliPluginEnv != "" {
        for _, str := range strings.Split(cliPluginEnv, ",") {
            cliPluginName := strings.Trim(str, " \t\r\n")

            if pluginInstance, err := NewCliPlugin(cliPluginName); err == nil {
                cliPlugins[cliPluginName] = pluginInstance
            } else {
                utils.Errorf("Unable to init CLI plugin: %s", cliPluginName)
            }
        }
    }

    return &DockerCli{
        Proto:      proto,
        Address:    addr,
        Remote:     cliRemote,
        Plugins:    cliPlugins,
        In:         in,
        Out:        out,
        Err:        err,
        IsTerminal: isTerminal,
        TerminalFd: terminalFd,
        TlsConfig:  tlsConfig,
        Scheme:     scheme,
    }
}

type DockerCli struct {
    Proto      string
    Address    string
    Remote     CliRemote
    Plugins    map[string]CliPlugin
    ConfigFile *registry.ConfigFile
    In         io.ReadCloser
    Out        io.Writer
    Err        io.Writer
    IsTerminal bool
    TerminalFd uintptr
    TlsConfig  *tls.Config
    Scheme     string
}
