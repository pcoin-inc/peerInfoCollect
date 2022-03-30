package main

import (
	"peerInfoCollect/accounts/external"
	"peerInfoCollect/cmd/utils"
	"peerInfoCollect/eth/ethconfig"
	"peerInfoCollect/node"
	"peerInfoCollect/params"
	"gopkg.in/urfave/cli.v1"
	"peerInfoCollect/metrics"
	"peerInfoCollect/log"
	"peerInfoCollect/accounts/keystore"
	"fmt"
)

type ethstatsConfig struct {
	URL string `toml:",omitempty"`
}
//eth config
type gethConfig struct {
	Eth      ethconfig.Config
	Node     node.Config
	Ethstats ethstatsConfig
	Metrics  metrics.Config
}


func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit, gitDate)
	cfg.HTTPModules = append(cfg.HTTPModules, "eth")
	return cfg
}

// makeConfigNode loads geth configuration and creates a blank node instance.
func makeConfigNode(ctx *cli.Context) (*node.Node, gethConfig) {
	// Load defaults.
	cfg := gethConfig{
		Eth:     ethconfig.Defaults,
		Node:    defaultNodeConfig(),
		Metrics: metrics.DefaultConfig,
	}

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	// Node doesn't by default populate account manager backends
	if err := setAccountManagerBackends(stack); err != nil {
		utils.Fatalf("Failed to set account manager backends: %v", err)
	}

	utils.SetEthConfig(ctx, stack, &cfg.Eth)
	if ctx.GlobalIsSet(utils.EthStatsURLFlag.Name) {
		cfg.Ethstats.URL = ctx.GlobalString(utils.EthStatsURLFlag.Name)
	}
	//applyMetricConfig(ctx, &cfg)

	return stack, cfg
}


func setAccountManagerBackends(stack *node.Node) error {
	conf := stack.Config()
	am := stack.AccountManager()
	keydir := stack.KeyStoreDir()
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP


	// Assemble the supported backends
	if len(conf.ExternalSigner) > 0 {
		log.Info("Using external signer", "url", conf.ExternalSigner)
		if extapi, err := external.NewExternalBackend(conf.ExternalSigner); err == nil {
			am.AddBackend(extapi)
			return nil
		} else {
			return fmt.Errorf("error connecting to external signer: %v", err)
		}
	}

	// For now, we're using EITHER external signer OR local signers.
	// If/when we implement some form of lockfile for USB and keystore wallets,
	// we can have both, but it's very confusing for the user to see the same
	// accounts in both externally and locally, plus very racey.
	am.AddBackend(keystore.NewKeyStore(keydir, scryptN, scryptP))

	return nil
}
