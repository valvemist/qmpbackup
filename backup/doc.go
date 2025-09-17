// Package backup provides a high-level interface for orchestrating QEMU block device backups
// using QMP (QEMU Machine Protocol). It supports full and incremental backups, bitmap management,
// block device operations, and image creation.
//
// The package is designed to be used in CLI tools or automation workflows that interact with
// QEMU virtual machines. It wraps low-level QMP commands with structured configuration and
// logging support.
//
// Key features include:
//
//   - Full and incremental backup execution
//   - Dirty bitmap creation and removal
//   - Block device add/remove operations
//   - QCOW2 image creation if missing
//   - JSON command builders for QMP interaction
//   - Event listener for QEMU job completion
//
// Example usage:
//
//	cfg := backup.Config{ ... }
//	monitor := qmp.NewSocketMonitor(...)
//	result, err := backup.DoFullBackup(monitor, cfg)
//
// All exported functions are documented individually.
//
// For CLI orchestration, see cmd/backup/workflow.go.
package backup
