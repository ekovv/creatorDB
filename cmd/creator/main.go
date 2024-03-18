package main

import (
	"creatorDB/internal/config"
	"fmt"
	"log/slog"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)

}

func setupLogger(env string) *slog.Logger {

}
