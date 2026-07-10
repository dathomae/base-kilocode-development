# MCP Server Configuration and Testing

This document covers configuring clients to connect to your MCP server and testing the integration.

## Claude for Desktop Configuration

To connect your MCP server to Claude for Desktop, you need to edit the client configuration file.

### Configuration File Location

| Platform | Path                                                              |
| -------- | ----------------------------------------------------------------- |
| macOS    | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows  | `%APPDATA%/Claude/claude_desktop_config.json`                     |

Create the file if it does not exist.

### Configuration Format

Add your server to the `mcpServers` key in the configuration file:

```json
{
  "mcpServers": {
    "weather": {
      "command": "uv",
      "args": [
        "--directory",
        "/ABSOLUTE/PATH/TO/PARENT/FOLDER/weather",
        "run",
        "weather.py"
      ]
    }
  }
}
```

### Configuration Parameters

| Parameter     | Description                                          |
| ------------- | ---------------------------------------------------- |
| `command`     | The executable to run (e.g., `uv`, `python`, `node`) |
| `args`        | Array of arguments passed to the command             |
| `--directory` | Working directory for the server (use absolute path) |
| `run`         | uv command to run the script                         |
| `weather.py`  | Your server script filename                          |

### Important Notes

- Use absolute paths for the directory parameter. Get the absolute path by running `pwd` on macOS/Linux or `cd` on Windows Command Prompt.
- On Windows, use double backslashes (`\\`) or forward slashes (`/`) in JSON paths.
- The `mcpServers` key can contain multiple server configurations.

### Example: Multiple Servers

```json
{
  "mcpServers": {
    "weather": {
      "command": "uv",
      "args": [
        "--directory",
        "/home/user/projects/weather",
        "run",
        "weather.py"
      ]
    },
    "database": {
      "command": "python",
      "args": ["/home/user/projects/db_server/server.py"]
    }
  }
}
```

## Testing Your Server

### Step 1: Restart Claude for Desktop

After saving the configuration file, restart Claude for Desktop to load the new server configuration.

### Step 2: Verify Server Connection

1. Open Claude for Desktop
2. Look for the **Add files, connectors, and more** icon (a plus sign) in the chat interface
3. Click the plus icon
4. Hover over the **Connectors** menu
5. You should see your server listed (e.g., "weather")

If your server appears in the Connectors menu, it has been successfully registered.

### Step 3: Test Tool Execution

Test your server by asking Claude to use your tools. For a weather server:

- "What's the weather in Sacramento?"
- "What weather alerts are active for New York?"

When a tool is invoked:

1. Claude will show a message indicating it is using the tool
2. The tool executes on your server
3. Results are displayed in the conversation
4. Claude formulates a natural language response based on the results

## Understanding the Request Flow

When you ask a question that requires external data:

1. **User sends query**: "What's the weather in Sacramento?"
2. **Claude analyzes tools**: Determines that `get_forecast` is relevant
3. **Client executes tool**: Calls your server through the MCP protocol
4. **Server processes request**: Your Python code fetches and formats the data
5. **Results returned**: Raw data is sent back to Claude
6. **Response formulation**: Claude generates a natural language response
7. **User sees result**: Formatted weather information in the chat

## Debugging Configuration Issues

### Server Not Appearing in Connectors Menu

If your server does not appear:

1. Verify the configuration file path is correct
2. Check that the JSON syntax is valid
3. Ensure the absolute path in `--directory` is correct
4. Confirm the server script runs independently without errors
5. Restart Claude for Desktop after configuration changes

### Testing Server Independently

Test your server script directly to ensure it starts:

```bash
# For uv projects
cd /path/to/your/project
uv run weather.py

# For pip-based projects
cd /path/to/your/project
source .venv/bin/activate
python weather.py
```

If the server starts without errors but does not appear in Claude, the issue is likely in the configuration file.

## Platform-Specific Considerations

### Windows

- Use forward slashes (`/`) or escaped backslashes (`\\`) in paths
- The command may need to be the full path to the Python executable
- PowerShell syntax differs from Command Prompt for some operations

### macOS

- The Library folder is hidden by default; use `Cmd+Shift+.` in Finder to show hidden files
- Permissions may need to be granted for the Terminal to access certain directories

### Linux

- Claude for Desktop is not yet available on Linux
- Linux users can build custom MCP clients to test servers
- See the [Building a Client](https://modelcontextprotocol.io/docs/develop/build-client) guide for details

## Next Steps

If you encounter issues not covered here, see [troubleshooting.md](troubleshooting.md) for common problems and solutions, or consult the [official Debugging Guide](https://modelcontextprotocol.io/docs/develop/debugging).
