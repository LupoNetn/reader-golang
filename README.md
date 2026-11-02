# Readr 📖

> **A 100% Deterministic, Offline, Zero-AI Passage Retrieval Engine.**  
> Verbatim document search with sub-millisecond BM25 scoring, real-time WebSocket streaming, and zero cloud dependencies.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-16.4-black?logo=next.js&logoColor=white)](https://nextjs.org)
[![React](https://img.shields.io/badge/React-19.3-61DAFB?logo=react&logoColor=black)](https://react.dev)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind-4.0-38B2AC?logo=tailwind-css&logoColor=white)](https://tailwindcss.com)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## ⚡ Highlights

- **Zero LLM Hallucinations**: 100% deterministic passage retrieval. Answers are never synthesized or paraphrased; Readr returns verbatim excerpts directly from your source files.
- **Sub-Millisecond Execution**: Queries evaluate in `< 1ms` in memory using an optimized BM25 ranking algorithm with keyword boost and stopword filtering.
- **Strict Verification Guard**: If no candidate excerpt passes the mathematical relevance threshold (confidence $\ge 0.30$, coverage $\ge 40\%$), Readr strictly returns `not_found`.
- **Full Data Sovereignty**: Air-gapped on `localhost:3030`. No external API keys, no tracking, zero third-party cloud data transmission.
- **Full-Duplex WebSockets**: High-throughput JSON-RPC communication powering real-time status updates, chunk streaming, and 22ms typewriter citation rendering.
- **Atomic Persistence**: In-memory store protected by `sync.RWMutex` with atomic write-and-rename disk synchronization.

---

## 🏗️ System Architecture

Readr is divided into two distinct components:

```
readr-golang/
├── backend/                  # Go deterministic retrieval engine
│   ├── cmd/server/           # HTTP & WebSocket server entrypoint
│   ├── data/                 # JSON store persistence (store.json)
│   ├── internal/
│   │   ├── engine/           # BM25 ranking, tokenizer, in-memory store
│   │   ├── parsers/          # Text, Markdown, and PDF document parsers
│   │   └── websocket/        # Hub, client manager, and JSON-RPC protocol
│   ├── .env                  # Environment configuration (gitignored)
│   └── .gitignore            # Backend-specific ignore rules
│
└── frontend/                 # Next.js 16 (App Router) + Tailwind v4 UI
    ├── src/
    │   ├── app/              # Next.js root layout and pages
    │   ├── components/       # Search interface, typewriter, citation cards
    │   └── lib/              # WebSocket client & JSON-RPC bindings
    └── .gitignore            # Frontend-specific ignore rules
```

---

## 🚀 Quick Start

### 1. Prerequisites
- **Go**: `1.22+` (tested with Go 1.25)
- **Node.js**: `18+` or `20+`
- **npm** or **pnpm** / **yarn**

---

### 2. Backend Setup (Go Engine)

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```

2. Configure environment variables (optional, defaults to `3030`):
   ```bash
   # backend/.env
   PORT=3030
   ```

3. Download dependencies & start the server:
   ```bash
   go mod download
   go run ./cmd/server
   ```

The backend starts listening on `http://localhost:3030` with the WebSocket endpoint at `ws://localhost:3030/ws`.

---

### 3. Frontend Setup (Next.js)

1. Open a new terminal and navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

4. Open your browser and navigate to **`http://localhost:3000`**.

---

## 🔌 API & WebSocket Protocol

Readr uses JSON-RPC over WebSockets for bi-directional communication.

### Endpoints
| Protocol | Path | Description |
|---|---|---|
| `GET` | `/health` | Server status, document count, and indexed chunks |
| `WS` | `/ws` | Full-duplex JSON-RPC 2.0 communication channel |

### WebSocket Methods
- `documents/index`: Submit and parse raw text or document paths.
- `queries/retrieve`: Execute a deterministic BM25 search query across all chunks.
- `history/list`: Retrieve query execution logs and previous results.
- `notifications/progress`: Real-time streaming status during active scans.

---

## 🧪 Testing

To run the backend test suite across the search engine, parsers, and WebSocket handlers:

```bash
cd backend
go test -v ./...
```

---

## 📄 License

This project is licensed under the MIT License.
