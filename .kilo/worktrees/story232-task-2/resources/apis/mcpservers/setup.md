# MCP Server Setup

This document covers the system requirements, environment setup, and SDK installation needed to build MCP servers in Python.

## System Requirements

To build MCP servers in Python, you need:

| Requirement     | Version                 | Notes                                                 |
| --------------- | ----------------------- | ----------------------------------------------------- |
| Python          | 3.10 or higher          | Required for type hint features and SDK compatibility |
| MCP SDK         | 1.2.0 or higher         | Provides the FastMCP framework and protocol handling  |
| Package Manager | uv (recommended) or pip | For dependency management                             |

## Environment Setup

### Installing uv (Recommended)

The documentation recommends using `uv` for Python project management. Install it using:

**macOS/Linux:**

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**Windows:**

```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

### Creating a Project

Create a new Python project with uv:

```bash
# Create a new directory for the project
uv init weather

# Navigate into the project directory
cd weather

# Create and activate a virtual environment
uv venv
source .venv/bin/activate  # On macOS/Linux
.venv\Scripts\activate     # On Windows
```

## Installing Dependencies

Install the MCP SDK and required dependencies:

```bash
# Install MCP SDK with CLI support and HTTP client
uv add "mcp[cli]" httpx
```

If using pip instead of uv:

```bash
pip install "mcp[cli]" httpx
```

### Key Dependencies

| Package    | Purpose                                         |
| ---------- | ----------------------------------------------- |
| `mcp[cli]` | The MCP SDK with command-line interface tools   |
| `httpx`    | Async HTTP client for making external API calls |

## Project Structure

A typical MCP server project has this structure:

```
weather/
├── .venv/                  # Virtual environment
├── pyproject.toml          # Project configuration (created by uv)
├── uv.lock                 # Lock file for reproducible builds
└── weather.py              # Your MCP server implementation
```

## Verifying Installation

Create a test file to verify the installation works:

```python
# test_installation.py
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("test")

@mcp.tool()
def hello(name: str) -> str:
    """Say hello to someone."""
    return f"Hello, {name}!"

if __name__ == "__main__":
    mcp.run(transport="stdio")
```

Run the test:

```bash
python test_installation.py
```

If the server starts without errors, your installation is working correctly.

## Logging Considerations

### STDIO-Based Servers

For STDIO transport, you must never write to standard output (stdout). This includes:

- `print()` statements in Python
- `console.log()` in JavaScript
- `fmt.Println()` in Go

Writing to stdout corrupts the JSON-RPC messages and breaks the server.

**Good practice for STDIO servers:**

```python
import logging

# Configure logging to write to stderr
logging.basicConfig(level=logging.INFO, stream=sys.stderr)
logging.info("Processing request")
```

### HTTP-Based Servers

For HTTP transport, standard output logging is acceptable since it does not interfere with HTTP responses.

## Next Steps

Once your environment is set up, proceed to [building_servers.md](building_servers.md) to implement your first MCP server.
