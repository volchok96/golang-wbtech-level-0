package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Logger - интерфейс для работы с логированием
type Logger interface {
	Info(message string)
	Warn(message string, err error)
	Error(message string, err error)
	Fatal(message string, err error)
	Close()
}

// LogEntry - структура для записи логов в формате JSON
type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// loggerImpl - структура для хранения конфигурации логгера
type loggerImpl struct {
	logFile   *os.File
	toConsole bool
}

// NewLogger - конструктор для создания нового логгера с выводом в файл и/или консоль
func NewLogger(filePath string, toConsole bool) (Logger, error) {
	var logFile *os.File
	var err error

	// Если путь к файлу указан, открываем файл для записи логов
	if filePath != "" {
		logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
	}

	// Возвращаем новый экземпляр логгера
	return &loggerImpl{logFile: logFile, toConsole: toConsole}, nil
}

// logJSON - вспомогательная функция для записи логов в формате JSON
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

	// Если включён вывод в консоль, выводим сообщение
	if l.toConsole {
		fmt.Println(logMessage)
	}

	// Если файл открыт для записи, записываем в файл
	if l.logFile != nil {
		fmt.Fprintln(l.logFile, logMessage)
	}
}

// Info - логирование информационных сообщений
func (l *loggerImpl) Info(message string) {
	l.logJSON("info", message, nil)
}

// Warn - логирование предупреждений
func (l *loggerImpl) Warn(message string, err error) {
	l.logJSON("warn", message, err)
}

// Error - логирование ошибок
func (l *loggerImpl) Error(message string, err error) {
	l.logJSON("error", message, err)
}

// Fatal - логирование фатальных ошибок (с завершением программы)
func (l *loggerImpl) Fatal(message string, err error) {
	l.logJSON("fatal", message, err)
	os.Exit(1)
}

// Close - закрытие лог-файла, если он был открыт
func (l *loggerImpl) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}
