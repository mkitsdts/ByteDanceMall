package main

import (
	"bytedancemall/gateway/app"
	"log/slog"
)

func main() {
	server := app.NewServer()
	if err := server.Run(); err != nil {
		slog.Error(err.Error())
	}
}
