package client

import (
    "fmt"
)

type DockerCommand func(cli *DockerCli, args ...string) error

type CommandDetails struct {
    Run  DockerCommand
    Name string
    Help string
}

var (
    commands []*CommandDetails
)

func init() {
    /*
       AddCommand(&CommandDetails {
           Name: "help",
           Run: HelpCmd,
           Help: "Lists the command help for Docker",
       })
    */
}

func AddCommand(cmd *CommandDetails) {
    commands = append(commands, cmd)
}

func CommandDetailsFor(cmdStr string) (*CommandDetails, error) {
    var cmd *CommandDetails

    for _, details := range commands {
        if details.Name == cmdStr {
            cmd = details
        }
    }

    if cmd == nil {
        return nil, fmt.Errorf("Unknown command %s.", cmdStr)
    }

    return cmd, nil
}
