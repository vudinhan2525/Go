package log

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

type CustomFormatter struct {
	TimestampFormat string
	ColorEnabled    bool
}

func NewCustomFormatter() *CustomFormatter {
	return &CustomFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z", // ISO 8601 with milliseconds
		ColorEnabled:    true,
	}
}
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// Timestamp
	timestamp := entry.Time.Format(f.TimestampFormat)

	// Trace ID
	traceID := "-"
	if entry.Data["trace_id"] != nil {
		if id, ok := entry.Data["trace_id"].(string); ok && id != "" {
			traceID = id
		}
	}

	// Format additional fields
	var fields string
	if len(entry.Data) > 0 {
		var keys []string
		for k := range entry.Data {
			if k != "trace_id" && k != "status" && k != "method" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys) // Sort keys for consistent output

		var dataFields []string
		for _, k := range keys {
			dataFields = append(dataFields, fmt.Sprintf("%s=%v", k, entry.Data[k]))
		}
		if len(dataFields) > 0 {
			fields = fmt.Sprintf("[%s]", strings.Join(dataFields, ", "))
		}
	}

	// Combine
	method := "-"
	status := 0
	message := "-"
	if entry.Data["method"] != nil {
		if m, ok := entry.Data["method"].(string); ok {
			method = m
		}
	}
	if entry.Data["status"] != nil {
		if s, ok := entry.Data["status"].(int); ok {
			status = s
		}
	}
	if entry.Message != "" {
		message = entry.Message
	}

	// Dynamically build the formatter without empty placeholders
	var parts []string

	// Always add timestamp
	parts = append(parts, timestamp)

	// Add level
	parts = append(parts, fmt.Sprintf("|%s%3s%s|", LevelColor(*entry), strings.ToUpper(entry.Level.String()), ResetColor()))

	// Conditionally add method
	if method != "-" {
		parts = append(parts, fmt.Sprintf("|%s %3s %s|", MethodColor(method), method, ResetColor()))
	}

	// Conditionally add status
	if status > 0 {
		parts = append(parts, fmt.Sprintf("|%s %-3d %s|", StatusCodeColor(status), status, ResetColor()))
	}

	// Always add trace ID
	parts = append(parts, fmt.Sprintf("[%s]", traceID))

	// Always add message
	parts = append(parts, message)

	// Conditionally add fields
	if fields != "" {
		parts = append(parts, fields)
	}

	// Join parts and add newline
	formatter := strings.Join(parts, " ") + "\n"
	fmt.Fprint(b, formatter)

	return b.Bytes(), nil
}

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()

	// Configure lumberjack for log rotation
	Logger.SetOutput(&lumberjack.Logger{
		Filename:   "pkg/logger/docland.log",
		MaxSize:    10,   // Megabytes
		MaxBackups: 5,    // Number of backups to keep
		MaxAge:     30,   // Days
		Compress:   true, // Compress log files
	})

	// Use a custom formatter (assuming NewCustomFormatter() is defined elsewhere)
	customFormatter := NewCustomFormatter()
	Logger.SetFormatter(customFormatter)

	// Set log level and output
	Logger.SetLevel(logrus.DebugLevel)
	Logger.SetOutput(os.Stdout)
}

// CloseLogger closes the log file if it is open
func CloseLogger() error {
	if Logger != nil {
		if logFile, ok := Logger.Out.(*os.File); ok {
			return logFile.Close()
		}
	}
	return nil
}

// StatusCodeColor is the ANSI color for appropriately logging http status code to a terminal.
func StatusCodeColor(code int) string {
	switch {
	case code >= http.StatusContinue && code < http.StatusOK:
		return white
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

// MethodColor is the ANSI color for appropriately logging http method to a terminal.
func MethodColor(method string) string {

	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}
func ResetColor() string {
	return reset
}
func LevelColor(entry logrus.Entry) string {
	switch entry.Level {
	case logrus.DebugLevel:
		return magenta
	case logrus.InfoLevel:
		return blue
	case logrus.WarnLevel:
		return yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return red
	default:
		return blue
	}
}
