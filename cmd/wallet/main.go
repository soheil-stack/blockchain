package main

import (
	"log/slog"
	"os"

	"github.com/soheil-stack/blockchain/cmd/wallet/commands"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(log)

	commands.Execute()
}
