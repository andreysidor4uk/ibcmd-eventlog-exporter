package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/andreysidor4uk/http-gateway-1c/internal/config"
	"github.com/andreysidor4uk/http-gateway-1c/internal/logswriter"
)

func main() {
	cfg := config.MustLoad()

	ctx, canel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}

	logsWriter := logswriter.New(cfg)
	go func() {
		wg.Add(1)
		logsWriter.Start(ctx)
		wg.Done()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	canel()
	wg.Wait()
}
