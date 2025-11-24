# Pulse API Client (Alpha)

<img width="1022" height="269" alt="image" src="https://github.com/user-attachments/assets/0d552cc5-62bb-4cbb-a18e-69159ef64d31" />

Pulse is a **performance-focused**, **local-first**, **developer-first** API client built for modern development workflows. Built with Wails - Go + Svelte + TypeScript!

It unifies classic HTTP testing, gRPC, real-time streaming protocols, and event-driven systems into one clean, fast, desktop-native tool.

Currently fully supports - HTTP, gRPC, Kafka, WebSockets, SSE, gRPC Streams.

This project is **actively in development** and currently in **Alpha**.  
Breaking changes, UI shifts, and feature overhauls will happen frequently.

<img width="1519" height="954" alt="image" src="https://github.com/user-attachments/assets/4a8ad249-e3f5-4a31-9595-c26be276eadf" />


---

## ✨ Key Concepts

### Unified Workspaces
- Multiple custom **workspaces**
- Each workspace keeps its own:
  - Environments
  - Variables
  - Collections
  - History
  - UI layout state
- Fully locally persisted and securely encoded

### Environments & Variables
- Create unlimited custom environments
- Define variables (string, JSON, secret)
- Variable substitution in URLs, headers, bodies, and streaming tools
- Automatically saved and restored

### Collections
- Save HTTP requests, gRPC calls, streaming connections
- Folder support, drag & drop reordering
- Execute saved requests instantly
- Fully persisted on disk

### History
- Every request/connection is tracked:
  - Timestamp
  - Method or protocol
  - URL/topic/service
- Encoded local storage ensures fast access and privacy

---

## 🌐 Protocol Support (Alpha)

Pulse already supports a wide set of protocols, with more being built.

---

## 1. HTTP Client

### Features
- Full request builder:
  - Method, URL, params
  - Headers
  - Authentication helpers
  - Request body (raw, JSON, form-data)
- Syntax-highlighted editors
- Pretty JSON viewer for responses
- Response details:
  - Status code
  - Headers
  - Timing
  - Size
- Save requests to collections
- Fully compatible with environment variables

---

## 2. gRPC (Unary)

### Features
- Upload `.proto` files
- Paste raw proto definitions
- Reflection mode (where server supports it)
- Auto-generate:
  - Services
  - Methods
  - Types
- Editor for request messages
- Response previewer with:
  - Message tree
  - JSON view
  - Raw proto view

---

## 3. Streaming Lab (Unified Streaming Window)

A single tab where every real-time or bi-directional protocol lives together.

Supported right now:

### WebSocket
- Custom subprotocols
- Auto reconnect
- Ping/pong settings
- Custom headers
- Send custom messages (JSON, text, binary)
- Color-coded inbound/outbound messages
- Timestamping
- Pausing & filtering message flow

### Server-Sent Events (SSE)
- Include credentials (cookies)
- Auto reconnect
- Retry timeout
- Resume from Event ID
- Event-type filters
- Custom headers

### gRPC Streaming
- Upload/paste protos or use reflection
- Full stream support:
  - Server streaming
  - Client streaming
  - Bi-directional streaming
- Auto-generated message editors
- Structured message inspector
- Dedicated logs for stream events

---

## 4. Kafka (Alpha)

Full Kafka integration, designed for real developer debugging.

### Connection
- Bootstrap servers
- Client ID
- Connection timeout
- SASL authentication:
  - PLAIN
  - SCRAM–SHA-256
  - SCRAM–SHA-512
- TLS/SSL:
  - Custom certs
  - Skip verification option

### Topic Explorer
- Auto-loads all topics
- Partition counts
- Refresh button

### Consumer Panel
- Consumer groups
- Offset strategies:
  - Latest
  - Earliest
  - Custom offsets
- Auto-commit toggle
- Start/Stop controls
- Live message log:
  - Key
  - Value
  - Headers
  - Timestamp
  - Partition
  - Offset

### Producer Panel
- Select topic & partition
- Key & value fields
- Custom headers
- Compression:
  - none
  - gzip
  - snappy
  - lz4
  - zstd
- Acks:
  - 0
  - 1
  - all

### Persistence
- All Kafka settings are saved per tab
- Collapsible panels for a clean UI

---

## 🧩 Future Roadmap

### Coming Soon
- MQTT (WS + backend TCP)
- SSE replay mode
- Secrets manager and 3rd party secrets managers integration
- GraphQL explorer (query editor + schema browser)
- TCP/UDP raw socket inspector
- Redis streams
- PostgreSQL logical replication stream viewer
- gRPC metadata inspector
- AI-assisted request generation
- Multi-window mode
- Plugin system

### Planned Later
- Team sync features
- Cloud workspace backup
- Browser extension for capturing requests
- API testing automation runner
- Scripting

---

## 🛠️ Technology Stack

### Frontend
- **Svelte + TypeScript**
- Wails-native bridge
- Custom JSON viewer
- Monaco-powered editors
- Reactive stores for workspace state

### Backend
- **Go**
- HTTP proxying
- WebSocket/SSE relays
- gRPC client/streaming engine
- Kafka consumer/producer pipeline
- Local encrypted storage

---

## 🚧 Alpha Disclaimer

Pulse is **not production-ready yet**.  
Features may break, APIs may change, and a lot of UI polish is still in progress.

Feedback, issues, and feature requests are welcome.

---

## 🧭 Project Status

`Alpha – Actively Developing`

New features and big changes arrive weekly. Stay tuned.
