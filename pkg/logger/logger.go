package logger

import (
	"context"

	skLogger "github.com/sk-pkg/logger"
	"go.uber.org/zap"
)

var globalLogger *skLogger.Manager

// Init initializes the logger with the specified level and format
func Init(level, driver, path string) {
	// Create logger options
	opts := []skLogger.Option{
		skLogger.WithLevel(level),
	}

	// Set driver based on format
	// sk-pkg/logger supports "stdout" and "file" as drivers
	opts = append(opts, skLogger.WithDriver(driver))

	// Set log path
	opts = append(opts, skLogger.WithLogPath(path))

	// Enable color for console format
	opts = append(opts, skLogger.WithColor(true))

	// Initialize logger
	var err error
	globalLogger, err = skLogger.New(opts...)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

// Debug logs a debug message with key-value pairs
func Debug(msg string, keysAndValues ...interface{}) {
	if globalLogger == nil {
		return
	}
	fields := convertToZapFields(keysAndValues)
	globalLogger.Debug(context.Background(), msg, fields...)
}

// Info logs an info message with key-value pairs
func Info(msg string, keysAndValues ...interface{}) {
	if globalLogger == nil {
		return
	}
	fields := convertToZapFields(keysAndValues)
	globalLogger.Info(context.Background(), msg, fields...)
}

// Warn logs a warning message with key-value pairs
func Warn(msg string, keysAndValues ...interface{}) {
	if globalLogger == nil {
		return
	}
	fields := convertToZapFields(keysAndValues)
	globalLogger.Warn(context.Background(), msg, fields...)
}

// Error logs an error message with key-value pairs
func Error(msg string, keysAndValues ...interface{}) {
	if globalLogger == nil {
		return
	}
	fields := convertToZapFields(keysAndValues)
	globalLogger.Error(context.Background(), msg, fields...)
}

// Fatal logs a fatal message with key-value pairs and exits
func Fatal(msg string, keysAndValues ...interface{}) {
	if globalLogger == nil {
		panic(msg)
	}
	fields := convertToZapFields(keysAndValues)
	globalLogger.Fatal(context.Background(), msg, fields...)
}

// convertToZapFields converts key-value pairs to zap.Field slices
func convertToZapFields(keysAndValues []interface{}) []zap.Field {
	if len(keysAndValues) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 >= len(keysAndValues) {
			break
		}

		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}

		value := keysAndValues[i+1]
		fields = append(fields, zap.Any(key, value))
	}

	return fields
}
