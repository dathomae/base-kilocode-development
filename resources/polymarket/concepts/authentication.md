# Authentication

Two-level authentication for the Polymarket CLOB API.

## Public vs Authenticated

| Access Level | APIs | Description |
|---|---|---|
| **Public** | Gamma API, Data API, CLOB read endpoints | No authentication required |
| **Authenticated** | CLOB trading endpoints | Requires all 5 `POLY_*` L2 HTTP headers |

## Two-Level Authentication Model

### L1 Authentication (Private Key)

Uses the wallet's private key to sign an EIP-712 message proving ownership. The private key stays in user control; all trading is non-custodial.

**Used for:**
- Creating API credentials
- Deriving existing API credentials
- Signing and creating orders locally

### L2 Authentication (API Key)

Uses API credentials (apiKey, secret, passphrase) generated from L1 authentication. Requests are signed using HMAC-SHA256.

**Used for:**
- Cancel or get open orders
- Check balances and allowances
- Post signed orders

**Note**: Even with L2 headers, creating orders requires the user to sign the order payload with their private key.

## Getting API Credentials

### Using the Python SDK

```python
from py_clob_client_v2 import ClobClient
import os

client = ClobClient(
    host="https://clob.polymarket.com",
    chain_id=137,  # Polygon mainnet
    key=os.getenv("PRIVATE_KEY")
)

# Creates new credentials or derives existing ones
credentials = client.create_or_derive_api_key()
# Returns: { "apiKey": "...", "secret": "...", "passphrase": "..." }
```

### Using REST API

**Create**: `POST https://clob.polymarket.com/auth/api-key`
**Derive**: `GET https://clob.polymarket.com/auth/derive-api-key`

Required L1 headers:

| Header | Description |
|---|---|
| `POLY_ADDRESS` | Polygon signer address |
| `POLY_SIGNATURE` | CLOB EIP-712 signature |
| `POLY_TIMESTAMP` | Current UNIX timestamp |
| `POLY_NONCE` | Nonce (default: 0) |

The `POLY_SIGNATURE` is an EIP-712 signature of the `ClobAuth` struct over the `ClobAuthDomain`:

```python
domain = {
    "name": "ClobAuthDomain",
    "version": "1",
    "chainId": 137,
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
    "address": signingAddress,
    "timestamp": ts,
    "nonce": nonce,
    "message": "This message attests that I control the given wallet",
}
```

## L2 Authentication Headers

All trading endpoints require these 5 headers:

| Header | Description |
|---|---|
| `POLY_ADDRESS` | Polygon signer address |
| `POLY_SIGNATURE` | HMAC-SHA256 signature of the request |
| `POLY_TIMESTAMP` | Current UNIX timestamp |
| `POLY_API_KEY` | API `apiKey` value |
| `POLY_PASSPHRASE` | API `passphrase` value |

The `POLY_SIGNATURE` for L2 is an HMAC-SHA256 created using the API `secret` value, signing the request method, path, and body.

## Signature Types and Funder

When initializing the L2 client, specify **signatureType** and **funder** address:

| Type | Value | Description | Funder |
|---|---|---|---|
| **EOA** | `0` | Standard Ethereum wallet (MetaMask). Needs POL for gas. | EOA address |
| **POLY_PROXY** | `1` | Existing Polymarket proxy wallet flow (Magic Link). | Proxy address |
| **GNOSIS_SAFE** | `2` | Existing Gnosis Safe wallet flow. | Safe address |
| **POLY_1271** | `3` | Deposit wallet flow for **new API users**. Orders validated via ERC-1271. | Deposit wallet address |

**New API users should use deposit wallets with `POLY_1271` (signature type `3`).**

## Security Best Practices

- Never expose private keys — store in environment variables or secure key management
- Implement request signing on the server — never expose API secret in client-side code
- Never commit private keys to version control

## Troubleshooting

| Error | Cause | Solution |
|---|---|---|
| `INVALID_SIGNATURE` | Incorrect or improperly formatted private key | Verify key is valid hex string (starts with "0x") |
| `NONCE_ALREADY_USED` | Nonce already used to create API key | Use `deriveApiKey()` or a different nonce |
| `Invalid Funder Address` | Funder doesn't match wallet | Check profile address at polymarket.com/settings |
| Lost credentials + nonce | Cannot recover | Create new credentials with `createApiKey()` |
