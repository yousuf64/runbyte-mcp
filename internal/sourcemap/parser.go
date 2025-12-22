package sourcemap

import (
	"regexp"
	"strconv"
	"strings"
)

// stackParser parses JavaScript stack traces into structured stackFrame objects
type stackParser struct{}

// newStackParser creates a new stack parser
func newStackParser() *stackParser {
	return &stackParser{}
}

// ParseStackTrace parses a full stack trace (multiple lines) into an array of StackFrames
func (p *stackParser) ParseStackTrace(stackTrace string) []stackFrame {
	lines := strings.Split(stackTrace, "\n")
	frames := make([]stackFrame, 0)

	for _, line := range lines {
		if frame := p.ParseStackLine(line); frame != nil {
			frames = append(frames, *frame)
		}
	}

	return frames
}

// ParseStackLine parses a single line from a stack trace
// Handles formats like:
// - at functionName (file:line:column)
// - at file:line:column
// - at functionName (native)
// - at <anonymous> (file:line:column)
func (p *stackParser) ParseStackLine(line string) *stackFrame {
	trimmedLine := strings.TrimSpace(line)

	// Skip empty lines
	if trimmedLine == "" {
		return nil
	}

	// Check for native calls
	if strings.Contains(trimmedLine, "(native)") {
		nativePattern := regexp.MustCompile(`at\s+(.+?)\s+\(native\)`)
		if matches := nativePattern.FindStringSubmatch(trimmedLine); matches != nil {
			functionName := matches[1]
			return &stackFrame{
				Raw:          line,
				FunctionName: functionName,
				FileName:     "native",
				LineNumber:   nil,
				ColumnNumber: nil,
				IsNative:     true,
			}
		}
		// Fallback for native calls without proper format
		return &stackFrame{
			Raw:          line,
			FunctionName: "unknown",
			FileName:     "native",
			LineNumber:   nil,
			ColumnNumber: nil,
			IsNative:     true,
		}
	}

	// Pattern 1: at functionName (file:line:column)
	// Example: at getText (<input>:1:24611)
	pattern1 := regexp.MustCompile(`at\s+(.+?)\s+\((.+?):(\d+):(\d+)\)`)
	if matches := pattern1.FindStringSubmatch(trimmedLine); matches != nil {
		lineNum, _ := strconv.Atoi(matches[3])
		colNum, _ := strconv.Atoi(matches[4])
		return &stackFrame{
			Raw:          line,
			FunctionName: matches[1],
			FileName:     matches[2],
			LineNumber:   &lineNum,
			ColumnNumber: &colNum,
			IsNative:     false,
		}
	}

	// Pattern 2: at file:line:column (no function name)
	// Example: at <input>:1:24611
	pattern2 := regexp.MustCompile(`at\s+(.+?):(\d+):(\d+)`)
	if matches := pattern2.FindStringSubmatch(trimmedLine); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		return &stackFrame{
			Raw:          line,
			FunctionName: "<anonymous>",
			FileName:     matches[1],
			LineNumber:   &lineNum,
			ColumnNumber: &colNum,
			IsNative:     false,
		}
	}

	// Pattern 3: Just file:line:column (no "at")
	pattern3 := regexp.MustCompile(`^(.+?):(\d+):(\d+)$`)
	if matches := pattern3.FindStringSubmatch(trimmedLine); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		return &stackFrame{
			Raw:          line,
			FunctionName: "<anonymous>",
			FileName:     matches[1],
			LineNumber:   &lineNum,
			ColumnNumber: &colNum,
			IsNative:     false,
		}
	}

	// If we can't parse it, return nil
	return nil
}
