package logger

import (
	"fmt"
	"os"
	"strings"
)

// Logger provides methods for structured logging.
type Logger struct {
	indentationLevel int
	indentString     string
}

// New creates a new Logger instance.
func New() *Logger {
	return &Logger{indentString: "  "}
}

// Indent increases the logging indentation level by one.
func (l *Logger) Indent() {
	l.indentationLevel++
}

// Unindent decreases the logging indentation level by one, if possible.
func (l *Logger) Unindent() {
	l.indentationLevel = max(l.indentationLevel-1, 0)
}

// log is the core logging function.
// prefix: The log prefix (e.g., "[+]", "[-]").
// relativeIndent: 0 for standard logs, 1 for sub-logs.
// format, args: The message format and arguments.
func (l *Logger) log(prefix string, relativeIndent int, format string, args ...any) {
	indentLevel := max(l.indentationLevel+relativeIndent, 0)
	indentation := strings.Repeat(l.indentString, indentLevel)
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s%s\n", prefix, indentation, message)
}

// Info logs an informational message.
func (l *Logger) Info(format string, args ...any) {
	l.log("[+]", 0, format, args...)
}

// SubInfo logs an informational message, indented by the specified level deeper than current.
func (l *Logger) SubInfo(level int, format string, args ...any) {
	l.log("[+]", level, format, args...)
}

// Detail logs a detail/sub-step message.
func (l *Logger) Detail(format string, args ...any) {
	l.log("[-]", 0, format, args...)
}

// SubDetail logs a detail/sub-step message, indented by the specified level deeper than current.
func (l *Logger) SubDetail(level int, format string, args ...any) {
	l.log("[-]", level, format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...any) {
	l.log("[!]", 0, format, args...)
}

// SubWarn logs a warning message, indented by the specified level deeper than current.
func (l *Logger) SubWarn(level int, format string, args ...any) {
	l.log("[!]", level, format, args...)
}

// Error logs an error message.
// It does not exit the program; exiting should be handled by the caller if necessary.
func (l *Logger) Error(err error, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if err != nil {
		l.log("[!]", 0, "%s: %v", message, err)
	} else {
		l.log("[!]", 0, "%s", message)
	}
}

// Fatal logs an error message and then exits the program.
func (l *Logger) Fatal(err error, format string, args ...any) {
	l.Error(err, format, args...)
	os.Exit(1)
}

// Fatalf logs a formatted message and then exits the program.
func (l *Logger) Fatalf(format string, args ...any) {
	l.log("[!]", 0, format, args...)
	os.Exit(1)
}
