package main

import "regexp"

type Config struct {
	Name         string
	Level        LogLevel
	Development  bool
	ConsoleLevel *LogLevel
	FileConfig   map[string]LogLevel
	RedactRegex  map[*regexp.Regexp]string
	RedactFields []string
}
