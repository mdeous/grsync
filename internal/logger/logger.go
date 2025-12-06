package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
)

const (
	indentUnit = "  "
)

// Unicode icons for modern terminal output
const (
	IconSuccess = "✓"
	IconDetail  = "→"
	IconWarning = "⚠"
	IconError   = "✗"
	IconInfo    = "◉"
)

// Legacy color constants (kept for compatibility with waitForWifiConnection)
const (
	ColorReset    = "\033[0m"
	ColorBoldCyan = "\033[1;36m"
)

// Modern styles using lipgloss
var (
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")). // Bright green
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")). // Bright blue
			Bold(true)

	detailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")) // Gray

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Bright red
			Bold(true)

	// Highlight styles for emphasizing specific values
	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51")). // Bright cyan
			Bold(true)

	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")) // Purple/magenta

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")) // Yellow

	numberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")). // Pink/magenta
			Bold(true)
)

// log is the core unexported logging function with modern styling.
func log(icon string, style lipgloss.Style, indentLevel int, format string, args ...any) {
	actualIndentLevel := max(indentLevel, 0)
	indentation := strings.Repeat(indentUnit, actualIndentLevel)
	message := fmt.Sprintf(format, args...)

	styledIcon := style.Render(icon)
	fmt.Printf("%s %s%s\n", styledIcon, indentation, message)
}

// Info logs an informational message with a bright icon.
func Info(format string, args ...any) {
	log(IconInfo, infoStyle, 0, format, args...)
}

// SubInfo logs an informational message, indented by the specified level.
func SubInfo(level int, format string, args ...any) {
	log(IconInfo, infoStyle, level, format, args...)
}

// Detail logs a detail/sub-step message with dimmed styling.
func Detail(format string, args ...any) {
	log(IconDetail, detailStyle, 0, format, args...)
}

// SubDetail logs a detail/sub-step message, indented by the specified level.
func SubDetail(level int, format string, args ...any) {
	log(IconDetail, detailStyle, level, format, args...)
}

// Success logs a success message with a checkmark icon.
func Success(format string, args ...any) {
	log(IconSuccess, successStyle, 0, format, args...)
}

// SubSuccess logs a success message, indented by the specified level.
func SubSuccess(level int, format string, args ...any) {
	log(IconSuccess, successStyle, level, format, args...)
}

// Warn logs a warning message with a warning icon.
func Warn(format string, args ...any) {
	log(IconWarning, warnStyle, 0, format, args...)
}

// SubWarn logs a warning message, indented by the specified level.
func SubWarn(level int, format string, args ...any) {
	log(IconWarning, warnStyle, level, format, args...)
}

// Error logs an error message with an X icon.
// It does not exit the program; exiting should be handled by the caller if necessary.
func Error(err error, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if err != nil {
		log(IconError, errorStyle, 0, "%s: %v", message, err)
	} else {
		log(IconError, errorStyle, 0, "%s", message)
	}
}

// Fatal logs an error message and then exits the program.
func Fatal(err error, format string, args ...any) {
	Error(err, format, args...)
	os.Exit(1)
}

// Fatalf logs a formatted message and then exits the program.
func Fatalf(format string, args ...any) {
	log(IconError, errorStyle, 0, format, args...)
	os.Exit(1)
}

// NewSpinner creates a new spinner with a custom message and modern styling.
// The caller is responsible for starting and stopping the spinner.
func NewSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " " + message
	_ = s.Color("cyan", "bold")
	return s
}

// StartSpinner is a convenience function that creates and starts a spinner.
func StartSpinner(message string) *spinner.Spinner {
	s := NewSpinner(message)
	s.Start()
	return s
}

// StopSpinner stops a spinner and optionally prints a completion message.
func StopSpinner(s *spinner.Spinner, successMsg string) {
	if s == nil {
		return
	}
	s.Stop()
	if successMsg != "" {
		Success("%s", successMsg)
	}
}

// Highlight returns a highlighted (bright cyan) version of the text.
func Highlight(text string) string {
	return highlightStyle.Render(text)
}

// Accent returns an accented (purple) version of the text.
func Accent(text string) string {
	return accentStyle.Render(text)
}

// Path returns a path-styled (yellow) version of the text.
func Path(text string) string {
	return pathStyle.Render(text)
}

// Number returns a number-styled (bold pink) version of the text.
func Number(text string) string {
	return numberStyle.Render(text)
}
