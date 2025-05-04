package main

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
	name        string
	context     []zap.Field
	redactions  []redaction
	atomicLevel zap.AtomicLevel
	coreWrapper *multiCoreSyncWrapper
	mu          sync.RWMutex
}

// NewLogger creates a new Logger with the specified name and initial log level
func NewLogger(name string, level LogLevel) *Logger {
	// Create an atomic level that can be changed at runtime
	atomicLevel := zap.NewAtomicLevelAt(level)

	// Initialize the multi-core wrapper
	coreWrapper := &multiCoreSyncWrapper{cores: []zapcore.Core{}}

	// Create the logger
	zapLogger := zap.New(coreWrapper)

	return &Logger{
		Logger:      zapLogger,
		name:        name,
		context:     []zap.Field{zap.String("logger", name)},
		redactions:  []redaction{},
		atomicLevel: atomicLevel,
		coreWrapper: coreWrapper,
	}
}

func NewLoggerWithConfig(cfg Config) (*Logger, error) {
	logger := NewLogger(cfg.Name, cfg.Level)

	if cfg.ConsoleLevel != nil {
		logger.AddConsoleHandler(*cfg.ConsoleLevel, cfg.Development)
	}

	for path, level := range cfg.FileConfig {
		if err := logger.AddFileHandler(path, level); err != nil {
			return nil, err
		}
	}

	for regex, replacement := range cfg.RedactRegex {
		logger.AddRedaction(regex, replacement)
	}

	// Apply redact field keys
	if len(cfg.RedactFields) > 0 {
		fieldRedactor := createFieldRedactorCore(logger, cfg.RedactFields)
		logger.coreWrapper.AddCore(fieldRedactor)
	}
	return logger, nil
}

// Debug logs a message at Debug level with context fields
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Redact the message
	redactedMsg := l.redactMessage(msg)

	// Combine all context fields
	allFields := append([]zap.Field{}, l.context...)

	// Add any additional fields
	if len(fields) > 0 && fields[0] != nil {
		for k, v := range fields[0] {
			allFields = append(allFields, zap.Any(k, v))
		}
	}

	l.Logger.Debug(redactedMsg, allFields...)
}

// Info logs a message at Info level with context fields
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Redact the message
	redactedMsg := l.redactMessage(msg)

	// Combine all context fields
	allFields := append([]zap.Field{}, l.context...)

	// Add any additional fields
	if len(fields) > 0 && fields[0] != nil {
		for k, v := range fields[0] {
			allFields = append(allFields, zap.Any(k, v))
		}
	}

	l.Logger.Info(redactedMsg, allFields...)
}

// Warn logs a message at Warn level with context fields
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Redact the message
	redactedMsg := l.redactMessage(msg)

	// Combine all context fields
	allFields := append([]zap.Field{}, l.context...)

	// Add any additional fields
	if len(fields) > 0 && fields[0] != nil {
		for k, v := range fields[0] {
			allFields = append(allFields, zap.Any(k, v))
		}
	}

	l.Logger.Warn(redactedMsg, allFields...)
}

// Error logs a message at Error level with context fields
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Redact the message
	redactedMsg := l.redactMessage(msg)

	// Combine all context fields
	allFields := append([]zap.Field{}, l.context...)

	// Add any additional fields
	if len(fields) > 0 && fields[0] != nil {
		for k, v := range fields[0] {
			allFields = append(allFields, zap.Any(k, v))
		}
	}

	l.Logger.Error(redactedMsg, allFields...)
}

// Fatal logs a message at Fatal level with context fields
func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Redact the message
	redactedMsg := l.redactMessage(msg)

	// Combine all context fields
	allFields := append([]zap.Field{}, l.context...)

	// Add any additional fields
	if len(fields) > 0 && fields[0] != nil {
		for k, v := range fields[0] {
			allFields = append(allFields, zap.Any(k, v))
		}
	}

	l.Logger.Fatal(redactedMsg, allFields...)
}

// SetLevel sets the global minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.atomicLevel.SetLevel(level)
}

// Child creates a child logger with the given name
func (l *Logger) Child(name string) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Create hierarchical name
	childName := l.name
	if childName != "" {
		childName += "." + name
	} else {
		childName = name
	}

	// Create a new logger with the same settings
	child := &Logger{
		Logger:      l.Logger,
		name:        childName,
		context:     append([]zap.Field{}, l.context...),
		redactions:  append([]redaction{}, l.redactions...),
		atomicLevel: l.atomicLevel,
		coreWrapper: l.coreWrapper,
	}

	// Replace the logger name field
	for i, field := range child.context {
		if field.Key == "logger" {
			child.context[i] = zap.String("logger", childName)
			break
		}
	}

	return child
}

// WithContext creates a new logger with additional context fields
func (l *Logger) WithContext(fields map[string]interface{}) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Create a new logger with the same settings
	contextLogger := &Logger{
		Logger:      l.Logger,
		name:        l.name,
		context:     append([]zap.Field{}, l.context...),
		redactions:  append([]redaction{}, l.redactions...),
		atomicLevel: l.atomicLevel,
		coreWrapper: l.coreWrapper,
	}

	// Add the new context fields
	for key, value := range fields {
		contextLogger.context = append(contextLogger.context, zap.Any(key, value))
	}

	return contextLogger
}
