# Dependencies

FastAPI includes a powerful but intuitive **Dependency Injection** system that makes it easy to integrate components and share logic across path operation functions.

## What is Dependency Injection

**Dependency Injection** means that your code (path operation functions) can declare things it requires to work ("dependencies"), and FastAPI will automatically provide ("inject") those dependencies.

This approach:

- Reduces code duplication
- Makes testing easier
- Improves code organization
- Handles authentication, database connections, and validation automatically

## Creating Dependencies

A dependency is simply a function that FastAPI will call before executing your path operation function.

### Simple Dependency

```python
from fastapi import Depends, FastAPI

app = FastAPI()

async def common_parameters(q: str | None = None, skip: int = 0, limit: int = 100):
    return {"q": q, "skip": skip, "limit": limit}

@app.get("/items/")
async def read_items(commons: Annotated[dict, Depends(common_parameters)]):
    return commons

@app.get("/users/")
async def read_users(commons: Annotated[dict, Depends(common_parameters)]):
    return commons
```

## Using Depends

Import `Depends` from `fastapi` and declare dependencies in your path operation function parameters:

```python
from fastapi import Depends

@app.get("/items/")
async def read_items(commons: Annotated[dict, Depends(common_parameters)]):
    return commons
```

FastAPI will:

1. Analyze the dependency's parameters
2. Extract matching values from the request
3. Call the dependency function with those values
4. Return the result to your path operation function

## Dependencies with async

Dependencies can be declared with `async def` or normal `def`, just like path operation functions:

```python
async def async_dependency():
    return {"async": "data"}

def sync_dependency():
    return {"sync": "data"}

@app.get("/items/")
async def read_items(
    async_data: Annotated[dict, Depends(async_dependency)],
    sync_data: Annotated[dict, Depends(sync_dependency)]
):
    return {"async": async_data, "sync": sync_data}
```

## Dependencies that Return Values

Dependencies can return any Python object:

```python
from typing import Annotated
from fastapi import Depends, FastAPI

app = FastAPI()

class CommonQueryParams:
    def __init__(self, q: str | None = None, skip: int = 0, limit: int = 100):
        self.q = q
        self.skip = skip
        self.limit = limit

@app.get("/items/")
async def read_items(commons: Annotated[CommonQueryParams, Depends(CommonQueryParams)]):
    response = {}
    if commons.q:
        response.update({"q": commons.q})
    return response
```

## Sub-dependencies

Dependencies can have their own dependencies, creating a hierarchy:

```python
async def get_db():
    db = DBSession()
    try:
        yield db
    finally:
        db.close()

async def get_current_user(db: Annotated[DBSession, Depends(get_db)]):
    # Use the database session from get_db
    return db.query(User).first()

@app.get("/users/me")
async def read_current_user(user: Annotated[User, Depends(get_current_user)]):
    return user
```

## Dependencies with yield

Use `yield` in dependencies to execute cleanup code after the path operation finishes:

```python
async def get_db():
    db = SessionLocal()
    try:
        yield db  # Provide the dependency
    finally:
        db.close()  # Cleanup after request
```

This is particularly useful for:

- Database session management
- Resource cleanup
- Transaction handling

## Common Use Cases

### Database Sessions

```python
async def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
```

### Authentication

```python
async def get_current_user(token: Annotated[str, Depends(oauth2_scheme)]):
    user = verify_token(token)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid authentication")
    return user
```

### Validation

```python
def validate_item_id(item_id: int):
    if item_id < 0:
        raise HTTPException(status_code=400, detail="Item ID must be positive")
    return item_id
```

## References

- Dependencies Documentation: https://fastapi.tiangolo.com/tutorial/dependencies/
- FastAPI Dependency Injection: https://fastapi.tiangolo.com/tutorial/dependencies/dependencies-in-path-operation-decorators/
