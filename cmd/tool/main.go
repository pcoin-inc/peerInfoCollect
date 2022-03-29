package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"os"
	"peerInfoCollect/internal/debug"
)


var (
	app = newApp()
)


func newApp() *cli.App {
	app := cli.NewApp()
	app.Author = "echoWu"
	app.Usage = "peerInfo tool"
	return app
}


func init()  {
	app.Action = tool
	app.Flags = append(app.Flags,debug.Flags...)

	app.Before = func(ctx *cli.Context) error {
		return debug.Setup(ctx)
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()

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

