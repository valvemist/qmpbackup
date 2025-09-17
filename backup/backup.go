package backup

import (
	"github.com/digitalocean/go-qemu/qmp"
)

// DoFullBackup performs a full backup using QMP commands based on the provided configuration.
func DoFullBackup(monitor *qmp.SocketMonitor, cfg Config) ([]byte, error) {
	return RunQMPAndLog(monitor, BuildFullBackupJSON(cfg))
}

// DoIncrementalBackup performs an incremental backup using QMP commands and bitmap tracking.
func DoIncrementalBackup(monitor *qmp.SocketMonitor, cfg Config) ([]byte, error) {
	return RunQMPAndLog(monitor, BuildIncrementalBackupJSON(cfg))
}
