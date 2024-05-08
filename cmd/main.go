package main

import (
	"context"
	_ "embed"
	"fmt"
	"go/build"
	"os"
	"time"

	"github.com/ava-labs/avalanche-network-runner/local"
	"github.com/ava-labs/avalanche-network-runner/network"
	"github.com/ava-labs/avalanchego/config"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/urfave/cli/v2"

	"go.uber.org/zap"

	"github.com/consideritdone/landslide-runner/internal"
)

const (
	healthyTimeout = 2 * time.Minute
	subnetFileName = "pjSL9ksard4YE96omaiTkGL5H6XX2W5VEo3ZgWC9S2P6gzs9A"
)

var (
	goPath = os.ExpandEnv("$GOPATH")

	//go:embed data/genesis.json
	genesis []byte
)

func main() {
	// Create the logger
	logFactory := logging.NewFactory(logging.Config{
		DisplayLevel: logging.Info,
		LogLevel:     logging.Info,
	})
	log, err := logFactory.Make("main")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if goPath == "" {
		goPath = build.Default.GOPATH
	}

	binaryPath := "/tmp/e2e-test-landslide/avalanchego"
	workDir := "/tmp/e2e-test-landslide/nodes"

	os.RemoveAll(workDir)
	os.MkdirAll(workDir, 0777)

	app := &cli.App{
		Name:  "main",
		Usage: "runNodes landslidevm tests",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "spin up network and deploy landslidevm as a subnet",
				Action: func(cCtx *cli.Context) error {
					nw, err := createNetwork(log, binaryPath, workDir)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					_, err = runNodes(log, binaryPath, nw)
					if err != nil {
						log.Fatal("error starting nodes", zap.Error(err))
						return cli.Exit("exiting", 1)
					}

					internal.GracefulShutdown(nw, log)
					return nil
				},
			},
			{
				Name:  "e2e",
				Usage: "spin up landslide subnet and run end-to-end tests",
				Subcommands: []*cli.Command{
					{
						Name:  "kvstore",
						Usage: "kvstore end-to-end tests",
						Action: func(cCtx *cli.Context) error {
							log.Info("not implemented yet")
							return nil
						},
					},
					{
						Name:  "wasm",
						Usage: "wasm end-to-end tests",
						Action: func(cCtx *cli.Context) error {
							log.Info("not implemented yet")
							return nil
						},
					},
					{
						Name:  "osmosis",
						Usage: "osmosis end-to-end tests",
						Action: func(cCtx *cli.Context) error {
							log.Info("not implemented yet")
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal("fatal error", zap.Error(err))
		os.Exit(1)
	}
}

func runNodes(log logging.Logger, binaryPath string, nw network.Network) ([]string, error) {
	// Wait until the nodes in the network are ready
	if err := internal.Await(nw, log, healthyTimeout); err != nil {
		return nil, err
	}

	// Add some chain
	nodeNames, err := nw.GetNodeNames()
	if err != nil {
		return nil, err
	}

	for i := range nodeNames {
		node, err := nw.GetNode(nodeNames[i])
		if err != nil {
			return nil, err
		}
		if _, err := internal.Copy(
			fmt.Sprintf("%s/plugins/%s", binaryPath, subnetFileName),
			fmt.Sprintf("%s/plugins/%s", node.GetDataDir(), subnetFileName),
		); err != nil {
			return nil, err
		}
	}

	chains, err := nw.CreateBlockchains(context.Background(), []network.BlockchainSpec{
		{
			VMName:      "landslidevm",
			Genesis:     genesis,
			ChainConfig: []byte(`{"warp-api-enabled": true}`),
			SubnetSpec: &network.SubnetSpec{
				SubnetConfig: nil,
				Participants: nodeNames,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Wait until the nodes in the network are ready
	if err := internal.Await(nw, log, healthyTimeout); err != nil {
		return nil, err
	}

	rpcUrls := make([]string, len(nodeNames))
	for i := range nodeNames {
		node, err := nw.GetNode(nodeNames[i])
		if err != nil {
			return nil, err
		}
		rpcUrls[i] = fmt.Sprintf("http://127.0.0.1:%d/ext/bc/%s/rpc", node.GetAPIPort(), chains[0])
		log.Info("subnet rpc url", zap.String("node", nodeNames[i]), zap.String("url", rpcUrls[i]))
	}

	return rpcUrls, nil
}

func createNetwork(log logging.Logger, binaryPath string, workDir string) (network.Network, error) {
	nwConfig, err := local.NewDefaultConfig(fmt.Sprintf("%s/avalanchego", binaryPath))
	if err != nil {
		return nil, err
	}

	nwConfig.Flags["log-level"] = "INFO"

	// adjust avalanche port to not conflict with default ports 9650
	for _, cfg := range nwConfig.NodeConfigs {
		httpPort := cfg.Flags[config.HTTPPortKey].(int)
		stakingPort := cfg.Flags[config.StakingPortKey].(int)

		cfg.Flags[config.HTTPPortKey] = httpPort + 100
		cfg.Flags[config.StakingPortKey] = stakingPort + 100
	}

	nw, err := local.NewNetwork(log, nwConfig, workDir, "", true, false, true)
	if err != nil {
		return nil, err
	}

	return nw, err
}
