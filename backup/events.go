package backup

import (
	"context"

	"github.com/digitalocean/go-qemu/qmp"
)

// Events listens for QEMU events and invokes the callback when events are received.
// It cancels the context when a BLOCK_JOB_COMPLETED event is detected.
func Events(ctx context.Context, monitor *qmp.SocketMonitor, cancel context.CancelFunc, callback func(event qmp.Event)) {
	go func() {
		stream, _ := monitor.Events(ctx)
		for e := range stream {
			if callback != nil {
				callback(e)
			}
			if e.Event == "BLOCK_JOB_COMPLETED" {
				cancel()
			}
		}
	}()
}
