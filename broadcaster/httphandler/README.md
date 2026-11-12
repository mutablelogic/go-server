# Broadcaster API

## Change Operations

Changes are streamed to a client as they happen, such as to the rows of the server's tables.

- Each change has an `action`, which is `create`, `update` or `delete`.
- It has the `schema` and `table` which changed, and the `row`, or rows, which changed, such as by an identifier, when it says.
- It has `data` when the table says more of the change, such as a row's new serial, `{"serial":3}`.
- The server says which schemas and tables it publishes changes of.

### GET /changes

Watch the changes, as a stream of server-sent events, until the client closes the connection.

- Each change is a `change` event, whose data is the change, as JSON, such as `{"action":"create","schema":"llm","table":"provider","row":"ollama"}`.
- The stream is watching once it has started, so a client which then loads what it shows misses no changes.
- When the server is stopping, a `shutdown` event is sent, and the stream ends, as there will be no more changes.
- A client which doesn't read its changes has up to 16 kept for it, and misses those after them.
- Returns 406 if the client doesn't accept `text/event-stream`, or 503 if the server is stopping.
