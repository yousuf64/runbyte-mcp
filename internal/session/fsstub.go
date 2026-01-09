package session

import (
	"fmt"
	"os"
	"path/filepath"
)

const fsStubTemplate = `/**
 * Runbyte Sandbox Filesystem API
 * 
 * Available directories:
 * - workspace: Persistent user data (read/write)
 * - cache: Temporary cache (read/write, may be cleared)
 * - temp: Ephemeral data (read/write, cleared after execution)
 * - config: Read-only configuration
 * 
 * @example
 * ` + "```typescript" + `
 * import * as fs from '@runbyte/fs';
 * 
 * const data = await fs.readFile('./workspace/config.json');
 * await fs.writeFile('./cache/result.json', JSON.stringify(result));
 * ` + "```" + `
 */

// @ts-ignore - Injected by WASM runtime
const ws = globalThis.__runbyte_workspace;

/**
 * Read a file from any directory
 * @param path - Path with directory prefix (e.g., './workspace/data.txt')
 * @returns File contents as string
 * @throws Error if file not found or path is invalid
 */
export async function readFile(path: string): Promise<string> {
    return ws.readFile(path);
}

/**
 * Write content to a file in a writable directory
 * @param path - Path with directory prefix (e.g., './workspace/output.txt')
 * @param content - Content to write
 * @throws Error if directory is read-only or limits exceeded
 */
export async function writeFile(path: string, content: string): Promise<void> {
    return ws.writeFile(path, content);
}

/**
 * List files and directories at the given path
 * @param path - Directory path (e.g., './workspace' or './workspace/subdir')
 * @returns Array of file/directory names (directories end with '/')
 */
export async function listFiles(path: string): Promise<string[]> {
    return ws.listFiles(path);
}

/**
 * Delete a file from a writable directory
 * @param path - Path to file to delete
 * @throws Error if directory is read-only or file not found
 */
export async function deleteFile(path: string): Promise<void> {
    return ws.deleteFile(path);
}

/**
 * Check if a file exists (convenience wrapper)
 * @param path - Path to check
 */
export async function exists(path: string): Promise<boolean> {
    try {
        await readFile(path);
        return true;
    } catch {
        return false;
    }
}

/**
 * Read and parse a JSON file
 * @param path - Path to JSON file
 */
export async function readJSON<T = any>(path: string): Promise<T> {
    const content = await readFile(path);
    return JSON.parse(content);
}

/**
 * Write an object as JSON
 * @param path - Path to write to
 * @param data - Object to serialize
 * @param pretty - Use pretty printing (default: true)
 */
export async function writeJSON(path: string, data: any, pretty: boolean = true): Promise<void> {
    const content = pretty ? JSON.stringify(data, null, 2) : JSON.stringify(data);
    return writeFile(path, content);
}
`

// generateFsStub generates the @runbyte/fs TypeScript module
func generateFsStub(bundleDir string) error {
	fsDir := filepath.Join(bundleDir, "builtin", "@runbyte", "fs")
	if err := os.MkdirAll(fsDir, 0755); err != nil {
		return fmt.Errorf("failed to create @runbyte/fs directory: %w", err)
	}

	stubPath := filepath.Join(fsDir, "index.ts")
	if err := os.WriteFile(stubPath, []byte(fsStubTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write @runbyte/fs stub: %w", err)
	}

	return nil
}
