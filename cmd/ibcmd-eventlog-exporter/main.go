package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/andreysidor4uk/http-gateway-1c/internal/config"
	"github.com/andreysidor4uk/http-gateway-1c/internal/logsreader"
	"github.com/andreysidor4uk/http-gateway-1c/internal/logswriter"
	"github.com/andreysidor4uk/http-gateway-1c/internal/retentioncontroller"
)

const logFileName = "logs.log"

func main() {
	cfg := config.MustLoad()

	logFile, err := initFileLogger()
	if err != nil {
		exit(fmt.Errorf("init file logger: %w", err))
	}
	defer logFile.Close()

	ctx, canel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	logsWriter := logswriter.New(cfg)
	wg.Go(func() {
		err := logsWriter.Start(ctx)
		if err != nil {
			exit(fmt.Errorf("start logs writer: %w", err))
		}
	})

	logsReader := logsreader.New(cfg)
	wg.Go(func() {
		err := logsReader.Start(ctx, logsWriter.WriteChannel())
		if err != nil {
			exit(fmt.Errorf("start logs reader: %w", err))
		}
	})

	retentionController := retentioncontroller.New(cfg)
	wg.Go(func() {
		retentionController.Start(ctx)
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	canel()
	wg.Wait()
}

func exit(err error) {
	slog.Error(err.Error())
	os.Exit(1)
}

func initFileLogger() (*os.File, error) {
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	return logFile, nil
}
