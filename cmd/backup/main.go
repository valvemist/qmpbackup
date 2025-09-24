package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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
	fmt.Printf("[%s] %s ", r.Level, r.Message)
	r.Attrs(func(a slog.Attr) bool {
		fmt.Printf("%v", a.Value)
		return true
	})
	// Include file and line number
	src := r.Source()
	if src != nil {
		fmt.Printf(" (%s:%d) ", filepath.Base(src.File), src.Line)
	}

	fmt.Println()
	return nil
}

// WithAttrs returns a new handler with the given attributes.
func (h *customHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

// WithGroup returns a new handler with the given group name.
func (h *customHandler) WithGroup(_ string) slog.Handler { return h }

func parseFlags() (verbose bool, cfg qmpbackup.Config) {
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.BoolVar(&cfg.CleanAll, "clean", false, "Clean existing backup objects")
	flag.StringVar(&cfg.SocketFile, "socket", "", "Path to QMP socket (required)")
	flag.StringVar(&cfg.BackupFile, "backupFile", "", "Backup file base name (required)")
	flag.StringVar(&cfg.DeviceToBackup, "device", "drive0", "Device to backup (default: drive0)")
	flag.IntVar(&cfg.IncLevel, "inc", -1, "Incremental level (-1 means full backup)")
	flag.Parse()
	if cfg.SocketFile == "" || cfg.BackupFile == "" {
		flag.Usage()
		os.Exit(1)
	}
	cfg.NodeTarget = "target0-node" // set default NodeTarget
	return
}

// main is the entry point for the backup CLI tool.
func main() {
	// catch ctrl-c
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	level := slog.LevelInfo
	logger = slog.New(&customHandler{level: level})
	qmpbackup.SetLogger(logger)

	verbose, cfg := parseFlags()

	if verbose {
		level = slog.LevelDebug
	}

	qmpbackup.GenerateBackupFilename(&cfg)

	monitor, err := qmp.NewSocketMonitor("unix", cfg.SocketFile, 2*time.Second)
	if err != nil {
		logger.Error("Failed to connect to QMP", "error", err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	monitor.Connect()
	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
		monitor.Disconnect()
		logger.Info("Program finished.")
	}()
	if cfg.CleanAll {
		CleanAll(monitor, cfg)
		return
	}
	doneCh := make(chan struct{})
	wg.Go(func() {
		qmpbackup.Events(ctx, monitor, func(event qmp.Event) {
			logger.Debug("Event received", event.Event, event.Data)
			handleEvents(event, logger, cfg, func(str string) {
				logger.Info("From callback:", str)
				close(doneCh)
				return
			})
		})
	})

	if _, err := RunBackupWorkflow(monitor, cfg); err != nil {
		logger.Error("RunBackupWorkflow failed", err.Error())
		return
	}
	logger.Debug("Entering select loop")
	for {
		select {
		case <-ctx.Done():
			logger.Info("Context done received in select loop")
			return
		case <-doneCh:
			logger.Info("Done received in select loop")
			return
		case sig := <-sigs:
			fmt.Println("Interrupt received:", sig)
			return
		}
	}
}
func handleEvents(event qmp.Event, logger *slog.Logger, cfg qmpbackup.Config, callback func(string)) {
	str := fmt.Sprintf("%s: %v", event.Event, event.Data)
	switch event.Event {
	case "BLOCK_JOB_ERROR":
		logger.Error(fmt.Sprintf("%v. ", event.Data))
		if strings.Contains(str, "write") {
			logger.Error(fmt.Sprintf("If running full backup, qcow2 image %s must be empty.", cfg.BackupFile))
		}

	case "BLOCK_JOB_COMPLETED":
		logger.Info("Block job completed, sending info to err ch")
		callback("BLOCK_JOB_COMPLETED")
	default:
		logger.Debug(str)
	}

}
