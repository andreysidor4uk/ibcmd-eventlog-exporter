package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/andreysidor4uk/http-gateway-1c/internal/config"
	"github.com/andreysidor4uk/http-gateway-1c/internal/logsreader"
	"github.com/andreysidor4uk/http-gateway-1c/internal/logswriter"
)

func main() {
	logFile, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.MustLoad()

	ctx, canel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	logsWriter := logswriter.New(cfg)
	wg.Go(func() {
		logsWriter.Start(ctx)
	})

	logsReader := logsreader.New(cfg)
	wg.Go(func() {
		logsReader.Start(ctx, logsWriter.GetWriteChannel())
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	canel()
	wg.Wait()
}
