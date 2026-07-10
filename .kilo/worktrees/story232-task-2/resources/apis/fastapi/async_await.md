# Concurrency and async / await

FastAPI supports asynchronous programming using Python's `async` and `await` syntax. This document explains how to use async/await in path operation functions and provides background on asynchronous code, concurrency, and parallelism.

## TL;DR

If you are using third-party libraries that tell you to call them with `await`:

```python
results = await some_library()
```

Then declare your path operation functions with `async def`:

```python
@app.get('/')
async def read_results():
    results = await some_library()
    return results
```

**Note**: You can only use `await` inside of functions created with `async def`.

If you are using a third-party library that does not support `await` (common with database libraries), use normal `def`:

```python
@app.get('/')
def results():
    results = some_library()
    return results
```

**Key point**: If your application does not need to communicate with anything else and wait for responses, use `async def` even without `await` inside.

## Asynchronous Code

Asynchronous code is code that can pause execution while waiting for slow I/O operations, allowing the computer to do other work during that time. This is particularly useful for web applications because:

- Web servers wait for clients to send requests over the network
- They wait for responses to be received by clients
- They perform operations like reading from databases or calling remote APIs

### Common I/O Operations

Asynchronous code typically involves waiting for:

- Data from clients through the network
- Data to be received by clients
- File contents to be read from disk
- Data to be written to disk
- Remote API operations
- Database operations

## Concurrency vs. Parallelism

### Concurrency

Concurrency means handling multiple tasks by interleaving them. While one task waits, another task can execute. This is like ordering at a fast-food restaurant:

1. You place your order
2. You receive a number and wait at a table
3. While waiting, you can do other things (like talking with friends)
4. When your number is called, you pick up your food

**Concurrency is ideal for I/O-bound operations** where there is significant waiting time between operations.

### Parallelism

Parallelism means executing multiple tasks simultaneously using multiple processors or cores. This is like having multiple cashiers at a restaurant, each taking and preparing orders independently.

**Parallelism is ideal for CPU-bound operations** like:

- Audio or image processing
- Computer vision
- Machine learning computations
- Deep learning model training

## Using async and await

### async def

Functions declared with `async def` are coroutines that can be paused and resumed:

```python
async def get_burgers(number: int):
    # Do some asynchronous stuff
    return burgers
```

### await

The `await` keyword tells Python to wait for an operation to complete before continuing:

```python
burgers = await get_burgers(2)
```

**Important rules:**

1. `await` can only be used inside `async def` functions
2. Functions defined with `async def` must be called with `await`
3. You cannot mix normal function calls with async functions without `await`

### Example in FastAPI

```python
@app.get('/burgers')
async def read_burgers():
    burgers = await get_burgers(2)
    return burgers
```

FastAPI handles the async execution automatically, so you do not need to worry about the "first" async function call.

## Technical Details

### Coroutines

A **coroutine** is the object returned by an `async def` function. Python knows it can start, pause, and resume coroutines when encountering `await` inside them. This functionality is comparable to Go's Goroutines.

### Path Operation Functions

When you declare a path operation function with normal `def` instead of `async def`, FastAPI runs it in an external threadpool that is then awaited. This prevents the function from blocking the server.

**Best practice**: Use `async def` unless your code performs blocking I/O operations.

### Mixing def and async def

You can mix `def` and `async def` in your path operation functions as needed. FastAPI handles both correctly and optimizes performance accordingly.

## Related Libraries

FastAPI and Starlette are based on **AnyIO**, making them compatible with both Python's standard library `asyncio` and Trio. You can also use **Asyncer** for combining async code with regular (blocking/synchronous) code.

## References

- Concurrency and async/await: https://fastapi.tiangolo.com/async/
- AnyIO: https://anyio.readthedocs.io/
- Asyncer: https://asyncer.tiangolo.com/
