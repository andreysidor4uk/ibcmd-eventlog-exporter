package logsreader

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/andreysidor4uk/http-gateway-1c/internal/config"
)

type LogsReader struct {
	cfg *config.Config
}

func New(cfg *config.Config) *LogsReader {
	return &LogsReader{
		cfg: cfg,
	}
}

func (logsReader *LogsReader) Start(ctx context.Context, writeChan chan []byte) {
	if _, err := os.Stat(logsReader.cfg.IbcmdPath); err != nil {
		log.Fatal(err.Error())
	}

	if _, err := os.Stat(logsReader.cfg.JournalDir); err != nil {
		log.Fatal(err.Error())
	}

	ticker := time.NewTicker(logsReader.cfg.PauseDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logs, err := logsReader.exportLogs(ctx)
			if err != nil {
				slog.Error(err.Error())
				break
			}
			if len(logs) > 0 {
				writeChan <- logs
			}
			ticker.Reset(logsReader.cfg.PauseDuration)
		}
	}
}

func (logsReader *LogsReader) exportLogs(ctx context.Context) ([]byte, error) {
	tempFile, err := os.CreateTemp("", "logsreader")
	if err != nil {
		return nil, err
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	position, err := os.ReadFile("position")
	if err != nil && os.IsNotExist(err) {
		position = []byte(logsReader.cfg.StartDate.Format(time.RFC3339))
	} else if err != nil {
		return nil, err
	}
	dateFrom, _ := time.Parse(time.RFC3339, string(position))
	dateFrom.Add(time.Second)

	dateTo := time.Now()

	params := []string{
		"eventlog",
		"export",
		"-f",
		"json",
		"--skip-root",
		fmt.Sprintf("--from=%v", dateFormatFor1C(dateFrom)),
		fmt.Sprintf("--to=%v", dateFormatFor1C(dateTo)),
		fmt.Sprintf("--out=\"%v\"", tempFile.Name()),
		fmt.Sprintf("\"%v\"", logsReader.cfg.JournalDir)}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		params = append([]string{
			"-Command",
			`& "` + logsReader.cfg.IbcmdPath + `"`},
			params...)

		cmd = exec.CommandContext(ctx,
			"powershell",
			params...)
	} else {
		cmd = exec.CommandContext(ctx,
			logsReader.cfg.IbcmdPath,
			params...)
	}
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	err = os.WriteFile("position", []byte(dateTo.Format(time.RFC3339)), 0666)
	if err != nil {
		return nil, err
	}

	logs, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func dateFormatFor1C(date time.Time) string {
	return fmt.Sprintf("%vT%v", date.Format(time.DateOnly), date.Format(time.TimeOnly))
}
