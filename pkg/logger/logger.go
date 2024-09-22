package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LogEntry - структура для записи логов в формате JSON
type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Logger - структура для хранения конфигурации логгера
type Logger struct {
	logFile   *os.File
	toConsole bool
}

// NewLogger - конструктор для создания нового логгера с выводом в файл и/или консоль
func NewLogger(filePath string, toConsole bool) (*Logger, error) {
	var logFile *os.File
	var err error

	// Если путь к файлу указан, то открываем файл для записи логов
	if filePath != "" {
		logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
	}

	return &Logger{logFile: logFile, toConsole: toConsole}, nil
}

// logJSON - вспомогательная функция для записи логов в формате JSON
func (l *Logger) logJSON(level, message string, err error) {
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

	// Если включён вывод в консоль, то выводим в консоль
	if l.toConsole {
		fmt.Println(logMessage)
	}

	// Если файл открыт для записи, то записываем в файл
	if l.logFile != nil {
		fmt.Fprintln(l.logFile, logMessage)
	}
}

// Info - логирование информационных сообщений
func (l *Logger) Info(message string) {
	l.logJSON("info", message, nil)
}

// Warn - логирование предупреждений
func (l *Logger) Warn(message string, err error) {
	l.logJSON("warn", message, nil)
}

// Error - логирование ошибок
func (l *Logger) Error(message string, err error) {
	l.logJSON("error", message, err)
}

// Fatal - логирование фатальных ошибок (с завершением программы)
func (l *Logger) Fatal(message string, err error) {
	l.logJSON("fatal", message, err)
	os.Exit(1)
}

// Close - закрытие лог-файла, если он был открыт
func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}
