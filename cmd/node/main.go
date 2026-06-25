package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/soheil-stack/blockchain/cmd/node/handlers/private"
	"github.com/soheil-stack/blockchain/cmd/node/handlers/public"
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/state"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	if err := run(log); err != nil {
		slog.Error("node terminated", "err", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func run(log *slog.Logger) error {
	cfg := struct {
		Beneficiary       string
		NameServiceFolder string
		SelectStrategy    string
	}{
		Beneficiary:       getEnv("BENEFICIARY", "beneficiary"),
		NameServiceFolder: getEnv("NAME_SERVICE_FOLDER", "zblock/accounts"),
		SelectStrategy:    getEnv("SELECT_STRATEGY", "tip"),
	}

	log.Info(
		"config loaded",
		"beneficiary", cfg.Beneficiary,
		"name_service_folder", cfg.NameServiceFolder,
		"select_strategy", cfg.SelectStrategy,
	)

	evHandler := func(v string, args ...any) {
		log.Debug("blockchain event", "msg", fmt.Sprintf(v, args...))
	}

	genesis, err := core.LoadGenesis()
	if err != nil {
		return fmt.Errorf("loading genesis: %w", err)
	}

	log.Info("genesis loaded")

	beneficiaryPath := fmt.Sprintf("%s/%s.ecdsa", cfg.NameServiceFolder, cfg.Beneficiary)
	beneficiaryPrivateKey, err := crypto.LoadECDSA(beneficiaryPath)
	if err != nil {
		return fmt.Errorf("loading beneficiary private key [%s]: %w", beneficiaryPath, err)
	}

	beneficiaryAddress := crypto.PubkeyToAddress(beneficiaryPrivateKey.PublicKey)
	log.Info("beneficiary key loaded", "address", beneficiaryAddress.Hex())

	state, err := state.NewState(state.StateConfig{
		Beneficiary:    beneficiaryAddress,
		Genesis:        genesis,
		EvHandler:      evHandler,
		SelectStrategy: cfg.SelectStrategy,
	})
	if err != nil {
		return fmt.Errorf("initializing state: %w", err)
	}
	defer func() {
		log.Info("shutting down state")
		_ = state.Shutdown()
	}()

	log.Info("state initialized")

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
		log.Info("public server starting", "addr", publicServer.Addr)
		if err := publicServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("public server failed", "err", err)
			serverError <- err
		}
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
		log.Info("private server starting", "addr", privateServer.Addr)
		if err := privateServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("private server failed", "err", err)
			serverError <- err
		}
	}()

	select {
	case err := <-serverError:
		return fmt.Errorf("server error: %w", err)
	case <-sigint:
		log.Info("shutdown signal received, draining connections")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := publicServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("could not shutdown public server gracefully: %w", err)
		}
		log.Info("public server stopped")

		if err := privateServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("could not shutdown private server gracefully: %w", err)
		}
		log.Info("private server stopped")
	}

	log.Info("node shutdown complete")
	return nil
}
