package logswriter

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/andreysidor4uk/http-gateway-1c/internal/config"
)

type LogsWriter struct {
	cfg          *config.Config
	writeChannel chan []byte
}

func New(cfg *config.Config) *LogsWriter {
	logsWriter := LogsWriter{
		cfg:          cfg,
		writeChannel: make(chan []byte),
	}

	return &logsWriter
}

func (logWriter *LogsWriter) WriteChannel() chan []byte {
	return logWriter.writeChannel
}

func (logsWriter LogsWriter) Start(ctx context.Context) error {
	err := checkDir(logsWriter.cfg.LogsDir)
	if err != nil {
		return fmt.Errorf("logs did: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-logsWriter.writeChannel:
			if !ok {
				break
			}
			if len(msg) == 0 {
				break
			}
			if err := logsWriter.writeLog(msg); err != nil {
				slog.Error(err.Error())
			}
		}
	}
}

func (logsWriter LogsWriter) writeLog(msg []byte) error {
	filePath := path.Join(logsWriter.cfg.LogsDir, getFilename())

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file for logs writer: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(msg); err != nil {
		return fmt.Errorf("logwriter write logs: %w", err)
	}

	if _, err := f.Write([]byte("\n")); err != nil {
		return fmt.Errorf("logs writer write log: %w", err)
	}

	return nil
}

func getFilename() string {
	t := time.Now()
	return t.Format(time.DateOnly) + ".log"
}

func checkDir(dirPath string) error {
	_, err := os.ReadDir(dirPath)
	if err != nil {
		err := os.Mkdir(dirPath, 0750)
		if err != nil {
			return err
		}
	}

	return nil
}
