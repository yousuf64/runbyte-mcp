package sourcemap

import (
	"fmt"
	gosourcemap "github.com/go-sourcemap/sourcemap"
	"os"
)

func Map(sourceMap string, stack string, debug bool) (string, error) {
	consumer, err := gosourcemap.Parse("", []byte(sourceMap))
	if err != nil {
		return "", err
	}

	parser := newStackParser()
	mappedFrames := mapStackFrames(consumer, parser.ParseStackTrace(stack))

	// Print warnings for unmapped frames to stderr
	for _, frame := range mappedFrames {
		if !frame.Mapped && !frame.IsNative && frame.LineNumber != nil && frame.ColumnNumber != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to map position for %s:%d:%d\n",
				frame.FileName, *frame.LineNumber, *frame.ColumnNumber)
		}
	}

	// Format and output
	f := newFormatter()
	var output string
	if debug {
		output = f.FormatWithMetadata(mappedFrames)
	} else {
		output = f.FormatStackTrace(mappedFrames)
	}

	return output, nil
}

// mapStackFrame maps a single stack frame to its original position
func mapStackFrame(consumer *gosourcemap.Consumer, frame stackFrame) mappedStackFrame {
	// If it's a native call or we don't have position info, return as-is
	if frame.IsNative || frame.LineNumber == nil || frame.ColumnNumber == nil || consumer == nil {
		return mappedStackFrame{
			stackFrame:           frame,
			OriginalFileName:     nil,
			OriginalLineNumber:   nil,
			OriginalColumnNumber: nil,
			OriginalName:         nil,
			Mapped:               false,
		}
	}

	// Use the source map to find the original position
	// Note: go-sourcemap expects 1-indexed line and 0-indexed column
	file, functionName, line, col, ok := consumer.Source(
		*frame.LineNumber,
		*frame.ColumnNumber-1, // Convert to 0-indexed for the library
	)

	// Only consider it mapped if we get a valid result AND the source is different from input
	if ok && file != "" && line > 0 {
		// Convert column back to 1-indexed
		colPlusOne := col + 1

		var origName *string
		if functionName != "" {
			origName = &functionName
		}

		return mappedStackFrame{
			stackFrame:           frame,
			OriginalFileName:     &file,
			OriginalLineNumber:   &line,
			OriginalColumnNumber: &colPlusOne,
			OriginalName:         origName,
			Mapped:               true,
		}
	}

	// No valid mapping found
	return mappedStackFrame{
		stackFrame:           frame,
		OriginalFileName:     nil,
		OriginalLineNumber:   nil,
		OriginalColumnNumber: nil,
		OriginalName:         nil,
		Mapped:               false,
	}
}

// MapStackFrames maps an array of stack frames
func mapStackFrames(consumer *gosourcemap.Consumer, frames []stackFrame) []mappedStackFrame {
	mappedFrames := make([]mappedStackFrame, len(frames))
	for i, frame := range frames {
		mappedFrames[i] = mapStackFrame(consumer, frame)
	}
	return mappedFrames
}
