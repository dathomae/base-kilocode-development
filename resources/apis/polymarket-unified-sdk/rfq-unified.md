# RFQ in the Unified SDK

The RFQ (Request for Quote) system in the unified SDK provides an event-driven interface for OTC-style trading. Unlike the `py-clob-client-v2` procedural RFQ client, the unified SDK uses an `RfqSession` abstraction that emits typed events.

## Key Types

```python
from polymarket.rfq import (
    RfqSession,              # Main RFQ session manager
    RfqEvent,               # Base event class (union)
    RfqQuoteRequestEvent,   # New quote request received
    RfqConfirmationRequestEvent,  # Quote accepted, needs confirmation
    RfqConfirmation,        # Accept/reject a confirmation request
    RfqConfirmationDecision,  # ACCEPT or REJECT
    RfqExecutionUpdateEvent,  # Trade execution status update
    RfqTradeEvent,          # Trade completed
    RfqDirection,           # BUY or SELL
    RfqSide,                # MAKER or TAKER
    RfqExecutionStatus,      # PENDING, CONFIRMED, etc.
    RfqId,                  # RFQ request ID type
    RfqQuoteId,             # Quote ID type
    RfqQuoteReference,      # Quote reference for linking
    RfqRequestedSize,       # Requested size in RFQ
    RfqRequestedSizeUnit,   # Unit (TOKEN or USDC)
    RfqQuoteSource,         # Source of the quote
    RfqErrorCode,           # RFQ error codes
    RfqCancelQuoteAck,      # Quote cancellation acknowledgment
)
```

## Event Types

| Event | Description |
|---|---|
| `RfqQuoteRequestEvent` | A new quote request is available |
| `RfqConfirmationRequestEvent` | A quote has been accepted and needs confirmation |
| `RfqExecutionUpdateEvent` | Trade execution progress update |
| `RfqTradeEvent` | A trade has been completed |

## Usage Flow

Since the RFQ module is in active development, check the [py-sdk repository](https://github.com/Polymarket/py-sdk) for the latest session-flow patterns and examples.

```python
from polymarket import SecureClient
from polymarket.rfq import RfqSession, RfqDirection

client = SecureClient.create(private_key=os.environ["PRIVATE_KEY"])

# Create an RFQ session (conceptual — check latest API)
session = RfqSession(client)

# Listen for events
for event in session.events():
    match event:
        case RfqQuoteRequestEvent():
            # Respond with a quote
            ...
        case RfqConfirmationRequestEvent():
            # Confirm or reject
            session.respond(event.request_id, RfqConfirmationDecision.ACCEPT)
        case RfqTradeEvent():
            print(f"Trade completed: {event.trade_id}")
```

## Differences from py-clob-client-v2 RFQ

| Aspect | py-clob-client-v2 | Unified SDK |
|---|---|---|
| Pattern | Procedural method calls | Event-driven session |
| Types | Dataclasses | Pydantic models |
| Match types | COMPLEMENTARY, MINT, MERGE | Integrated into events |
| Response model | Direct API dicts | Typed acknowledgement types |
