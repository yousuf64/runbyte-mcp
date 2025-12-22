package sourcemap

// stackFrame represents a single stack frame parsed from a stack trace
type stackFrame struct {
	// The raw original line from the stack trace
	Raw string
	// Function name (or '<anonymous>' if anonymous)
	FunctionName string
	// Source file path
	FileName string
	// Line number (1-indexed), nil if not available
	LineNumber *int
	// Column number (1-indexed), nil if not available
	ColumnNumber *int
	// Whether this is a native call
	IsNative bool
}

// mappedStackFrame represents a mapped stack frame with original source information
type mappedStackFrame struct {
	stackFrame
	// Original source file path (from source map)
	OriginalFileName *string
	// Original line number (1-indexed)
	OriginalLineNumber *int
	// Original column number (1-indexed)
	OriginalColumnNumber *int
	// Original function/symbol name from source map
	OriginalName *string
	// Whether mapping was successful
	Mapped bool
}
