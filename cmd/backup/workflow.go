// workflow.go contains CLI-specific orchestration logic for backup flows.
// If reuse is needed outside cmd/backup, consider moving RunBackupWorkflow to qmpbackup.
package main

import (
	"fmt"
	"strings"

	"github.com/digitalocean/go-qemu/qmp"
	qmpbackup "github.com/valvemist/qmpbackup/backup"
)

// RunBackupWorkflow executes the full backup workflow based on the provided configuration.
func RunBackupWorkflow(monitor *qmp.SocketMonitor, cfg qmpbackup.Config) ([]byte, error) {
	if cfg.CleanBitmap {
		logger.Info("Cleaning bitmap and exiting")
		return qmpbackup.RunBitmapRemove(monitor, cfg)
	}

	if err := AddBlockDeviceEvenIfFileNotExists(monitor, cfg); err != nil {
		logger.Info("prepareBackupTarget:", err)
		return nil, err
	}

	if res, err := PushBackup(monitor, cfg); err != nil {
		logger.Info("backup:", err)
		logger.Info("In case of error with bitmap operation, run program with -cleanBitmap")
		return res, err
	}

	return []byte("OK"), nil
}

// AddBlockDeviceEvenIfFileNotExists attempts to add a block device, creating the image if missing.
func AddBlockDeviceEvenIfFileNotExists(monitor *qmp.SocketMonitor, cfg qmpbackup.Config) error {
	if _, err := qmpbackup.RunBlockDevAdd(monitor, cfg); err != nil {
		msg := err.Error()

		if strings.HasSuffix(msg, "No such file or directory") {
			logger.Info("Missing file detected, attempting to create image", cfg.BackupFile)
			if err = HandleMissingFile(monitor, cfg); err != nil {
				return err
			}
		}

		if strings.Contains(msg, "Duplicated nodes") {
			logger.Info("Trying to remove duplicated node. Please re-run the program")
			if _, err = qmpbackup.RunBlockDevDel(monitor, cfg); err != nil {
				logger.Info("Deleting node failed", err)
			}
			return err
		}

	}
	return nil
}

// HandleMissingFile creates a missing QCOW2 image and retries adding the block device.
func HandleMissingFile(monitor *qmp.SocketMonitor, cfg qmpbackup.Config) error {
	size, err := qmpbackup.RunGetVirtualSize(monitor, cfg)
	if err != nil {
		return err
	}
	if err := qmpbackup.RunCreateQCOW2Image(cfg, size); err != nil {
		return fmt.Errorf("createQCOW2Image: %s", err)
	}
	if _, err := qmpbackup.RunBlockDevAdd(monitor, cfg); err != nil {
		return fmt.Errorf("blockDevAdd retry failed: %s", err)
	}
	return nil
}

// PushBackup performs either a full or incremental backup depending on the configuration.
func PushBackup(monitor *qmp.SocketMonitor, cfg qmpbackup.Config) ([]byte, error) {
	if cfg.IncLevel < 0 {
		if _, err := qmpbackup.RunBitmapAdd(monitor, cfg); err != nil && !strings.Contains(err.Error(), "Bitmap already exists") {
			return nil, err
		}
		return qmpbackup.DoFullBackup(monitor, cfg)
	}
	return qmpbackup.DoIncrementalBackup(monitor, cfg)
}
