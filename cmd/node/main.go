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
	"github.com/soheil-stack/blockchain/internal/core"
	"github.com/soheil-stack/blockchain/internal/nameservice"
	"github.com/soheil-stack/blockchain/internal/server"
	"github.com/soheil-stack/blockchain/internal/state"
	"github.com/soheil-stack/blockchain/internal/storage"
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
		DBPath            string
		SelectStrategy    string
		Host              string
	}{
		Beneficiary:       getEnv("BENEFICIARY", "beneficiary"),
		NameServiceFolder: getEnv("NAME_SERVICE_FOLDER", "zblock/accounts"),
		DBPath:            getEnv("DB_PATH", "zblock/miner"),
		SelectStrategy:    getEnv("SELECT_STRATEGY", "tip"),
		Host:              getEnv("HOST", "0.0.0.0:8080"),
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

	diskStorage, err := storage.NewDisk(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}

	// TODO: do not hardcode origin peers
	originPeers := []string{"0.0.0.0:8080"}

	peerSet := core.NewPeerSet()
	for _, host := range originPeers {
		peerSet.Add(core.NewPeer(host))
	}
	peerSet.Add(core.NewPeer(cfg.Host))

	st, err := state.NewState(state.StateConfig{
		Beneficiary:    beneficiaryAddress,
		Genesis:        genesis,
		EvHandler:      evHandler,
		SelectStrategy: cfg.SelectStrategy,
		Storage:        diskStorage,
		KnownPeers:     peerSet,
		Host:           cfg.Host,
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

	serverHandler := server.New(st, ns)
	server := http.Server{
		Addr:         cfg.Host,
		Handler:      serverHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  time.Minute,
	}
	go func() {
		slog.Info("http server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", "err", err)
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

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("could not shutdown http server gracefully: %w", err)
		}
		slog.Info("http server stopped")
	}

	slog.Info("node shutdown complete")
	return nil
}
