package backup

import "fmt"

// Config holds configuration of backup operation
type Config struct {
	CleanBitmap    bool
	SocketFile     string
	BackupFile     string
	BackingFile    string
	DeviceToBackup string
	NodeTarget     string
	IncLevel       int
}

// GenerateBackupFilename generates a backup filename based on the configuration parameters.
func GenerateBackupFilename(cfg *Config) {
	switch {
	case cfg.IncLevel < 0:
		cfg.BackupFile = cfg.BackupFile + ".full.qcow2"
	case cfg.IncLevel == 0:
		cfg.BackingFile = cfg.BackupFile + ".full.qcow2"
		cfg.BackupFile = cfg.BackupFile + ".inc0.qcow2"
	default:
		cfg.BackingFile = cfg.BackupFile + ".inc" + itoa(cfg.IncLevel-1) + ".qcow2"
		cfg.BackupFile = cfg.BackupFile + ".inc" + itoa(cfg.IncLevel) + ".qcow2"
	}
}

// itoa converts an integer to its string representation.
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
