package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"os"
	"peerInfoCollect/cmd/utils"
	"peerInfoCollect/console/prompt"
	"peerInfoCollect/internal/debug"
	"peerInfoCollect/params"
)

const (
	clientIdentifier = "geth" // Client identifier to advertise over the network
)

var (
	gitCommit = ""
	gitDate   = ""
	app = newApp(gitCommit,gitDate)
	nodeFlags = []cli.Flag{
		utils.PasswordFileFlag,
		utils.UnlockedAccountFlag,
		utils.DataDirFlag,
	}
)


func newApp(gitCommit, gitDate string) *cli.App {
	app := cli.NewApp()
	app.Author = "echoWu"
	app.Version = params.VersionWithCommit(gitCommit, gitDate)
	app.Usage = "peerInfo tool"
	return app
}


func init()  {
	app.Action = tool
	app.Commands = []cli.Command{
		accountCommand,
	}
	app.Flags = append(app.Flags,nodeFlags...)
	app.Flags = append(app.Flags,debug.Flags...)
	app.Before = func(ctx *cli.Context) error {
		return debug.Setup(ctx)
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		prompt.Stdin.Close()
		return nil
	}
}

func main()  {
	if err := app.Run(os.Args);err!= nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func tool()  {

}

