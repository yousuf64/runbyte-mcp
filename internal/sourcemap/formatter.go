package sourcemap

import (
	"fmt"
	"regexp"
	"strings"
)

// formatter formats mapped stack frames back into readable stack trace format
type formatter struct{}

// newFormatter creates a new formatter
func newFormatter() *formatter {
	return &formatter{}
}

// FormatStackFrame formats a single mapped stack frame
func (f *formatter) FormatStackFrame(frame mappedStackFrame) string {
	// If the frame wasn't mapped or is native, return the original line
	if !frame.Mapped || frame.IsNative {
		return frame.Raw
	}

	// Build the mapped version
	functionName := frame.FunctionName
	if frame.OriginalName != nil && *frame.OriginalName != "" {
		functionName = *frame.OriginalName
	}

	fileName := frame.FileName
	if frame.OriginalFileName != nil {
		fileName = *frame.OriginalFileName
	}

	line := 0
	if frame.OriginalLineNumber != nil {
		line = *frame.OriginalLineNumber
	} else if frame.LineNumber != nil {
		line = *frame.LineNumber
	}

	column := 0
	if frame.OriginalColumnNumber != nil {
		column = *frame.OriginalColumnNumber
	} else if frame.ColumnNumber != nil {
		column = *frame.ColumnNumber
	}

	// Preserve the original indentation by extracting it from the raw line
	indentPattern := regexp.MustCompile(`^(\s*)`)
	matches := indentPattern.FindStringSubmatch(frame.Raw)
	indent := ""
	if len(matches) > 1 {
		indent = matches[1]
	}

	// Format: at functionName (file:line:column)
	return fmt.Sprintf("%sat %s (%s:%d:%d)", indent, functionName, fileName, line, column)
}

// FormatStackTrace formats an array of mapped stack frames into a complete stack trace
func (f *formatter) FormatStackTrace(frames []mappedStackFrame) string {
	lines := make([]string, len(frames))
	for i, frame := range frames {
		lines[i] = f.FormatStackFrame(frame)
	}
	return strings.Join(lines, "\n")
}

// FormatWithMetadata formats with additional metadata (for debugging)
func (f *formatter) FormatWithMetadata(frames []mappedStackFrame) string {
	lines := make([]string, len(frames))
	for i, frame := range frames {
		formatted := f.FormatStackFrame(frame)
		mappingStatus := "✗ unmapped"
		if frame.Mapped {
			mappingStatus = "✓ mapped"
		}
		lines[i] = fmt.Sprintf("%s %s", formatted, mappingStatus)
	}
	return strings.Join(lines, "\n")
}
