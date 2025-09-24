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
	logger.Info("Starting backup workflow")
	if cfg.IncLevel < 0 { // full backup
		if res, err := qmpbackup.RunBitmapAdd(monitor, cfg); err != nil {
			logger.Info("RunBitmapAdd:", err)
			return res, err
		}
	}

	if err := AddBlockDeviceEvenIfFileNotExists(monitor, cfg); err != nil {
		logger.Info("prepareBackupTarget:", err)
		return nil, err
	}
	logger.Info("Added block device, going to PushBackup")
	if res, err := PushBackup(monitor, cfg); err != nil {
		logger.Info("backup:", err)
		logger.Info("In case of error with bitmap operation, run program with -clean")
		return res, err
	}
	logger.Info("Backup workflow exiting ")
	return []byte("OK"), nil
}

func CleanAll(monitor *qmp.SocketMonitor, cfg qmpbackup.Config) ([]byte, error) {
	logger.Info("Cleaning all and exiting")
	res, err := qmpbackup.RunBitmapRemove(monitor, cfg)
	if err != nil {
		logger.Info("RunBitmapRemove:", err)
	}

	res, err = BlockJobCancel(monitor, cfg)
	if err != nil {
		logger.Info("BlockJobCancel:", err)
	}

	res, err = BlockDevDel(monitor, cfg)
	if err != nil {
		logger.Info("BlockDevDel:", err)
	}
	return res, err
}

// BlockJobCancel cancels current background job if exists.
func BlockJobCancel(monitor *qmp.SocketMonitor, cfg qmpbackup.Config) ([]byte, error) {
	return qmpbackup.RunBlockJobCancel(monitor, cfg)
}

// BlockDevDel removes a block device from the QEMU monitor.
func BlockDevDel(monitor *qmp.SocketMonitor, cfg qmpbackup.Config) ([]byte, error) {
	return qmpbackup.RunBlockDevDel(monitor, cfg)
}

// AddBlockDeviceEvenIfFileNotExists attempts to add a block device, creating the image if missing.
func AddBlockDeviceEvenIfFileNotExists(monitor *qmp.SocketMonitor, cfg qmpbackup.Config) error {
	if _, err := qmpbackup.RunBlockDevAdd(monitor, cfg); err != nil {
		if strings.HasSuffix(err.Error(), "No such file or directory") {
			logger.Info("Missing file detected, attempting to create image", cfg.BackupFile)
			if err = HandleMissingFile(monitor, cfg); err != nil {
				return err
			}
		}
	}
	logger.Info("Added block device, continue with backup", cfg.BackupFile)
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
	if _, err := qmpbackup.RunBlockDevDel(monitor, cfg); err != nil {
		return fmt.Errorf("BlockDevDel failed: %s", err)
	}
	if _, err := qmpbackup.RunBlockDevAdd(monitor, cfg); err != nil {
		return fmt.Errorf("blockDevAdd failed: %s", err)
	}
	return nil
}

// PushBackup performs either a full or incremental backup depending on the configuration.
func PushBackup(monitor *qmp.SocketMonitor, cfg qmpbackup.Config) ([]byte, error) {
	if cfg.IncLevel < 0 {
		res, err := qmpbackup.DoFullBackup(monitor, cfg)
		logger.Debug("DoFullBackup:", res, err)
		return res, err
	}
	return qmpbackup.DoIncrementalBackup(monitor, cfg)
}
