package main

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AddConsoleHandler adds a console output handler
func (l *Logger) AddConsoleHandler(level LogLevel, development bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create encoder configuration
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create a console encoder
	var encoder zapcore.Encoder
	if development {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create a level enabler
	levelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= level
	})

	// Create a core
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), levelEnabler)

	// Add the core to the wrapper
	l.coreWrapper.AddCore(l.createRedactingCore(core))
}

// AddFileHandler adds a file output handler
func (l *Logger) AddFileHandler(filePath string, level LogLevel) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Open the log file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Create encoder configuration
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create a JSON encoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create a level enabler
	levelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= level
	})

	// Create a core
	core := zapcore.NewCore(encoder, zapcore.AddSync(file), levelEnabler)

	// Add the core to the wrapper
	l.coreWrapper.AddCore(l.createRedactingCore(core))

	return nil
}

// createRedactingCore wraps a core with redaction functionality
func (l *Logger) createRedactingCore(core zapcore.Core) zapcore.Core {
	return &redactingCore{
		Core:   core,
		logger: l,
	}
}

// redactingCore is a zapcore.Core wrapper that redacts log messages
type redactingCore struct {
	zapcore.Core
	logger *Logger
}

// Check implements zapcore.Core
func (rc *redactingCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	// Redact the message
	redactedEntry := ent
	redactedEntry.Message = rc.logger.redactMessage(ent.Message)

	return rc.Core.Check(redactedEntry, ce)
}
