// Package main provides an example command-line interface for the qmpbackup tool.
//
// This package defines the entry point for executing QEMU block device backups using
// the qmpbackup library. It handles configuration parsing, logging setup, and orchestration
// of the backup workflow.
//
// The CLI supports full and incremental backups, bitmap management, and automatic image creation
// when the target file is missing.
//
// For core backup logic, see the qmpbackup package.
package main
