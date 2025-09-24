// blockops.go contains low-level QMP operations for block devices,
// including bitmap management, image creation, and virtual size querying.

package backup

import (
	"fmt"
	"os/exec"

	"github.com/tidwall/gjson"

	"github.com/digitalocean/go-qemu/qmp"
)

// Block device operations

// RunBlockDevAdd adds a block device to the QEMU monitor using the provided configuration.
func RunBlockDevAdd(monitor *qmp.SocketMonitor, cfg Config) ([]byte, error) {
	return RunQMPAndLog(monitor, BuildBlockDevAddJSON(cfg))
}

// RunBlockDevDel removes a block device node from the QEMU monitor.
func RunBlockDevDel(monitor *qmp.SocketMonitor, cfg Config) ([]byte, error) {
	return RunQMPAndLog(monitor, BuildBlockDevDelJSON(cfg))
}

// Bitmap operations

// RunBitmapAdd adds a bitmap to the block device to track changes for incremental backup.
func RunBitmapAdd(monitor *qmp.SocketMonitor, cfg Config) ([]byte, error) {
	return RunQMPAndLog(monitor, BuildBlockDirtyBitmapAddJSON(cfg))
}

// RunBitmapRemove removes the bitmap from the QEMU block device as part of cleanup.
func RunBitmapRemove(monitor *qmp.SocketMonitor, cfg Config) ([]byte, error) {
	return RunQMPAndLog(monitor, BuildBlockDirtyBitmapRemoveJSON(cfg))
}

// RunBitmapRemove removes the bitmap from the QEMU block device as part of cleanup.
func RunBlockJobCancel(monitor *qmp.SocketMonitor, cfg Config) ([]byte, error) {
	return RunQMPAndLog(monitor, BuildBlockJobCancelJSON(cfg))
}

// Virtual image handling functions

// RunCreateQCOW2Image creates a new QCOW2 image file with the specified size.
func RunCreateQCOW2Image(cfg Config, size string) error {
	var cmd *exec.Cmd
	if cfg.BackingFile == "" {
		cmd = exec.Command("qemu-img", "create", "-f", "qcow2", cfg.BackupFile, size)
	} else {
		cmd = exec.Command("qemu-img", "create", "-f", "qcow2", cfg.BackupFile, "-b", cfg.BackingFile, "-F", "qcow2")
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create image: %v\nOutput: %s", err, string(output))
	}
	log.Info(fmt.Sprintf("Image created successfully: %s\n", cfg.BackupFile))
	return nil
}

// RunGetVirtualSize retrieves the virtual size of the block device from QEMU.
func RunGetVirtualSize(monitor *qmp.SocketMonitor, cfg Config) (string, error) {
	raw, err := RunQMPAndLog(monitor, BuildQueryBlockJSON())
	if err != nil {
		return "", err
	}

	devices := gjson.GetBytes(raw, "return").Array()
	for _, device := range devices {
		if device.Get("device").String() == cfg.DeviceToBackup {
			sizeBytes := device.Get("inserted.image.virtual-size").Int()
			sizeGiB := float64(sizeBytes) / (1024 * 1024 * 1024)
			sizeStr := fmt.Sprintf("%.0fG", sizeGiB)
			log.Info(fmt.Sprintf("Source image size: %s\n", sizeStr))
			return sizeStr, nil
		}
	}

	return "", fmt.Errorf("device %s not found or has no virtual size. Is -device parameter set correctly?", cfg.DeviceToBackup)
}
