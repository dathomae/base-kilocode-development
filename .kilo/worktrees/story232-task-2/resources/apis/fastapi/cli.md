# FastAPI CLI

FastAPI CLI is a command line program that serves FastAPI applications and manages FastAPI projects.

## Installation

When you install FastAPI with `pip install "fastapi[standard]"`, it includes a package called `fastapi-cli`. This package provides the `fastapi` command in the terminal.

## Commands

### `fastapi dev` - Development Mode

The `fastapi dev` command runs your FastAPI application in development mode with auto-reload enabled.

```bash
fastapi dev main.py
```

**Features:**

- **Auto-reload**: Automatically reloads the server when code changes are detected
- **Localhost binding**: Listens on IP address `127.0.0.1` (localhost only)
- **Development optimization**: Optimized for development with resource-intensive watching

**Important**: Auto-reload is resource-intensive and may be less stable than when disabled. Use it only for development.

**Example output:**

```
module  main.py
code    Importing the FastAPI app object from the module
app     Using import string: main:app
server  Server started at http://127.0.0.1:8000
server  Documentation at http://127.0.0.1:8000/docs
INFO    Will watch for changes in these directories: ['/home/user/code/awesomeapp']
INFO    Uvicorn running on http://127.0.0.1:8000 (Press CTRL+C to quit)
INFO    Started reloader process [383138] using WatchFiles
INFO    Started server process [383153]
INFO    Waiting for application startup.
INFO    Application startup complete.
```

### `fastapi run` - Production Mode

The `fastapi run` command runs FastAPI in production mode by default.

```bash
fastapi run
```

**Features:**

- **Auto-reload disabled**: More stable and efficient for production
- **Public binding**: Listens on IP address `0.0.0.0`, making the application accessible from all available IP addresses
- **Production-ready**: Designed for containerized deployment and production environments

**Note**: In most production scenarios, you should use a "termination proxy" handling HTTPS on top of your application. Your deployment provider may handle this, or you might need to configure it yourself.

## How It Works

FastAPI CLI takes the path to your Python program (e.g., `main.py`) and:

1. Automatically detects the FastAPI instance (commonly named `app`)
2. Determines the correct import process
3. Serves the application using Uvicorn

Internally, FastAPI CLI uses **Uvicorn**, a high-performance, production-ready ASGI server.

## Common Usage Patterns

### Running a specific file:

```bash
fastapi dev main.py
```

### Running in production:

```bash
fastapi run
```

## References

- FastAPI CLI Documentation: https://fastapi.tiangolo.com/fastapi-cli/
- Uvicorn: https://www.uvicorn.org/
