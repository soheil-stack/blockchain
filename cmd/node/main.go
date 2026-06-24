package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/soheil-stack/blockchain/cmd/node/handlers/private"
	"github.com/soheil-stack/blockchain/cmd/node/handlers/public"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/state"
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

	state, err := state.NewState(state.StateConfig{
		Beneficiary:    crypto.PubkeyToAddress(beneficiaryPrivateKey.PublicKey),
		Genesis:        genesis,
		EvHandler:      evHandler,
		SelectStrategy: cfg.SelectStrategy,
	})
	if err != nil {
		return err
	}
	defer state.Shutdown()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	serverError := make(chan error, 1)

	publicHandler := public.NewServer(state)
	publicServer := http.Server{
		Addr:         ":8080",
		Handler:      publicHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  time.Minute,
	}
	go func() {
		serverError <- publicServer.ListenAndServe()
	}()

	privateHandler := private.NewServer(state)
	privateServer := http.Server{
		Addr:         ":8081",
		Handler:      privateHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  time.Minute,
	}
	go func() {
		serverError <- privateServer.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		return fmt.Errorf("server error: %w", err)
	case <-sigint:
		if err := publicServer.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("could not shutdown public server gracefully: %w", err)
		}

		if err := privateServer.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("could not shutdown private server gracefully: %w", err)
		}
	}

	return nil
}
