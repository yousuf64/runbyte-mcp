package sandbox

import (
	"context"

	extism "github.com/extism/go-sdk"
)

// TODO: rename to createFileSystemHostFunctions
// createWorkspaceHostFunctions creates all filesystem-related host functions
func createWorkspaceHostFunctions(sfs *SandboxFileSystem) []extism.HostFunction {
	return []extism.HostFunction{
		createReadFileHostFunc(sfs),
		createWriteFileHostFunc(sfs),
		createListFilesHostFunc(sfs),
		createDeleteFileHostFunc(sfs),
	}
}

// createReadFileHostFunc creates the host function for reading files
func createReadFileHostFunc(sfs *SandboxFileSystem) extism.HostFunction {
	return extism.NewHostFunctionWithStack(
		"workspace_readFile",
		func(ctx context.Context, plugin *extism.CurrentPlugin, stack []uint64) {
			offset := stack[0]
			inputData, err := plugin.ReadBytes(offset)
			if err != nil {
				plugin.Logf(extism.LogLevelError, "Failed to read input: %v", err)
				stack[0] = 0
				return
			}

			plugin.Log(extism.LogLevelDebug, "Reading file from sandbox filesystem")

			// Delegate to SandboxFileSystem
			responseData := sfs.HandleReadFile(inputData)

			// Write response
			responseOffset, err := plugin.WriteBytes(responseData)
			if err != nil {
				plugin.Logf(extism.LogLevelError, "Failed to write response: %v", err)
				stack[0] = 0
				return
			}

			stack[0] = responseOffset
		},
		[]extism.ValueType{extism.ValueTypeI64},
		[]extism.ValueType{extism.ValueTypeI64},
	)
}

// createWriteFileHostFunc creates the host function for writing files
func createWriteFileHostFunc(sfs *SandboxFileSystem) extism.HostFunction {
	return extism.NewHostFunctionWithStack(
		"workspace_writeFile",
		func(ctx context.Context, plugin *extism.CurrentPlugin, stack []uint64) {
			offset := stack[0]
			inputData, err := plugin.ReadBytes(offset)
			if err != nil {
				plugin.Logf(extism.LogLevelError, "Failed to read input: %v", err)
				stack[0] = 0
				return
			}

			plugin.Log(extism.LogLevelDebug, "Writing file to sandbox filesystem")

			// Delegate to SandboxFileSystem
			responseData := sfs.HandleWriteFile(inputData)

			// Write response
			responseOffset, err := plugin.WriteBytes(responseData)
			if err != nil {
				plugin.Logf(extism.LogLevelError, "Failed to write response: %v", err)
				stack[0] = 0
				return
			}

			stack[0] = responseOffset
		},
		[]extism.ValueType{extism.ValueTypeI64},
		[]extism.ValueType{extism.ValueTypeI64},
	)
}

// createListFilesHostFunc creates the host function for listing files
func createListFilesHostFunc(sfs *SandboxFileSystem) extism.HostFunction {
	return extism.NewHostFunctionWithStack(
		"workspace_listFiles",
		func(ctx context.Context, plugin *extism.CurrentPlugin, stack []uint64) {
			offset := stack[0]
			inputData, err := plugin.ReadBytes(offset)
			if err != nil {
				plugin.Logf(extism.LogLevelError, "Failed to read input: %v", err)
				stack[0] = 0
				return
			}

			plugin.Log(extism.LogLevelDebug, "Listing files in sandbox filesystem")

			// Delegate to SandboxFileSystem
			responseData := sfs.HandleListFiles(inputData)

			// Write response
			responseOffset, err := plugin.WriteBytes(responseData)
			if err != nil {
				plugin.Logf(extism.LogLevelError, "Failed to write response: %v", err)
				stack[0] = 0
				return
			}

			stack[0] = responseOffset
		},
		[]extism.ValueType{extism.ValueTypeI64},
		[]extism.ValueType{extism.ValueTypeI64},
	)
}

// createDeleteFileHostFunc creates the host function for deleting files
func createDeleteFileHostFunc(sfs *SandboxFileSystem) extism.HostFunction {
	return extism.NewHostFunctionWithStack(
		"workspace_deleteFile",
		func(ctx context.Context, plugin *extism.CurrentPlugin, stack []uint64) {
			offset := stack[0]
			inputData, err := plugin.ReadBytes(offset)
			if err != nil {
				plugin.Logf(extism.LogLevelError, "Failed to read input: %v", err)
				stack[0] = 0
				return
			}

			plugin.Log(extism.LogLevelDebug, "Deleting file from sandbox filesystem")

			// Delegate to SandboxFileSystem
			responseData := sfs.HandleDeleteFile(inputData)

			// Write response
			responseOffset, err := plugin.WriteBytes(responseData)
			if err != nil {
				plugin.Logf(extism.LogLevelError, "Failed to write response: %v", err)
				stack[0] = 0
				return
			}

			stack[0] = responseOffset
		},
		[]extism.ValueType{extism.ValueTypeI64},
		[]extism.ValueType{extism.ValueTypeI64},
	)
}
