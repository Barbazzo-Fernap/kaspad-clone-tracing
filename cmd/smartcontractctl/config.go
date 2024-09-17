package main

import (
	"os"
	"path/filepath"

	"github.com/bugnanetwork/bugnad/infrastructure/config"
	"github.com/bugnanetwork/bugnad/util"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

const (
	defaultRPCServer      = "localhost"
	defaultLogFilename    = "bugnasc.log"
	defaultErrLogFilename = "bugnasc_err.log"
)

const (
	deploySubCmd = "deploy"
)

var (
	defaultAppDir     = util.AppDir("bugnasc", false)
	defaultLogFile    = filepath.Join(defaultAppDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(defaultAppDir, defaultErrLogFilename)
)

type configFlags struct {
	RPCServer  string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	PrivateKey string `short:"p" long:"privatekey" description:"Private key to use for signing"`
	config.NetworkFlags
}

type deployConfig struct {
	ContractCode string `short:"c" long:"contractcode" description:"Path to the contract code"`
	configFlags
}

func parseCommandLine() (subCommand string, config interface{}) {
	cfg := &configFlags{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)

	initLog(defaultLogFile, defaultErrLogFile)

	deployConf := &deployConfig{}
	parser.AddCommand(deploySubCmd, "Deploy a contract", "Deploys a contract to the network", deployConf)

	_, err := parser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
		return "", nil
	}

	switch parser.Command.Active.Name {
	case deploySubCmd:
		combineNetworkFlags(&deployConf.NetworkFlags, &cfg.NetworkFlags)
		err := deployConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err.Error())
		}
		validateDeployConfig(deployConf)
		config = deployConf
	}

	return parser.Command.Active.Name, config
}

func combineNetworkFlags(dst, src *config.NetworkFlags) {
	dst.Testnet = dst.Testnet || src.Testnet
	dst.Simnet = dst.Simnet || src.Simnet
	dst.Devnet = dst.Devnet || src.Devnet
	if dst.OverrideDAGParamsFile == "" {
		dst.OverrideDAGParamsFile = src.OverrideDAGParamsFile
	}
}

func validateDeployConfig(cfg *deployConfig) {
	if cfg.ContractCode == "" {
		printErrorAndExit("contract code is required")
	}

	if cfg.PrivateKey == "" {
		printErrorAndExit("private key is required")
	}
}
