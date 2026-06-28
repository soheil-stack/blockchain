// Package commands
package commands

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

var (
	accountPath string
	accountName string
	nodeURL     string
)

var cmd = &cli.Command{
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "account-path",
			Value:       "zblock/accounts",
			Usage:       "Path to the directory with private keys",
			Destination: &accountPath,
		},
		&cli.StringFlag{
			Name:        "account",
			Value:       "soheil.ecdsa",
			Usage:       "The account to use",
			Destination: &accountName,
		},
		&cli.StringFlag{
			Name:        "node-url",
			Value:       "http://localhost:8080",
			Usage:       "The node url",
			Destination: &nodeURL,
		},
	},
	Commands: []*cli.Command{
		generateCommand,
		accountCommand,
		balanceCommand,
		sendCommand,
	},
}

func Execute() {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("running wallet command", "err", err)
		os.Exit(1)
	}
}

func getPrivateKeyPath() string {
	if !strings.HasSuffix(accountName, ".ecdsa") {
		accountName += ".ecdsa"
	}

	return filepath.Join(accountPath, accountName)
}
