# RFQ Client — Request for Quote Protocol

The `RfqClient` provides OTC-style trading via the Request for Quote (RFQ) protocol. It is accessible as `client.rfq` on any L2-authenticated `ClobClient`.

## Overview

RFQ enables peer-to-peer negotiation of trades:
1. A **requester** creates an RFQ request specifying what they want to trade
2. **Quoters** respond with quotes (offers)
3. The requester can accept a quote, and the quoter approves the resulting order
4. The trade is executed via a V1 signed order

All RFQ operations require L2 authentication.

## Core Types

### `RfqUserRequest`
Simplified input for creating a request:
- `token_id` — Token being traded
- `price` — Price per token (0 < price < 1)
- `side` — "BUY" or "SELL"
- `size` — Size in conditional tokens

### `RfqUserQuote`
Simplified input for creating a quote:
- `request_id` — ID of the RFQ request being quoted
- `token_id` — Token being traded
- `price` — Quoted price per token
- `side` — Quoter's side ("BUY" or "SELL")
- `size` — Size in conditional tokens

### `MatchType`
Enum indicating the type of trade match:
- `COMPLEMENTARY` — Standard opposite-side match (requester buys, quoter sells, or vice versa)
- `MINT` — Requester creates tokens (split from collateral via complement token)
- `MERGE` — Requester redeems tokens (merge to collateral via complement token)

### `GetRfqRequestsParams`
Filtering options for `get_rfq_requests()`:
- `request_ids` — Filter by specific IDs
- `state` — "active" or "inactive"
- `markets` — Filter by condition IDs
- `size_min`/`size_max` — Size range
- `size_usdc_min`/`size_usdc_max` — USDC size range
- `price_min`/`price_max` — Price range
- `sort_by` — "price", "expiry", "size", or "created"
- `sort_dir` — "asc" or "desc"
- `limit` — Pagination limit
- `offset` — Pagination cursor (base64)

### `GetRfqQuotesParams`
Similar filtering for quotes:
- `quote_ids` / `request_ids` — Filter by IDs
- `state`, `markets`, `size_*`, `price_*`, `sort_by`, `sort_dir`, `limit`, `offset`

## Request-Side Methods

### `create_rfq_request(user_request: RfqUserRequest, options=None) -> dict`
Creates and posts an RFQ request. Handles tick size resolution, price/size rounding, and conversion to onchain units.

### `cancel_rfq_request(params: CancelRfqRequestParams) -> str`
Cancels an active RFQ request by ID.

### `get_rfq_requests(params: GetRfqRequestsParams = None) -> dict`
Fetches RFQ requests with optional filtering. Returns paginated results.

## Quote-Side Methods

### `create_rfq_quote(user_quote: RfqUserQuote, options=None) -> dict`
Creates and posts a quote in response to an RFQ request.

### `get_rfq_requester_quotes(params: GetRfqQuotesParams = None) -> dict`
Gets quotes on requests created by the authenticated user (requester view).

### `get_rfq_quoter_quotes(params: GetRfqQuotesParams = None) -> dict`
Gets quotes created by the authenticated user (quoter view).

### `get_rfq_best_quote(params: GetRfqBestQuoteParams = None) -> dict`
Gets the best (cheapest buy / highest sell) quote for a request.

### `cancel_rfq_quote(params: CancelRfqQuoteParams) -> str`
Cancels an active RFQ quote by ID.

## Trade Execution Methods

### `accept_rfq_quote(params: AcceptQuoteParams) -> str`
**Requester side**: Accepts a quote. Fetches the quote, builds a V1 signed order matching it, and submits the acceptance. The acceptance includes the full signed order payload.

### `approve_rfq_order(params: ApproveOrderParams) -> str`
**Quoter side**: Approves a quote. Fetches the quote, builds a V1 signed order, and submits the approval.

Both methods use V1 orders (not V2) because the RFQ protocol requires the V1 order fields (taker, nonce, feeRateBps).

## Match Type Handling

For `COMPLEMENTARY` matches:
- Side is flipped (requester's BUY becomes quoter's SELL)
- Token is directly from the quote

For `MINT` and `MERGE` matches:
- Token is the `complement` token from the quote
- Price is inverted (`1 - quote_price`)
- The requester operates on the complement token

## Configuration

### `rfq_config() -> dict`
Gets RFQ configuration from the server.

## Response Types

### `RfqRequest`
Full RFQ request: `request_id`, `user_address`, `proxy_address`, `token`, `complement`, `condition`, `side`, `size_in`, `size_out`, `price`, `accepted_quote_id`, `state`, `expiry`, `created_at`, `updated_at`

### `RfqQuote`
Full RFQ quote: `quote_id`, `request_id`, `user_address`, `proxy_address`, `complement`, `condition`, `token`, `side`, `size_in`, `size_out`, `price`, `state`, `expiry`, `created_at`, `updated_at`

### `RfqPaginatedResponse`
Paginated list: `data: List`, `next_cursor`, `limit`, `count`, `total_count`
