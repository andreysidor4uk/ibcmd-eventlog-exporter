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

	var wg sync.WaitGroup

	logsWriter := logswriter.New(cfg)
	wg.Go(func() {
		logsWriter.Start(ctx)
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	canel()
	wg.Wait()
}
