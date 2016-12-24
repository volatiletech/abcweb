package app

import (
	"strings"

	"github.com/aarondl/zapcolors"
	"github.com/uber-go/zap"
)

// InitLogger initializes a zap logger
func (s State) InitLogger() {
	// JSON logging for production. Should be coupled with a log analyzer
	// like newrelic, elk, logstash etc.
	if s.Config.LogJSON {
		s.Log = zap.New(
			zap.NewJSONEncoder(),
		)
	} else { // Enable colored logging
		s.Log = zap.New(
			zapcolors.NewColorEncoder(zapcolors.TextTimeFormat("2006-01-02 15:04:05 MST")),
		)
	}

	// Set the minimum log level, defined in the app configuration
	switch strings.ToLower(s.Config.LogLevel) {
	case "debug":
		s.Log.SetLevel(zap.DebugLevel)
	case "info":
		s.Log.SetLevel(zap.InfoLevel)
	case "warn":
		s.Log.SetLevel(zap.WarnLevel)
	case "error":
		s.Log.SetLevel(zap.ErrorLevel)
	case "panic":
		s.Log.SetLevel(zap.PanicLevel)
	case "fatal":
		s.Log.SetLevel(zap.FatalLevel)
	default:
		s.Log.SetLevel(zap.WarnLevel)
	}
}
