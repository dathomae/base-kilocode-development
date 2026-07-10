# Signing & Headers — L1 and L2 Authentication

## Signer

The `Signer` class wraps `eth_account.Account` for ECDSA signing:

```python
from py_clob_client_v2.signer import Signer

signer = Signer(private_key, chain_id=137)
signer.address()       # -> str (Ethereum address)
signer.get_chain_id()  # -> int
signer.sign(message_hash)  # -> hex signature string
```

The `Signer` is required for:
1. L1 authentication (EIP-712 signed header for creating/deriving API keys)
2. Order signing (EIP-712 signed order payloads)

It is **not** used for L2 authentication headers — those use HMAC with API credentials.

## Header Generation

### `headers/headers.py`

Two functions generate headers for authenticated requests:

```python
from py_clob_client_v2.headers.headers import create_level_1_headers, create_level_2_headers
```

### L1 Headers — `create_level_1_headers(signer, nonce, timestamp) -> dict`

Returns:
```python
{
    "POLY_ADDRESS": signer.address(),
    "POLY_SIGNATURE": "<EIP-712 signature of ClobAuth struct>",
    "POLY_TIMESTAMP": timestamp,
    "POLY_NONCE": nonce or 0,
}
```

**EIP-712 domain** for `ClobAuth`:

```python
domain = {
    "name": "ClobAuthDomain",
    "version": "1",
    "chainId": chain_id,
}

types = {
    "ClobAuth": [
        {"name": "address", "type": "address"},
        {"name": "timestamp", "type": "string"},
        {"name": "nonce", "type": "uint256"},
        {"name": "message", "type": "string"},
    ]
}

value = {
    "address": signer.address(),
    "timestamp": str(ts),
    "nonce": nonce or 0,
    "message": "This message attests that I control the given wallet",
}
```

Used for endpoints:
- `POST /auth/api-key` (create API credentials)
- `GET /auth/derive-api-key` (derive API credentials)

### L2 Headers — `create_level_2_headers(signer, creds, request_args, timestamp) -> dict`

Returns:
```python
{
    "POLY_ADDRESS": signer.address(),
    "POLY_SIGNATURE": "<HMAC-SHA256 signature>",
    "POLY_TIMESTAMP": timestamp,
    "POLY_API_KEY": creds.api_key,
    "POLY_PASSPHRASE": creds.api_passphrase,
}
```

The HMAC-SHA256 signature is created using:
- **Secret**: base64-decoded `creds.api_secret`
- **Message**: `timestamp + method + request_path + serialized_body`

Used for all trading endpoints:
- `POST /order`, `GET /orders`, `DELETE /order`
- `GET /balance-allowance`, `GET /trades`
- `POST /orders`, `DELETE /cancel-all`
- API key management endpoints
- Builder/RFQ endpoints

## L2 Signing in ClobClient

The `ClobClient._l2_headers(method, endpoint, body, serialized_body)` method:
1. Checks level 2 auth availability
2. Creates `RequestArgs(method, request_path, body, serialized_body)`
3. Calls `create_level_2_headers(signer, creds, request_args, timestamp)`

The `serialized_body` parameter allows the client to pass a custom serialization (e.g., `json.dumps(body, separators=(",", ":"), ensure_ascii=False)`) for compact, deterministic JSON.

## Signing Directory

### `signing/eip712.py`
EIP-712 typed data signing implementation. Used for:
- L1 authentication message signing
- Order data signing (via `order_utils` modules)

### `signing/hmac.py`
HMAC-SHA256 signing implementation for L2 authentication headers.

## Order Signing (EIP-712)

Orders are signed using EIP-712 typed data. The signing happens in `order_utils/` via:
- `ExchangeOrderBuilderV1.build_signed_order()` — signs `OrderDataV1`
- `ExchangeOrderBuilderV2.build_signed_order()` — signs `OrderDataV2`

Both use different EIP-712 domain definitions based on the exchange contract version. The order struct includes token IDs, amounts, expiration, and exchange-specific fields.

## Key Security Points

- **L1 auth**: The private key is used to sign an EIP-712 message proving ownership. The private key never leaves the client.
- **L2 auth**: API credentials (key, secret, passphrase) authenticate requests via HMAC. The secret is base64-decoded for signing.
- **Order signing**: Every order must be signed by the private key with an EIP-712 signature. Even with L2 credentials, order creation requires the private key.
- **No custody**: The private key is used for signing only; Polymarket never takes custody of funds.
