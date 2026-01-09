/**
 * Type definitions for Runbyte Plugin
 */

declare module "main" {
    /**
     * Main entry point - executes user code in sandbox
     * Extism exports take no params and return an I32
     */
    export function executeCode(): I32;
}

declare module "extism:host" {
    interface user {
        /**
         * Call an MCP tool on a downstream server
         * @param ptr Pointer to JSON string containing {serverName, toolName, args}
         * @returns Pointer to JSON string containing {result, error}
         */
        callMcpTool(ptr: I64): I64;

        /**
         * Read a file from the sandbox filesystem
         * @param ptr Pointer to JSON string containing {path}
         * @returns Pointer to JSON string containing {success, data, error}
         */
        workspace_readFile(ptr: I64): I64;

        /**
         * Write a file to the sandbox filesystem
         * @param ptr Pointer to JSON string containing {path, content}
         * @returns Pointer to JSON string containing {success, error}
         */
        workspace_writeFile(ptr: I64): I64;

        /**
         * List files in a sandbox filesystem directory
         * @param ptr Pointer to JSON string containing {path}
         * @returns Pointer to JSON string containing {success, files, error}
         */
        workspace_listFiles(ptr: I64): I64;

        /**
         * Delete a file from the sandbox filesystem
         * @param ptr Pointer to JSON string containing {path}
         * @returns Pointer to JSON string containing {success, error}
         */
        workspace_deleteFile(ptr: I64): I64;
    }
}

