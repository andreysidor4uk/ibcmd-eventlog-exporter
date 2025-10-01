package retentioncontroller

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andreysidor4uk/http-gateway-1c/internal/config"
)

const tickerDelay = time.Hour

type RetentionController struct {
	cfg *config.Config
}

func New(cfg *config.Config) *RetentionController {
	return &RetentionController{
		cfg: cfg,
	}
}

func (retentionController *RetentionController) Start(ctx context.Context) {
	ticker := time.NewTicker(tickerDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			entries, err := os.ReadDir(retentionController.cfg.LogsDir)
			if err != nil {
				slog.Error(err.Error())
				break
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				if !needDeleteFile(entry.Name(), retentionController.cfg.RetentionPeriod) {
					continue
				}

				fullPath := filepath.Join(retentionController.cfg.LogsDir, entry.Name())

				err := os.Remove(fullPath)
				if err != nil {
					slog.Error(fmt.Errorf("retention controller: %w", err).Error())
					continue
				}

				slog.Info(fmt.Sprintf("removed file %s", fullPath))
			}
		}
	}
}

func needDeleteFile(fileName string, retentionPeriod time.Duration) bool {
	nameWithoutExtension := nameWithoutExtension(fileName)
	date, err := time.Parse(time.DateOnly, nameWithoutExtension)
	if err != nil {
		return false
	}

	if retentionPeriod > time.Since(date) {
		return false
	}
	return true
}

func nameWithoutExtension(fileName string) string {
	return strings.Split(fileName, ".")[0]
}
