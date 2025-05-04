package main

import (
	"regexp"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type redaction struct {
	regex       *regexp.Regexp
	replacement string
}

// redactMessage applies all registered redactions to a message
func (l *Logger) redactMessage(message string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	redacted := message
	for _, r := range l.redactions {
		redacted = r.regex.ReplaceAllString(redacted, r.replacement)
	}
	return redacted
}

// redactField redacts string values in fields if needed
func (l *Logger) redactField(field zapcore.Field) zapcore.Field {
	if field.Type == zapcore.StringType {
		// Get the string value
		str := field.String

		// Apply redactions
		redacted := l.redactMessage(str)

		// Replace the field if it changed
		if redacted != str {
			return zap.String(field.Key, redacted)
		}
	}
	return field
}

// AddRedaction adds a new redaction pattern
func (l *Logger) AddRedaction(pattern *regexp.Regexp, replacement string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.redactions = append(l.redactions, redaction{
		regex:       pattern,
		replacement: replacement,
	})
}

// fieldRedactingCore redacts specific field keys
type fieldRedactingCore struct {
	zapcore.Core
	redactKeys map[string]struct{}
}

func (f *fieldRedactingCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	redactedFields := make([]zapcore.Field, 0, len(fields))
	for _, field := range fields {
		if _, ok := f.redactKeys[field.Key]; ok && field.Type == zapcore.StringType {
			redactedFields = append(redactedFields, zap.String(field.Key, "***REDACTED***"))
		} else {
			redactedFields = append(redactedFields, field)
		}
	}
	return f.Core.Write(ent, redactedFields)
}

func createFieldRedactorCore(logger *Logger, keys []string) zapcore.Core {
	keyMap := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keyMap[k] = struct{}{}
	}

	// Wrap a no-op core; it just intercepts logs to redact
	// In practice, this should wrap real cores, but for this example, we intercept via redacting layer
	noopCore := zapcore.NewNopCore()
	return &fieldRedactingCore{
		Core:       noopCore,
		redactKeys: keyMap,
	}
}
