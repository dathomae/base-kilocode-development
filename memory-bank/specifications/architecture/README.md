# Architecture

Architecture records component placement, responsibilities, explicit non-ownership, interfaces, and
failure behavior. Each component should have an overview and a contract. The architecture is complete
enough for design work only when ownership is non-overlapping and every interface covers important
error, authorization, retry, idempotency, and partial-failure cases.
