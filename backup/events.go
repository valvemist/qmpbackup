package backup

import (
	"context"

	"github.com/digitalocean/go-qemu/qmp"
)

func Events(ctx context.Context, monitor *qmp.SocketMonitor, callback func(qmp.Event)) {
	stream, _ := monitor.Events(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Debug("Returning from event loop...")
			return // Exit gracefully if context is cancelled
		case e, ok := <-stream:
			if !ok {
				log.Debug("Event loop stream is closed. Exiting...")
				return // Exit if the stream is closed
			}
			callback(e)
		}
	}
}
