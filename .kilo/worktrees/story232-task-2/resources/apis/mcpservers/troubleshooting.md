# Troubleshooting MCP Servers

This document covers common issues when building and running MCP servers, along with solutions and debugging approaches.

## Common Issues

### Server Not Appearing in Claude for Desktop

| Symptom                       | Possible Cause              | Solution                                                                     |
| ----------------------------- | --------------------------- | ---------------------------------------------------------------------------- |
| Server not in Connectors menu | Invalid JSON in config file | Validate JSON syntax using a linter                                          |
| Server not in Connectors menu | Wrong config file path      | Verify path matches your platform (see [configuration.md](configuration.md)) |
| Server not in Connectors menu | Relative path used          | Use absolute paths for the `--directory` argument                            |
| Server not in Connectors menu | Server crashes on startup   | Test the server independently from the command line                          |
| Server not in Connectors menu | Missing file extension      | Ensure the script filename includes `.py`                                    |

### STDIO Communication Errors

STDIO transport requires that your server never writes to standard output (stdout) except for JSON-RPC messages.

| Symptom                    | Cause                        | Solution                                         |
| -------------------------- | ---------------------------- | ------------------------------------------------ |
| Server crashes immediately | `print()` statements in code | Remove all `print()` calls or redirect to stderr |
| Intermittent failures      | Logging to stdout            | Configure logging to use `sys.stderr` instead    |
| Corrupted messages         | Debug output from libraries  | Set logging levels to suppress output            |

**Proper logging configuration:**

```python
import logging
import sys

# Configure logging to write to stderr
logging.basicConfig(
    level=logging.INFO,
    stream=sys.stderr,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

# Use logging instead of print
logging.info("Processing request")
```

### Tool Execution Errors

| Symptom                   | Cause                           | Solution                                      |
| ------------------------- | ------------------------------- | --------------------------------------------- |
| Tool not found            | Missing `@mcp.tool()` decorator | Ensure all tools use the decorator            |
| Wrong parameter types     | Missing type hints              | Add type hints to all parameters              |
| LLM uses tool incorrectly | Poor docstring                  | Write clear docstrings with Args descriptions |
| Async errors              | Missing `await`                 | Use `await` for async function calls          |

### Import Errors

| Symptom               | Cause                             | Solution                                             |
| --------------------- | --------------------------------- | ---------------------------------------------------- |
| `ModuleNotFoundError` | Virtual environment not activated | Activate the virtual environment before running      |
| `ModuleNotFoundError` | Missing dependencies              | Install required packages: `uv add "mcp[cli]" httpx` |
| Import path issues    | Wrong Python interpreter          | Ensure the config file uses the correct command path |

## Debugging Approaches

### Testing the Server Independently

Before connecting to Claude for Desktop, verify your server runs correctly:

```bash
# Navigate to your project
cd /path/to/your/project

# Activate the virtual environment
source .venv/bin/activate  # macOS/Linux
.venv\Scripts\activate     # Windows

# Run the server directly
python weather.py
```

If the server starts without errors, the issue is likely in the client configuration.

### Checking JSON-RPC Communication

For STDIO-based servers, you can test the communication manually:

1. Start your server in a terminal
2. Send a valid JSON-RPC request to stdin
3. Check if the server responds correctly

Example initialization request:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": { "name": "test", "version": "1.0" }
  }
}
```

### Using the MCP Inspector

The MCP SDK includes tools for debugging servers. Check the [official Debugging Guide](https://modelcontextprotocol.io/docs/develop/debugging) for details on using the MCP inspector.

## Platform-Specific Issues

### Windows

**Issue**: Path separators in JSON

- Windows paths use backslashes (`\`)
- JSON requires backslashes to be escaped (`\\`)
- Solution: Use forward slashes (`/`) instead, which work on all platforms

**Issue**: Command not found

- Windows may not find `uv` or `python` in PATH
- Solution: Use the full path to the executable:

```json
{
  "command": "C:\\Users\\username\\AppData\\Local\\Programs\\Python\\Python311\\python.exe",
  "args": ["C:/projects/weather/weather.py"]
}
```

### macOS

**Issue**: Permission denied

- macOS may block execution of downloaded files
- Solution: Grant permissions in System Preferences > Security & Privacy

**Issue**: Hidden Library folder

- The `~/Library` folder is hidden in Finder
- Solution: Press `Cmd+Shift+.` in Finder to show hidden files

### Linux

**Issue**: Claude for Desktop not available

- Claude for Desktop does not support Linux
- Solution: Build a custom MCP client for testing (see [Building a Client](https://modelcontextprotocol.io/docs/develop/build-client))

## Best Practices for Avoiding Issues

1. **Always use type hints**: Without them, FastMCP cannot generate proper tool schemas
2. **Write comprehensive docstrings**: The LLM relies on these to understand your tools
3. **Handle errors gracefully**: Return informative strings rather than raising exceptions
4. **Test incrementally**: Verify each tool works before adding more
5. **Use logging properly**: Write logs to stderr, never stdout
6. **Validate JSON config**: Use a JSON linter before saving configuration files

## Getting Help

If you encounter issues not covered in this document:

1. Check the [official Debugging Guide](https://modelcontextprotocol.io/docs/develop/debugging)
2. Review the [Example Servers Gallery](https://modelcontextprotocol.io/docs/develop/example-servers) for working implementations
3. Consult the [MCP Documentation](https://modelcontextprotocol.io/docs/develop/build-server) for the latest updates

## Quick Checklist

Before seeking help, verify:

- [ ] Python 3.10 or higher is installed
- [ ] MCP SDK 1.2.0 or higher is installed
- [ ] All tools have `@mcp.tool()` decorator
- [ ] All parameters have type hints
- [ ] No `print()` statements in the code
- [ ] Configuration file uses absolute paths
- [ ] JSON syntax in config file is valid
- [ ] Server runs independently without errors
- [ ] Claude for Desktop was restarted after config changes
