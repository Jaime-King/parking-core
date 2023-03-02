package logger

import (
	"io"
	"os"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
    // Create a new logger with a JSON formatter
    Log = logrus.New()
    Log.SetFormatter(&logrus.JSONFormatter{})

    // Create the log folder
    if err := os.MkdirAll("logs", 0755); err != nil {
        Log.WithError(err).Error("Failed to create logs directory")
    }

    // Set the output to Stdout and a file
    file, err := os.OpenFile("logs/logfile.json", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
    if err != nil {
        Log.WithError(err).Error("Failed to create log file")
    }

    // Get logging level from the .env file, or default to DEBUG
    level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
    if err != nil {
        level = logrus.DebugLevel
        Log.WithError(err).Warn("Log level not correctly defined in .env file, defaulting to DEBUG")
    }
    Log.SetLevel(level)
    Log.SetOutput(io.MultiWriter(os.Stdout, file))  
}