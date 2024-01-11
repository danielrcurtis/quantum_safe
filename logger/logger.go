package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger
var hostname string
var serviceName string
var initLog map[string]interface{}

func init() {
	initLog = make(map[string]interface{})

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		initLog["hostnameMessage"] = fmt.Sprintf("Error retrieving hostname: %v", err) + "Setting hostname to unkw"
		hostname = "unkw"
	}

	serviceName = os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		initLog["serviceNameMessage"] = "SERVICE_NAME is not set, using Melina as default"
		serviceName = "Quantum Safe"
	}

	// Create logs directory if not exists
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		if err := os.Mkdir("./logs", 0755); err != nil {
			fmt.Printf("Warning: Unable to create log directory './logs': %v\n", err)
		}
	}
	// Open or create log files in the logs directory
	file, err := os.OpenFile("./logs/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	errorLog, err := os.OpenFile("./logs/errors.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	logLevel := os.Getenv("LOG_LEVEL")
	var zapLevel zapcore.Level
	if logLevel == "" {
		logLevel = "debug"
	}
	if logLevel == "debug" {
		zapLevel = zap.DebugLevel
	} else if logLevel == "info" {
		zapLevel = zap.InfoLevel
	} else if logLevel == "warn" {
		zapLevel = zap.WarnLevel
	} else if logLevel == "error" {
		zapLevel = zap.ErrorLevel
	} else if logLevel == "fatal" {
		zapLevel = zap.FatalLevel
	} else if logLevel == "panic" {
		zapLevel = zap.PanicLevel
	} else {
		zapLevel = zap.DebugLevel
	}
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),  // Setting log level
		Development:      true,                            // Development mode (stacktrace on warnings and above)
		Encoding:         "json",                          // or "console"
		OutputPaths:      []string{"stdout", file.Name()}, // Output to console
		ErrorOutputPaths: []string{"stderr", errorLog.Name()},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",
			LevelKey:   "level",
			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			// You can add more configuration for the encoder here...
		},
		InitialFields: map[string]interface{}{
			"hostname":    hostname,
			"serviceName": serviceName,
		},
	}

	Log, err = cfg.Build()
	if err != nil {
		panic(err)
	}
	Log.Debug("Logger initialized")
	if len(initLog) > 0 {
		for key, value := range initLog {
			Log.Sugar().Info("%s, %v\n", key, value)
		}
	}
	Log.Info("Logger set to " + logLevel + " level")
}

// WithFields adds structured context to the logger.
func WithFields(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}
