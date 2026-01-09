/**
 * Runbyte JavaScript Plugin
 *
 * This plugin runs user code in a WebAssembly sandbox with access to
 * downstream MCP tools via the callMcpTool function.
 */

async function executeCode() {
    try {
        const {callMcpTool, workspace_readFile, workspace_writeFile, workspace_listFiles, workspace_deleteFile} = Host.getFunctions();
        // TODO: Make sure callMcpTool is not accessible

        /**
         * Call an MCP tool on a downstream server
         * @param {string} serverName - Name of the MCP server
         * @param {string} toolName - Name of the tool to call
         * @param {object} args - Arguments to pass to the tool
         * @returns {any} The result from the MCP tool
         */
        function callTool(serverName, toolName, args) {
            const msg = {
                serverName,
                toolName,
                args: args || {}
            };

            // Call host function
            const mem = Memory.fromString(JSON.stringify(msg));
            const offset = callMcpTool(mem.offset);
            const response = Memory.find(offset).readJsonObject();

            // Check if the call was successful
            if (response.error) {
                throw new Error(response.error);
            }

            // Try to parse as JSON, fallback to raw string
            try {
                return JSON.parse(response.result);
            } catch {
                return response.result;
            }
        }

        // Workspace filesystem API
        const workspace = {
            async readFile(path) {
                const msg = { path };
                const mem = Memory.fromString(JSON.stringify(msg));
                const offset = workspace_readFile(mem.offset);
                const response = Memory.find(offset).readJsonObject();
                
                if (!response.success) {
                    throw new Error(response.error);
                }
                return response.data;
            },
            
            async writeFile(path, content) {
                const msg = { path, content };
                const mem = Memory.fromString(JSON.stringify(msg));
                const offset = workspace_writeFile(mem.offset);
                const response = Memory.find(offset).readJsonObject();
                
                if (!response.success) {
                    throw new Error(response.error);
                }
            },
            
            async listFiles(path) {
                const msg = { path };
                const mem = Memory.fromString(JSON.stringify(msg));
                const offset = workspace_listFiles(mem.offset);
                const response = Memory.find(offset).readJsonObject();
                
                if (!response.success) {
                    throw new Error(response.error);
                }
                return response.files;
            },
            
            async deleteFile(path) {
                const msg = { path };
                const mem = Memory.fromString(JSON.stringify(msg));
                const offset = workspace_deleteFile(mem.offset);
                const response = Memory.find(offset).readJsonObject();
                
                if (!response.success) {
                    throw new Error(response.error);
                }
            }
        };
        
        // Expose to bundled code
        globalThis.__runbyte_workspace = workspace;
        globalThis.__runbyte_callTool = callTool;

        // Get user's code from input
        const code = Host.inputString();

        try {
            // Execute user's code
            // User can call: callMcpTool("github", "list_repos", {})
            const result = await eval(code);

            // Return result as JSON string
            Host.outputString(JSON.stringify({
                error: null,
                stack: null,
                result: JSON.stringify(result)
            }))
        } catch (error) {
            // Return error information
            Host.outputString(JSON.stringify({
                error: error.message,
                stack: error.stack,
                result: null
            }));
        }
    } catch (error) {
        Host.outputString(JSON.stringify({
            error: error.message,
            stack: null,
            result: null
        }))
    }
}

module.exports = { executeCode };

