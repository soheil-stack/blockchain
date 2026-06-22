package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/soheil-stack/blockchain/internal/core"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	return nil
}

func createPrivateKey() error {
	privatekey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}

	err = crypto.SaveECDSA("jane.ecdsa", privatekey)
	if err != nil {
		return err
	}

	fmt.Println(crypto.PubkeyToAddress(privatekey.PublicKey))

	return nil
}

func verifyTransaction() error {
	prv1, err := crypto.LoadECDSA("scratch/accounts/soheil.ecdsa")
	if err != nil {
		return err
	}
	from := crypto.PubkeyToAddress(prv1.PublicKey)

	prv2, err := crypto.LoadECDSA("scratch/accounts/atiyeh.ecdsa")
	if err != nil {
		return err
	}
	to := crypto.PubkeyToAddress(prv2.PublicKey)

	tx := core.NewTransaction(1, 0, from, to, 10, 0, nil)
	err = tx.Sign(prv1)
	if err != nil {
		return err
	}

	verified, err := tx.Verify(1)
	if err != nil {
		return err
	}

	fmt.Println("transaction verified:", verified)

	return nil
}
