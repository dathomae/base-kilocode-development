# Building MCP Servers

This document provides a step-by-step guide to implementing MCP servers using the FastMCP framework in Python.

## Server Initialization

Every MCP server starts with the `FastMCP` class from the `mcp.server.fastmcp` module. The class handles protocol details automatically and uses Python type hints to generate tool definitions.

### Basic Server Structure

```python
from typing import Any
import httpx
from mcp.server.fastmcp import FastMCP

# Initialize the FastMCP server with a name
mcp = FastMCP("weather")

# Constants for external API
NWS_API_BASE = "https://api.weather.gov"
USER_AGENT = "weather-app/1.0"
```

The server name passed to `FastMCP()` identifies your server to clients and should be descriptive of its functionality.

## Creating Tools

Tools are the primary way MCP servers expose functionality to LLMs. Use the `@mcp.tool()` decorator to register functions as tools.

### Tool Definition

```python
@mcp.tool()
async def get_alerts(state: str) -> str:
    """Get weather alerts for a US state.

    Args:
        state: Two-letter US state code (e.g. CA, NY)
    """
    url = f"{NWS_API_BASE}/alerts/active/area/{state}"
    data = await make_nws_request(url)

    if not data or "features" not in data:
        return "Unable to fetch alerts or no alerts found."

    if not data["features"]:
        return "No active alerts for this state."

    alerts = [format_alert(feature) for feature in data["features"]]
    return "\n---\n".join(alerts)
```

### Key Principles for Tool Definitions

1. **Type hints are required**: The FastMCP framework uses type hints to generate the tool schema. Without them, the tool will not work correctly.

2. **Docstrings become descriptions**: The docstring (and Args section) provides information to the LLM about what the tool does and what parameters it expects.

3. **Async support**: Tools can be async functions, which is recommended for I/O operations like HTTP requests.

4. **Return types**: Specify return types so the framework knows what kind of data the tool produces.

## Helper Functions

Not all functions need to be exposed as tools. Private helper functions keep your code organized:

```python
async def make_nws_request(url: str) -> dict[str, Any] | None:
    """Make a request to the NWS API with proper error handling."""
    headers = {
        "User-Agent": USER_AGENT,
        "Accept": "application/geo+json"
    }
    async with httpx.AsyncClient() as client:
        try:
            response = await client.get(url, headers=headers, timeout=30.0)
            response.raise_for_status()
            return response.json()
        except Exception:
            return None


def format_alert(feature: dict) -> str:
    """Format an alert feature into a readable string."""
    props = feature["properties"]
    return f"""
Event: {props.get("event", "Unknown")}
Area: {props.get("areaDesc", "Unknown")}
Severity: {props.get("severity", "Unknown")}
Description: {props.get("description", "No description available")}
Instructions: {props.get("instruction", "No specific instructions provided")}
"""
```

## Multi-Step Tool Example

Some tools require multiple API calls to complete their task:

```python
@mcp.tool()
async def get_forecast(latitude: float, longitude: float) -> str:
    """Get weather forecast for a location.

    Args:
        latitude: Latitude of the location
        longitude: Longitude of the location
    """
    # First get the forecast grid endpoint
    points_url = f"{NWS_API_BASE}/points/{latitude},{longitude}"
    points_data = await make_nws_request(points_url)

    if not points_data:
        return "Unable to fetch forecast data for this location."

    # Get the forecast URL from the points response
    forecast_url = points_data["properties"]["forecast"]
    forecast_data = await make_nws_request(forecast_url)

    if not forecast_data:
        return "Unable to fetch detailed forecast."

    # Format the forecast periods
    periods = forecast_data["properties"]["periods"]
    forecasts = []
    for period in periods[:5]:  # Show next 5 periods
        forecast = f"""
{period["name"]}:
Temperature: {period["temperature"]}°{period["temperatureUnit"]}
Wind: {period["windSpeed"]} {period["windDirection"]}
Forecast: {period["detailedForecast"]}
"""
        forecasts.append(forecast)

    return "\n---\n".join(forecasts)
```

