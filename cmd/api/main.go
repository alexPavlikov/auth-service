package main

import (
	"log/slog"

	"github.com/alexPavlikov/auth-service/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		slog.Error("app failed", "error", err)
		return
	}
}
