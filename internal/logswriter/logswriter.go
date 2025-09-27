package logswriter

import (
	"context"
	"log"
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

func (logWriter *LogsWriter) GetWriteChannel() chan []byte {
	return logWriter.writeChannel
}

func (logsWriter LogsWriter) Start(ctx context.Context) {
	_, err := os.ReadDir(logsWriter.cfg.LogsDir)
	if err != nil {
		err := os.Mkdir(logsWriter.cfg.LogsDir, 0750)
		if err != nil {
			log.Panic(err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-logsWriter.writeChannel:
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
		return err
	}
	defer f.Close()

	if _, err := f.Write(msg); err != nil {
		return err
	}

	if _, err := f.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

func getFilename() string {
	t := time.Now()
	return t.Format("2006-01-02") + ".log"
}
