package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Logger interface {
	Info(message string)
	Warn(message string, err error)
	Error(message string, err error)
	Fatal(message string, err error)
	Close()
}

type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type loggerImpl struct {
	logFile   *os.File
	toConsole bool
}

func NewLogger(filePath string, toConsole bool) (Logger, error) {
	var logFile *os.File
	var err error

	if filePath != "" {
		logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
	}

	return &loggerImpl{logFile: logFile, toConsole: toConsole}, nil
}

func (l *loggerImpl) logJSON(level, message string, err error) {
	entry := LogEntry{
		Time:    time.Now().Format(time.RFC3339),
		Level:   level,
		Message: message,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	logData, _ := json.Marshal(entry)
	logMessage := string(logData)

	if l.toConsole {
		fmt.Println(logMessage)
	}

	if l.logFile != nil {
		fmt.Fprintln(l.logFile, logMessage)
	}
}

func (l *loggerImpl) Info(message string) {
	l.logJSON("info", message, nil)
}

func (l *loggerImpl) Warn(message string, err error) {
	l.logJSON("warn", message, err)
}

func (l *loggerImpl) Error(message string, err error) {
	l.logJSON("error", message, err)
}

func (l *loggerImpl) Fatal(message string, err error) {
	l.logJSON("fatal", message, err)
	os.Exit(1)
}

func (l *loggerImpl) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}
