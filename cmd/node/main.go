package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/ardanlabs/conf/v3"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/node"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := struct {
		Beneficiary       string `conf:"default:beneficiary"`
		NameServiceFolder string `conf:"default:zblock/accounts"`
		SelectStrategy    string `conf:"default:tip"`
	}{}

	help, err := conf.Parse("NODE", &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}

		return fmt.Errorf("parsing config: %w", err)
	}

	evHandler := func(v string, args ...any) {
		s := fmt.Sprintf(v, args...)
		log.Println(s)
	}

	evHandler("node: starting")

	genesis, err := core.LoadGenesis()
	if err != nil {
		return err
	}

	beneficiaryPath := fmt.Sprintf("%s/%s.ecdsa", cfg.NameServiceFolder, cfg.Beneficiary)
	beneficiaryPrivateKey, err := crypto.LoadECDSA(beneficiaryPath)
	if err != nil {
		return fmt.Errorf("unable to load beneficiary private key: %w", err)
	}

	n, err := node.New(node.Config{
		Beneficiary:    crypto.PubkeyToAddress(beneficiaryPrivateKey.PublicKey),
		Genesis:        genesis,
		EvHandler:      evHandler,
		SelectStrategy: cfg.SelectStrategy,
	})
	if err != nil {
		return err
	}

	defer func() {
		_ = n.Shutdown()
	}()

	return nil
}
