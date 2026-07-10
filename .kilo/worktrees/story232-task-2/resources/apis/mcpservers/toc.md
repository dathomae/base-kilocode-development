# MCP Servers Documentation

This directory contains documentation for building Model Context Protocol (MCP) servers in Python. MCP servers allow you to expose tools, resources, and prompts to LLM applications like Claude for Desktop.

## Contents

**[overview.md](overview.md)**
: Introduction to MCP, core concepts, and architecture

**[setup.md](setup.md)**
: System requirements, environment setup, and SDK installation

**[building_servers.md](building_servers.md)**
: Step-by-step guide to implementing MCP servers with FastMCP

**[configuration.md](configuration.md)**
: Client configuration, testing, and integration with Claude for Desktop

**[troubleshooting.md](troubleshooting.md)**
: Common issues and debugging approaches

## Quick Reference

- **Python Version Required**: 3.10 or higher
- **SDK Version**: MCP SDK 1.2.0 or higher
- **Transport Options**: STDIO (default) or HTTP
- **Core Class**: `FastMCP` from `mcp.server.fastmcp`

## External Resources

- [Official MCP Documentation](https://modelcontextprotocol.io/docs/develop/build-server)
- [Debugging Guide](https://modelcontextprotocol.io/docs/develop/debugging)
- [Example Servers Gallery](https://modelcontextprotocol.io/docs/develop/example-servers)
