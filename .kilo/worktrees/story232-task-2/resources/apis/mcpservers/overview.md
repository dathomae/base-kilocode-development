# MCP Server Overview

The Model Context Protocol (MCP) is a protocol that enables applications like Claude for Desktop to interact with external tools and data sources through a standardized interface. This document explains the core concepts and architecture of MCP servers.

## What is MCP

MCP provides a way to extend the capabilities of LLM applications by connecting them to custom tools, resources, and prompts. An MCP server acts as a bridge between the LLM application and external functionality, allowing the LLM to perform actions like fetching data, executing commands, or accessing external APIs.

## Core Concepts

### Server Capabilities

An MCP server can expose three types of capabilities to clients:

| Capability    | Description                                                                                           |
| ------------- | ----------------------------------------------------------------------------------------------------- |
| **Resources** | File-like data that clients can read, such as API responses, file contents, or database query results |
| **Tools**     | Functions that the LLM can call to perform actions, with user approval typically required             |
| **Prompts**   | Pre-written templates that help users accomplish specific tasks                                       |

### Server Architecture

An MCP server communicates with clients through a transport layer. The protocol uses JSON-RPC for message exchange, enabling structured communication between the client and server.

```
┌─────────────────┐         ┌─────────────────┐         ┌─────────────────┐
│  User Query     │────────▶│  Claude/LLM     │────────▶│  MCP Server     │
│                 │         │  Client         │         │  (Your Code)    │
└─────────────────┘         └─────────────────┘         └─────────────────┘
                                   │                            │
                                   │◄───────────────────────────│
                                   │   Tool Execution Results   │
                                   │                            │
                                   ▼                            │
                            ┌─────────────────┐                │
                            │  External APIs  │◄───────────────┘
                            │  or Services    │    API Calls
                            └─────────────────┘
```

## How It Works

When a user asks a question that requires external data or functionality:

1. The client sends the question to the LLM.
2. The LLM analyzes the available tools and decides which ones to use.
3. The client executes the chosen tools through the MCP server.
4. The results are sent back to the LLM.
5. The LLM formulates a natural language response.
6. The response is displayed to the user.

## Transport Options

MCP servers support two transport mechanisms:

### STDIO Transport

The default and most common transport uses standard input/output for communication. This approach:

- Runs the server as a subprocess of the client
- Uses stdin for receiving messages and stdout for sending responses
- Requires that the server never writes to stdout except for JSON-RPC messages

### HTTP Transport

An alternative transport that uses HTTP for communication. This approach:

- Runs the server as a standalone HTTP service
- Allows standard output logging since it does not interfere with HTTP responses
- Enables remote server deployment

## FastMCP Framework

The Python MCP SDK provides the `FastMCP` class, which simplifies server creation by:

- Using Python type hints to automatically generate tool definitions
- Extracting documentation from docstrings
- Handling JSON-RPC protocol details automatically
- Providing decorators for registering tools, resources, and prompts

The FastMCP class reduces boilerplate code and makes server maintenance easier by keeping tool definitions synchronized with the actual function signatures.

## Client Support

While this documentation focuses on building servers, MCP servers can connect to various clients:

- **Claude for Desktop** - The primary reference client from Anthropic
- **Custom clients** - Applications you build that implement the MCP client protocol
- **Third-party clients** - Other applications that support MCP

## Next Steps

See [setup.md](setup.md) for system requirements and installation instructions, or [building_servers.md](building_servers.md) for a complete implementation guide.
