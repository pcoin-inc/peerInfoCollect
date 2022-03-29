package main

import (
	"gopkg.in/urfave/cli.v1"
	"peerInfoCollect/params"
	"runtime"
	"fmt"
)

var (
	VersionCheckUrlFlag = cli.StringFlag{
		Name:  "check.url",
		Usage: "URL to use when checking vulnerabilities",
		Value: "https://geth.ethereum.org/docs/vulnerabilities/vulnerabilities.json",
	}
	VersionCheckVersionFlag = cli.StringFlag{
		Name:  "check.version",
		Usage: "Version to check",
		Value: fmt.Sprintf("Geth/v%v/%v-%v/%v",
			params.VersionWithCommit(gitCommit, gitDate),
			runtime.GOOS, runtime.GOARCH, runtime.Version()),
	}
)