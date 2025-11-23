package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// GetLogger returns a singleton logger instance
func GetLogger() *logrus.Logger {
	if logger == nil {
		logger = logrus.New()

		// Set log level from environment or default to info
		level := os.Getenv("LOG_LEVEL")
		if level == "" {
			level = "info"
		}

		logLevel, err := logrus.ParseLevel(level)
		if err != nil {
			logLevel = logrus.InfoLevel
		}
		logger.SetLevel(logLevel)

		// Set formatter
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})

		// Set output to stdout by default
		logger.SetOutput(os.Stdout)
	}

	return logger
}