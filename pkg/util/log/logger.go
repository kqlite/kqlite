package log

import (
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/kqlite/kqlite/pkg/util/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LogLevelInfo  = 0
	LogLevelDebug = 1
)

func logTimeFormat(options *zap.Options) {
	options.EncoderConfigOptions = append(options.EncoderConfigOptions, func(encfg *zapcore.EncoderConfig) {
		encfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.StampMilli)
	})
}

// Log output to specified file location.
func logTo(filepath string) zap.Opts {
	logf, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return func(options *zap.Options) {}
	}
	return zap.WriteTo(logf)
}

// Enable devmode for more info and detail output.
func enableDevMode(loglevel int) zap.Opts {
	return func(options *zap.Options) {
		if loglevel > 0 {
			options.Development = true
		}
	}
}

// Creates and configures a logger with some common options like log level and devmode.
// A log file destination can be specified via the filepath argument or can be empty.
func CreateLogger(name string, loglevel int, filepath string) logr.Logger {
	// Set loglevel for more vebosity.
	if loglevel > 0 {
		return zap.New(logTimeFormat, enableDevMode(loglevel), logTo(filepath), zap.Level(zapcore.Level(-loglevel)))
	}

	logger := zap.New(logTimeFormat, enableDevMode(loglevel), logTo(filepath))
	if name != "" {
		return logger.WithName(name)
	}
	return logger
}
