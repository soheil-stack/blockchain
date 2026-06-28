package commands

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v3"
)

var accountCommand = &cli.Command{
	Name:  "account",
	Usage: "Print account for the specific wallet",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		privateKey, err := crypto.LoadECDSA(getPrivateKeyPath())
		if err != nil {
			return err
		}

		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		fmt.Println(address)

		return nil
	},
}
