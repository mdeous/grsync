package logger

import (
	"fmt"
	"os"
	"strings"
)

const (
	indentUnit = "  "
)

const (
	ColorReset    = "\033[0m"
	ColorBlue     = "\033[34m"
	ColorBoldBlue = "\033[1;34m"
	ColorRed      = "\033[31m"
	ColorBoldRed  = "\033[1;31m"
	ColorBoldCyan = "\033[1;36m"
)

// log is the core unexported logging function.
// prefix: The log prefix (e.g., "[+]", "[-]").
// indentLevel: The absolute number of indent units.
// colorCode: ANSI color code to apply to the output.
// format, args: The message format and arguments.
func log(prefix string, indentLevel int, colorCode string, format string, args ...any) {
	actualIndentLevel := max(indentLevel, 0) // Ensure indentLevel is not negative
	indentation := strings.Repeat(indentUnit, actualIndentLevel)
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s %s%s%s\n", colorCode, prefix, indentation, message, ColorReset)
}

// Info logs an informational message.
func Info(format string, args ...any) {
	log("[+]", 0, ColorBoldBlue, format, args...)
}

// SubInfo logs an informational message, indented by the specified level.
func SubInfo(level int, format string, args ...any) {
	log("[+]", level, ColorBoldBlue, format, args...)
}

// Detail logs a detail/sub-step message.
func Detail(format string, args ...any) {
	log("[-]", 0, ColorBlue, format, args...)
}

// SubDetail logs a detail/sub-step message, indented by the specified level.
func SubDetail(level int, format string, args ...any) {
	log("[-]", level, ColorBlue, format, args...)
}

// Warn logs a warning message.
func Warn(format string, args ...any) {
	log("[!]", 0, ColorRed, format, args...)
}

// SubWarn logs a warning message, indented by the specified level.
func SubWarn(level int, format string, args ...any) {
	log("[!]", level, ColorRed, format, args...)
}

// Error logs an error message.
// It does not exit the program; exiting should be handled by the caller if necessary.
func Error(err error, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if err != nil {
		log("[!]", 0, ColorBoldRed, "%s: %v", message, err)
	} else {
		log("[!]", 0, ColorBoldRed, "%s", message)
	}
}

// Fatal logs an error message and then exits the program.
func Fatal(err error, format string, args ...any) {
	Error(err, format, args...)
	os.Exit(1)
}

// Fatalf logs a formatted message and then exits the program.
func Fatalf(format string, args ...any) {
	log("[!]", 0, ColorBoldRed, format, args...)
	os.Exit(1)
}
