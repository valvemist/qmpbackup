package backup

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/digitalocean/go-qemu/qmp"
)

var log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level:     slog.LevelInfo,
	AddSource: true,
}))

// SetLogger sets the global logger used throughout the qmpbackup package.
func SetLogger(logger *slog.Logger) {
	if logger != nil {
		log = logger
	}
}

// RunQMPAndLog sends a raw QMP command to the monitor and logs the response.
func RunQMPAndLog(monitor *qmp.SocketMonitor, json string) ([]byte, error) {
	log.Debug(json)
	log.Debug("entering monitor.Run")
	raw, err := monitor.Run([]byte(json))
	log.Debug("monitor.Run:", string(raw), err)
	PrettyPrintJSON(string(raw))
	return raw, err
}

// PrettyPrintJSON formats and logs a JSON string for debugging purposes.
func PrettyPrintJSON(raw string) {
	var obj interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		// Invalid JSON, printing raw
		log.Debug(string(raw))
		return
	}

	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		// Failed to pretty-print, printing raw
		log.Debug(raw)
		return
	}

	log.Debug(string(pretty))
}
