package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/digitalocean/go-qemu/qmp"
	qmpbackup "github.com/valvemist/qmpbackup/backup"
)

var logger *slog.Logger

type customHandler struct {
	level slog.Leveler
}

// Enabled determines whether the customHandler should log messages at the given level.
func (h *customHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.level.Level()
}

// Handle processes a log record using the customHandler.
func (h *customHandler) Handle(_ context.Context, r slog.Record) error {
	fmt.Printf("[%s] %s", r.Level, r.Message)
	r.Attrs(func(a slog.Attr) bool {
		fmt.Printf("%v", a.Value)
		return true
	})

	fmt.Println()
	return nil
}

// WithAttrs returns a new handler with the given attributes.
func (h *customHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

// WithGroup returns a new handler with the given group name.
func (h *customHandler) WithGroup(_ string) slog.Handler { return h }

// main is the entry point for the backup CLI tool.
func main() {
	var verbose bool
	var cfg qmpbackup.Config
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.BoolVar(&cfg.CleanBitmap, "cleanBitmap", false, "Clean existing bitmap and exit")
	flag.StringVar(&cfg.SocketFile, "socket", "", "Path to QMP socket (required)")
	flag.StringVar(&cfg.BackupFile, "backupFile", "", "Backup file base name (required)")
	flag.StringVar(&cfg.DeviceToBackup, "device", "drive0", "Device to backup (default: drive0)")
	flag.IntVar(&cfg.IncLevel, "inc", -1, "Incremental level (-1 means full backup)")
	flag.Parse()

	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	logger = slog.New(&customHandler{level: level})
	qmpbackup.SetLogger(logger)

	if cfg.SocketFile == "" || cfg.BackupFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg.NodeTarget = "target0-node"
	qmpbackup.GenerateBackupFilename(&cfg)

	// Connect to QMP
	monitor, err := qmp.NewSocketMonitor("unix", cfg.SocketFile, 2*time.Second)
	if err != nil {
		logger.Error("Failed to connect to QMP", "error", err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	monitor.Connect()
	defer monitor.Disconnect()

	// Cleanup on exit
	defer func() {
		logger.Info("Cleaning up...")
		if _, err := qmpbackup.RunBlockDevDel(monitor, cfg); err != nil {
			logger.Warn("Failed to delete node", "error", err)
		}
		logger.Info("Program finished.")
	}()

	// Start event listener. event holds Data map[string]interface{}
	qmpbackup.Events(ctx, monitor, cancel, func(event qmp.Event) {
		handleEvents(event, logger, cfg)
	})

	// Run backup workflow
	if _, err := RunBackupWorkflow(monitor, cfg); err != nil {
		logger.Error("Workflow failed", "error", err)
		return
	}

	<-ctx.Done()
}

// handleEvents processes QEMU events and logs relevant information during backup.
func handleEvents(event qmp.Event, logger *slog.Logger, cfg qmpbackup.Config) {
	str := fmt.Sprintf("%s: %v", event.Event, event.Data)
	if event.Event == "BLOCK_JOB_ERROR" {
		logger.Warn(str)
	} else if strings.Contains(str, "Input/output error") {
		logger.Error(fmt.Sprintf("Couldn't write to image %v. If running full backup, qcow2 image must be empty", cfg.BackupFile))
	} else {
		logger.Debug(str)
	}
	if val, ok := event.Data["type"]; ok {
		strVal := fmt.Sprintf("%v", val)
		logger.Debug("Completed job type is " + strVal)
	}
}
