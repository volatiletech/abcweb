package app

import (
	"strings"

	"github.com/aarondl/zapcolors"
	"github.com/uber-go/zap"
)

// Initialize a zap logger
func (a AppState) InitLogger() {
	// JSON logging for production. Should be coupled with a log analyzer
	// like newrelic, elk, logstash etc.
	if a.Config.LogJSON {
		a.Log = zap.New(
			zap.NewJSONEncoder(),
		)
	} else { // Enable colored logging
		a.Log = zap.New(
			zapcolors.NewColorEncoder(zapcolors.TextTimeFormat("2006-01-02 15:04:05 MST")),
		)
	}

	// Set the minimum log level, defined in the app configuration
	switch strings.ToLower(a.Config.LogLevel) {
	case "debug":
		a.Log.SetLevel(zap.DebugLevel)
	case "info":
		a.Log.SetLevel(zap.InfoLevel)
	case "warn":
		a.Log.SetLevel(zap.WarnLevel)
	case "error":
		a.Log.SetLevel(zap.ErrorLevel)
	case "panic":
		a.Log.SetLevel(zap.PanicLevel)
	case "fatal":
		a.Log.SetLevel(zap.FatalLevel)
	default:
		a.Log.SetLevel(zap.WarnLevel)
	}
}
