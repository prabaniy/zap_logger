package main

import (
	"fmt"
	"regexp"

	"go.uber.org/zap/zapcore"
)

func main() {
	// Create a new root logger with default level set to INFO
	// Create a new root logger with redaction enabled
	logger := NewLogger("app", zapcore.InfoLevel)

	// Add console handler with development mode (human-readable output)
	logger.AddConsoleHandler(zapcore.DebugLevel, true)

	// Add file handler for persistent logging
	err := logger.AddFileHandler("application.log", zapcore.InfoLevel)
	if err != nil {
		panic("Failed to create file handler: " + err.Error())
	}

	// Add redaction patterns for sensitive information
	// Credit card numbers
	ccPattern := regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`)
	logger.AddRedaction(ccPattern, "XXXX-XXXX-XXXX-XXXX")

	// Email addresses
	emailPattern := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)
	logger.AddRedaction(emailPattern, "[EMAIL REDACTED]")

	// Basic logging examples
	logger.Debug("This is a debug message") // May not appear depending on level
	logger.Info("This is an info message")
	logger.Info("User information", map[string]interface{}{
		"user_id": 12345,
		"action":  "login",
	})

	// Examples with sensitive information that will be redacted
	logger.Warn("User payment failed with card 4111-1111-1111-1111")
	logger.Error("Failed to send email to user@example.com", map[string]interface{}{
		"error_code": 500,
	})

	// Create a child logger for a specific component
	authLogger := logger.Child("auth")
	authLogger.Info("Authentication service started")
	authLogger.Error("Some error mesage", map[string]interface{}{
		"error_message": fmt.Errorf("error message"),
	})

	// Create a contextual logger with additional fields
	requestLogger := logger.WithContext(map[string]interface{}{
		"request_id": "req-123456",
		"client_ip":  "192.168.1.1",
	})

	requestLogger.Info("Request received")

	// Add more context fields on the fly
	requestLogger.Info("Request processed", map[string]interface{}{
		"duration_ms": 235,
		"status":      200,
	})

	// Create a nested logger (child with context)
	userLogger := authLogger.WithContext(map[string]interface{}{
		"user_id": "user-789",
	})

	userLogger.Debug("User profile accessed")

	// Change log level at runtime
	logger.SetLevel(zapcore.WarnLevel)
	logger.Info("This won't be logged because level is now WARN")
	logger.Warn("But this warning will be logged")
}
