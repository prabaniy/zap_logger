# Zap Logger Module

A custom logging module built on top of [Zap](https://github.com/uber-go/zap) for structured and context-sensitive logging. This module provides:

- Log levels (Debug, Info, Warn, Error, Fatal)
- Contextual logging with dynamic fields
- Redaction of sensitive data (e.g., username, emails)
- Multiple logging handlers (Console and File)
- Child and contextual loggers for hierarchical log entries

## Features

- **Custom Logger**: Easily create a logger with console and file handlers.
- **Redaction**: Automatically redact sensitive information from log messages (e.g., user-info/email addresses).
- **Dynamic Log Levels**: Change log levels dynamically at runtime.
- **Contextual Logging**: Attach context to logs with dynamic fields (e.g., `request_id`, `user_id`).
- **Child Loggers**: Create child loggers to represent specific components or services.

## Installation

1. Clone this repository or add it as a submodule in your Go project.
2. Run the following command to install the dependencies:

```bash
go mod tidy