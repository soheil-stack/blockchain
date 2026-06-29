package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
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

		response, err := http.Get(fmt.Sprintf("%s/accounts/%s", nodeURL, address))
		if err != nil {
			return err
		}
		defer func() {
			_ = response.Body.Close()
		}()

		decoder := json.NewDecoder(response.Body)
		var account struct {
			Balance uint64 `json:"balance"`
		}
		err = decoder.Decode(&account)
		if err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}

		fmt.Println("Balance:", account.Balance)

		return nil
	},
}
