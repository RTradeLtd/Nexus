package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a default "sugared" logger based on dev toggle
func NewLogger(dev bool) (sugar *zap.SugaredLogger, err error) {
	var logger *zap.Logger
	if dev {
		// Log:         DebugLevel
		// Encoder:     console
		// Errors:      stderr
		// Sampling:    no
		// Stacktraces: WarningLevel
		// Colors:      capitals
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = config.Build()
	} else {
		// Log:         InfoLevel
		// Encoder:     json
		// Errors:      stderr
		// Sampling:    yes
		// Stacktraces: ErrorLevel
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return
	}

	return logger.Sugar(), nil
}

// NewProcessLogger creates a new logger that sets prefixes on fields for
// logging a specific process
func NewProcessLogger(l *zap.SugaredLogger, process string, fields ...interface{}) *zap.SugaredLogger {
	args := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i += 2 {
		args[i] = process + "." + fields[i].(string)
		args[i+1] = fields[i+1]
	}
	return l.With(args...)
}
