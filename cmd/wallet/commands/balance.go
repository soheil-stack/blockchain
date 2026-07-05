package commands

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/urfave/cli/v3"
)

var balanceCommand = &cli.Command{
	Name:  "balance",
	Usage: "Print balance for the specific wallet",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		privateKey, err := crypto.LoadECDSA(getPrivateKeyPath())
		if err != nil {
			return err
		}
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		fmt.Println("For Account:", address)

		var account struct {
			Balance uint64 `json:"balance"`
		}
		url := fmt.Sprintf("%s/accounts/%s", nodeURL, address)
		err = core.Send(http.MethodGet, url, nil, &account)
		if err != nil {
			return err
		}

		fmt.Println("Balance:", account.Balance)

		return nil
	},
}
