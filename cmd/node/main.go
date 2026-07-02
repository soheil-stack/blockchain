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
	"github.com/soheil-stack/blockchain/internal/nameservice"
	"github.com/soheil-stack/blockchain/internal/state"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(log)

	if err := run(); err != nil {
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

func run() error {
	cfg := struct {
		Beneficiary       string
		NameServiceFolder string
		SelectStrategy    string
	}{
		Beneficiary:       getEnv("BENEFICIARY", "beneficiary"),
		NameServiceFolder: getEnv("NAME_SERVICE_FOLDER", "zblock/accounts"),
		SelectStrategy:    getEnv("SELECT_STRATEGY", "tip"),
	}

	slog.Info(
		"config loaded",
		"beneficiary", cfg.Beneficiary,
		"name_service_folder", cfg.NameServiceFolder,
		"select_strategy", cfg.SelectStrategy,
	)

	evHandler := func(v string, args ...any) {
		slog.Debug("blockchain event", "msg", fmt.Sprintf(v, args...))
	}

	genesis, err := core.LoadGenesis()
	if err != nil {
		return fmt.Errorf("loading genesis: %w", err)
	}

	slog.Info("genesis loaded")

	beneficiaryPath := fmt.Sprintf("%s/%s.ecdsa", cfg.NameServiceFolder, cfg.Beneficiary)
	beneficiaryPrivateKey, err := crypto.LoadECDSA(beneficiaryPath)
	if err != nil {
		return fmt.Errorf("loading beneficiary private key [%s]: %w", beneficiaryPath, err)
	}

	beneficiaryAddress := crypto.PubkeyToAddress(beneficiaryPrivateKey.PublicKey)
	slog.Info("beneficiary key loaded", "address", beneficiaryAddress.Hex())

	st, err := state.NewState(state.StateConfig{
		Beneficiary:    beneficiaryAddress,
		Genesis:        genesis,
		EvHandler:      evHandler,
		SelectStrategy: cfg.SelectStrategy,
	})
	if err != nil {
		return fmt.Errorf("initializing state: %w", err)
	}
	defer func() {
		slog.Info("shutting down state")
		_ = st.Shutdown()
	}()

	slog.Info("state initialized")

	state.RunWorker(st, evHandler)

	ns, err := nameservice.New(cfg.NameServiceFolder)
	if err != nil {
		return fmt.Errorf("initializing name service: %w", err)
	}
	slog.Info("name service initialized")

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	serverError := make(chan error, 1)

	publicHandler := public.NewServer(st, ns)
	publicServer := http.Server{
		Addr:         ":8080",
		Handler:      publicHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  time.Minute,
	}
	go func() {
		slog.Info("public server starting", "addr", publicServer.Addr)
		if err := publicServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("public server failed", "err", err)
			serverError <- err
		}
	}()

	privateHandler := private.NewServer(st)
	privateServer := http.Server{
		Addr:         ":8081",
		Handler:      privateHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  time.Minute,
	}
	go func() {
		slog.Info("private server starting", "addr", privateServer.Addr)
		if err := privateServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("private server failed", "err", err)
			serverError <- err
		}
	}()

	select {
	case err := <-serverError:
		return fmt.Errorf("server error: %w", err)
	case <-sigint:
		slog.Info("shutdown signal received, draining connections")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := publicServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("could not shutdown public server gracefully: %w", err)
		}
		slog.Info("public server stopped")

		if err := privateServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("could not shutdown private server gracefully: %w", err)
		}
		slog.Info("private server stopped")
	}

	slog.Info("node shutdown complete")
	return nil
}
