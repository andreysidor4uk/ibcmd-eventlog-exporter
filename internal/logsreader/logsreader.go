package logsreader

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/andreysidor4uk/http-gateway-1c/internal/config"
)

const (
	tempFilenamePattern = "logsreader"
	positionFileName    = "position"
)

var basicIbcmdParams = []string{"eventlog", "export", "-f", "json", "--skip-root"}

type LogsReader struct {
	cfg *config.Config
}

func New(cfg *config.Config) *LogsReader {
	return &LogsReader{
		cfg: cfg,
	}
}

func (logsReader *LogsReader) Start(ctx context.Context, writeChan chan []byte) error {
	if _, err := os.Stat(logsReader.cfg.IbcmdPath); err != nil {
		return fmt.Errorf("ibcmd path: %w", err)
	}

	if _, err := os.Stat(logsReader.cfg.JournalDir); err != nil {
		return fmt.Errorf("journal dir: %w", err)
	}

	ticker := time.NewTicker(logsReader.cfg.PauseDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			logs, err := logsReader.extractLogs(ctx)
			if err != nil {
				slog.Error(fmt.Errorf("export logs: %w", err).Error())
				break
			}

			writeChan <- logs
		}
	}
}

func (logsReader *LogsReader) extractLogs(ctx context.Context) ([]byte, error) {
	tempFileName, err := tempFileName()
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFileName)

	dateFrom, err := logsReader.loadPosition()
	if err != nil {
		return nil, fmt.Errorf("load position: %w", err)
	}
	dateFrom = dateFrom.Add(time.Second)

	dateTo := time.Now()

	if err := logsReader.runIbcmdExportLogs(ctx, tempFileName, dateFrom, dateTo); err != nil {
		return nil, fmt.Errorf("ibcmd exec export logs: %w", err)
	}

	if err := logsReader.savePosition(dateTo); err != nil {
		return nil, fmt.Errorf("save position: %w", err)
	}

	logs, err := os.ReadFile(tempFileName)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (logsReader *LogsReader) loadPosition() (time.Time, error) {
	position, err := os.ReadFile(positionFileName)
	if err != nil && os.IsNotExist(err) {
		position = []byte(logsReader.cfg.StartDate.Format(time.RFC3339))
	} else if err != nil {
		return time.Time{}, err
	}

	date, err := time.Parse(time.RFC3339, string(position))
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}

func (logsReader *LogsReader) savePosition(position time.Time) error {
	return os.WriteFile(positionFileName, []byte(position.Format(time.RFC3339)), 0666)
}

func (logsReader *LogsReader) runIbcmdExportLogs(ctx context.Context, tempFileName string, dateFrom time.Time, dateTo time.Time) error {
	params := append(basicIbcmdParams,
		fmt.Sprintf("--from=%v", dateFormatFor1C(dateFrom)),
		fmt.Sprintf("--to=%v", dateFormatFor1C(dateTo)),
		fmt.Sprintf("--out=\"%v\"", tempFileName),
		fmt.Sprintf("\"%v\"", logsReader.cfg.JournalDir))

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		params = append([]string{
			"-Command",
			fmt.Sprintf(`& "%s"`, logsReader.cfg.IbcmdPath)},
			params...)

		cmd = exec.CommandContext(ctx,
			"powershell",
			params...)
	} else {
		cmd = exec.CommandContext(ctx,
			logsReader.cfg.IbcmdPath,
			params...)
	}

	return cmd.Run()
}

func tempFileName() (string, error) {
	tempFile, err := os.CreateTemp("", tempFilenamePattern)
	if err != nil {
		return "", err
	}
	tempFile.Close()

	return tempFile.Name(), nil
}

func dateFormatFor1C(date time.Time) string {
	return fmt.Sprintf("%vT%v", date.Format(time.DateOnly), date.Format(time.TimeOnly))
}
