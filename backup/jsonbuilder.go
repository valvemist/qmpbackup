package backup

import "github.com/tidwall/sjson"

// BuildBlockDirtyBitmapAddJSON returns the JSON command to add a dirty bitmap to a block device.
func BuildBlockDirtyBitmapAddJSON(cfg Config) string {
	json := `{}`
	json, _ = sjson.Set(json, "execute", "block-dirty-bitmap-add")
	json, _ = sjson.Set(json, "arguments.node", cfg.DeviceToBackup)
	json, _ = sjson.Set(json, "arguments.name", "bitmap0")
	return json
}

// BuildBlockDirtyBitmapRemoveJSON returns the JSON command to remove a dirty bitmap from a block device.
func BuildBlockDirtyBitmapRemoveJSON(cfg Config) string {
	json := `{}`
	json, _ = sjson.Set(json, "execute", "block-dirty-bitmap-remove")
	json, _ = sjson.Set(json, "arguments.node", cfg.DeviceToBackup)
	json, _ = sjson.Set(json, "arguments.name", "bitmap0")
	return json
}

// BuildFullBackupJSON constructs the JSON command for initiating a full backup.
func BuildFullBackupJSON(cfg Config) string {
	json := `{}`
	json, _ = sjson.Set(json, "execute", "transaction")
	json, _ = sjson.Set(json, "arguments.actions.0.type", "block-dirty-bitmap-clear")
	json, _ = sjson.Set(json, "arguments.actions.0.data.node", cfg.DeviceToBackup)
	json, _ = sjson.Set(json, "arguments.actions.0.data.name", "bitmap0")
	json, _ = sjson.Set(json, "arguments.actions.1.type", "blockdev-backup")
	json, _ = sjson.Set(json, "arguments.actions.1.data.device", cfg.DeviceToBackup)
	json, _ = sjson.Set(json, "arguments.actions.1.data.target", cfg.NodeTarget)
	json, _ = sjson.Set(json, "arguments.actions.1.data.sync", "full")
	json, _ = sjson.Set(json, "arguments.actions.1.data.auto-dismiss", true)
	json, _ = sjson.Set(json, "arguments.actions.1.data.compress", true)
	return json
}

// BuildIncrementalBackupJSON constructs the JSON command for initiating an incremental backup.
func BuildIncrementalBackupJSON(cfg Config) string {
	json := `{}`
	json, _ = sjson.Set(json, "execute", "blockdev-backup")
	json, _ = sjson.Set(json, "arguments.device", cfg.DeviceToBackup)
	json, _ = sjson.Set(json, "arguments.bitmap", "bitmap0")
	json, _ = sjson.Set(json, "arguments.target", cfg.NodeTarget)
	json, _ = sjson.Set(json, "arguments.sync", "incremental")
	return json
}

// BuildBlockDevAddJSON returns the JSON command to add a block device node.
func BuildBlockDevAddJSON(cfg Config) string {
	json := `{}`
	json, _ = sjson.Set(json, "execute", "blockdev-add")
	json, _ = sjson.Set(json, "arguments.node-name", cfg.NodeTarget)
	json, _ = sjson.Set(json, "arguments.driver", "qcow2")
	json, _ = sjson.Set(json, "arguments.file.driver", "file")
	json, _ = sjson.Set(json, "arguments.file.filename", cfg.BackupFile)
	return json
}

// BuildBlockDevDelJSON returns the JSON command to delete a block device node.
func BuildBlockDevDelJSON(cfg Config) string {
	json := `{}`
	json, _ = sjson.Set(json, "execute", "blockdev-del")
	json, _ = sjson.Set(json, "arguments.node-name", cfg.NodeTarget)
	return json
}

// BuildQueryBlockJSON returns the JSON command to query block device information.
func BuildQueryBlockJSON() string {
	json := `{}`
	json, _ = sjson.Set(json, "execute", "query-block")
	return json
}
