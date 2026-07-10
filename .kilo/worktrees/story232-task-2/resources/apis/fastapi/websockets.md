# WebSockets

FastAPI supports WebSockets for real-time bidirectional communication between clients and servers.

## Installation

To use WebSockets with FastAPI, install the `websockets` library:

```bash
pip install websockets
```

## Basic WebSocket Endpoint

### Creating a WebSocket Endpoint

Use the `@app.websocket()` decorator to create a WebSocket endpoint:

```python
from fastapi import FastAPI, WebSocket

app = FastAPI()

@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()
    while True:
        data = await websocket.receive_text()
        await websocket.send_text(f"Message text was: {data}")
```

### Key Steps

1. **Accept the connection**: `await websocket.accept()` must be called to establish the WebSocket connection
2. **Receive data**: Use `await websocket.receive_text()` to receive messages from the client
3. **Send data**: Use `await websocket.send_text()` to send messages to the client

## Complete Example with HTML Client

Here is a complete example with a simple HTML/JavaScript client:

```python
from fastapi import FastAPI, WebSocket
from fastapi.responses import HTMLResponse

app = FastAPI()

html = """
<!DOCTYPE html>
<html>
    <head>
        <title>Chat</title>
    </head>
    <body>
        <h1>WebSocket Chat</h1>
        <form action="" onsubmit="sendMessage(event)">
            <input type="text" id="messageText" autocomplete="off"/>
            <button>Send</button>
        </form>
        <ul id='messages'>
        </ul>
        <script>
            var ws = new WebSocket("ws://localhost:8000/ws");
            ws.onmessage = function(event) {
                var messages = document.getElementById('messages')
                var message = document.createElement('li')
                var content = document.createTextNode(event.data)
                message.appendChild(content)
                messages.appendChild(message)
            };
            function sendMessage(event) {
                var input = document.getElementById("messageText")
                ws.send(input.value)
                input.value = ''
                event.preventDefault()
            }
        </script>
    </body>
</html>
"""

@app.get("/")
async def get():
    return HTMLResponse(html)

@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()
    while True:
        data = await websocket.receive_text()
        await websocket.send_text(f"Message text was: {data}")
```

## WebSocket Methods

### Connection Management

- `await websocket.accept()`: Accept the WebSocket connection
- `await websocket.close()`: Close the WebSocket connection

### Receiving Data

- `await websocket.receive_text()`: Receive text data
- `await websocket.receive_bytes()`: Receive binary data
- `await websocket.receive_json()`: Receive JSON data (auto-parsed)

### Sending Data

- `await websocket.send_text(data)`: Send text data
- `await websocket.send_bytes(data)`: Send binary data
- `await websocket.send_json(data)`: Send JSON data (auto-serialized)

## WebSocket Client in Production

For production applications, you typically use:

- Modern frontend frameworks (React, Vue.js, Angular) with their WebSocket utilities
- Native mobile applications communicating directly with WebSocket endpoints
- Specialized WebSocket client libraries for your platform

The HTML/JavaScript example above is for demonstration purposes only and is not optimal for production use.

## Connection Handling

WebSocket connections are persistent. The `while True` loop keeps the connection open to handle multiple messages:

```python
@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()
    try:
        while True:
            data = await websocket.receive_text()
            # Process the data
            await websocket.send_text(f"Received: {data}")
    except WebSocketDisconnect:
        # Handle disconnection
        print("Client disconnected")
```

## References

- WebSockets Documentation: https://fastapi.tiangolo.com/advanced/websockets/
- WebSockets Library: https://websockets.readthedocs.io/