## Running the Server

To start the server, call the `run()` method with the desired transport:

```python
def main():
    # Initialize and run the server with STDIO transport
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
```

### Transport Options

| Transport | Use Case                    | Configuration                          |
| --------- | --------------------------- | -------------------------------------- |
| `stdio`   | Default, runs as subprocess | `mcp.run(transport="stdio")`           |
| `http`    | Standalone HTTP service     | `mcp.run(transport="http", port=8080)` |

## Complete Example

Here is a complete MCP server implementation:

```python
from typing import Any
import httpx
from mcp.server.fastmcp import FastMCP

# Initialize FastMCP server
mcp = FastMCP("weather")

# Constants
NWS_API_BASE = "https://api.weather.gov"
USER_AGENT = "weather-app/1.0"


async def make_nws_request(url: str) -> dict[str, Any] | None:
    """Make a request to the NWS API with proper error handling."""
    headers = {
        "User-Agent": USER_AGENT,
        "Accept": "application/geo+json"
    }
    async with httpx.AsyncClient() as client:
        try:
            response = await client.get(url, headers=headers, timeout=30.0)
            response.raise_for_status()
            return response.json()
        except Exception:
            return None


def format_alert(feature: dict) -> str:
    """Format an alert feature into a readable string."""
    props = feature["properties"]
    return f"""
Event: {props.get("event", "Unknown")}
Area: {props.get("areaDesc", "Unknown")}
Severity: {props.get("severity", "Unknown")}
Description: {props.get("description", "No description available")}
Instructions: {props.get("instruction", "No specific instructions provided")}
"""


@mcp.tool()
async def get_alerts(state: str) -> str:
    """Get weather alerts for a US state.

    Args:
        state: Two-letter US state code (e.g. CA, NY)
    """
    url = f"{NWS_API_BASE}/alerts/active/area/{state}"
    data = await make_nws_request(url)

    if not data or "features" not in data:
        return "Unable to fetch alerts or no alerts found."

    if not data["features"]:
        return "No active alerts for this state."

    alerts = [format_alert(feature) for feature in data["features"]]
    return "\n---\n".join(alerts)


@mcp.tool()
async def get_forecast(latitude: float, longitude: float) -> str:
    """Get weather forecast for a location.

    Args:
        latitude: Latitude of the location
        longitude: Longitude of the location
    """
    points_url = f"{NWS_API_BASE}/points/{latitude},{longitude}"
    points_data = await make_nws_request(points_url)

    if not points_data:
        return "Unable to fetch forecast data for this location."

    forecast_url = points_data["properties"]["forecast"]
    forecast_data = await make_nws_request(forecast_url)

    if not forecast_data:
        return "Unable to fetch detailed forecast."

    periods = forecast_data["properties"]["periods"]
    forecasts = []
    for period in periods[:5]:
        forecast = f"""
{period["name"]}:
Temperature: {period["temperature"]}°{period["temperatureUnit"]}
Wind: {period["windSpeed"]} {period["windDirection"]}
Forecast: {period["detailedForecast"]}
"""
        forecasts.append(forecast)

    return "\n---\n".join(forecasts)


def main():
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
```

## Best Practices

1. **Use descriptive names**: Tool and server names should clearly indicate their purpose.

2. **Document parameters**: Always include type hints and docstrings so the LLM understands how to use your tools.

3. **Handle errors gracefully**: Return informative error messages rather than raising exceptions.

4. **Keep tools focused**: Each tool should do one thing well. If a tool becomes complex, consider breaking it into smaller tools.

5. **Use async for I/O**: Always use async functions for network requests or other I/O operations to prevent blocking.

## Next Steps

After implementing your server, see [configuration.md](configuration.md) to learn how to configure clients and test your server.
