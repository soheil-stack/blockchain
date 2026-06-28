package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/urfave/cli/v3"
)

var (
	from  string
	to    string
	value uint64
	tip   uint64
	nonce uint64
)

var sendCommand = &cli.Command{
	Name:  "send",
	Usage: "Send transaction",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "from",
			Usage:       "The account to send from",
			Destination: &from,
		},
		&cli.StringFlag{
			Name:        "to",
			Usage:       "The account to send to",
			Destination: &to,
		},
		&cli.Uint64Flag{
			Name:        "value",
			Usage:       "The value to send",
			Destination: &value,
		},
		&cli.Uint64Flag{
			Name:        "tip",
			Usage:       "The tip to send",
			Destination: &tip,
		},
		&cli.Uint64Flag{
			Name:        "nonce",
			Usage:       "Transaction nonce",
			Destination: &nonce,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		privateKey, err := crypto.LoadECDSA(getPrivateKeyPath())
		if err != nil {
			return err
		}

		fromAddress := common.HexToAddress(from)
		toAddress := common.HexToAddress(to)

		// TODO: do not hardcode chainID and data
		chainID := uint64(1)
		data := []byte{}

		tx := core.NewTransaction(chainID, nonce, fromAddress, toAddress, value, tip, data)

		err = tx.Sign(privateKey)
		if err != nil {
			return err
		}

		txBytes, err := json.Marshal(tx)
		if err != nil {
			return err
		}

		response, err := http.Post(fmt.Sprintf("%s/transactions", nodeURL), "application/json", bytes.NewBuffer(txBytes))
		if err != nil {
			return err
		}
		defer func() {
			_ = response.Body.Close()
		}()

		return nil
	},
}
