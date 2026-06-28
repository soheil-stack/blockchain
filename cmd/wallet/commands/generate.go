package commands

import (
	"context"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v3"
)

var generateCommand = &cli.Command{
	Name:  "generate",
	Usage: "Generate new key pair",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			return err
		}

		return crypto.SaveECDSA(getPrivateKeyPath(), privateKey)
	},
}
