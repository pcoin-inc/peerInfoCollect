package main

import (
	"peerInfoCollect/eth/ethconfig"
	"peerInfoCollect/node"
	"peerInfoCollect/params"
)

type ethstatsConfig struct {
	URL string `toml:",omitempty"`
}
//eth config
type gethConfig struct {
	Eth      ethconfig.Config
	Node     node.Config
	Ethstats ethstatsConfig
	//Metrics  metrics.Config
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit, gitDate)
	cfg.HTTPModules = append(cfg.HTTPModules, "eth")
	return cfg
}

