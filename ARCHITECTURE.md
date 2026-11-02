# Readr: Technical Architecture & Systems Engineering Manual

> **A 100% Deterministic, Zero-AI Passage Retrieval Engine**  
> Built with Golang (Go 1.22+), Next.js 16 (React 19 + Tailwind CSS), and Full-Duplex WebSockets (JSON-RPC 2.0 / MCP Protocol).

---

## 1. Executive Summary & Zero-AI Retrieval Philosophy

### The Problem with Generative / Vector RAG
Modern document question-answering systems heavily rely on Large Language Models (LLMs) and vector databases. While powerful, they suffer from critical architectural drawbacks:
1. **Hallucination & Fabrication**: Models synthesize answers, often inventing facts or blending disparate context clauses.
2. **Probabilistic Uncertainty**: Cosine similarity over floating-point embeddings is non-deterministic; changing a single word can drastically alter the retrieved chunks.
3. **High Latency & Costs**: External API roundtrips take 1,500ms to 4,000ms and cost money per million tokens.
4. **Third-Party Data Exposure**: Documents leave your machine and travel to proprietary cloud providers.

### The Readr Solution
**Readr** abandons opaque embeddings and LLMs completely. It is a **100% deterministic, offline, sub-millisecond passage retrieval engine**:
* **Verbatim Extraction**: Answers are never paraphrased or generated. Readr locates the literal source excerpt from the source file. If no passage satisfies the mathematical threshold, it strictly returns `"not_found"`.
* **Instant Speed**: Queries execute in **< 1ms** directly against an in-memory inverted index guarded by read-write mutexes.
* **Full Data Sovereignty**: Operates completely air-gapped on `localhost:3030`. Zero tokens, zero external API keys, zero cloud dependencies.

---

## 2. End-to-End System Topology

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                               NEXT.JS 16 FRONTEND (Port 3000)                               │
│  ┌─────────────────────────┐  ┌───────────────────────────┐  ┌───────────────────────────┐  │
│  │ TopNav & Sidebar        │  │ 22ms Typewriter Engine    │  │ Frameless Prose & Drawers │  │
│  └───────────┬─────────────┘  └─────────────┬─────────────┘  └─────────────┬─────────────┘  │
│              └──────────────────────────────┼──────────────────────────────┘                │
│                                             ▼                                               │
│                            ┌───────────────────────────────────┐                            │
│                            │   McpWebSocketClient (lib/ws)     │                            │
│                            │   • JSON-RPC 2.0 Multiplexing     │                            │
│                            │   • Exponential Backoff Reconnect │                            │
│                            └────────────────┬──────────────────┘                            │
└─────────────────────────────────────────────┼───────────────────────────────────────────────┘
                                              │
                    Full-Duplex TCP / WS      │ JSON-RPC 2.0 Frames over
                    ws://localhost:3030/ws    │ Ping/Pong Heartbeats (25s)
                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                GOLANG BACKEND (Port 3030)                                   │
│  ┌───────────────────────────────────────────────────────────────────────────────────────┐  │
│  │ WebSocket Hub & Protocol Dispatcher (internal/websocket)                              │  │
│  │ • Client Read/Write Pumps (Goroutines)  • JSON-RPC 2.0 Method Routing                 │  │
│  │ • tools/call: ask_document, ingest_document, delete_document                            │  │
│  │ • resources/read: readr://documents, readr://history                                  │  │
│  └──────────────────────────────────────────┬────────────────────────────────────────────┘  │
│                                             │                                               │
│                      ┌──────────────────────┴──────────────────────┐                        │
│                      ▼                                             ▼                        │
│  ┌──────────────────────────────────────┐     ┌──────────────────────────────────────────┐  │
│  │ Ingestion & Parsers                  │     │ Mathematical BM25 Engine                 │  │
│  │ • PDF: Pure Go stream extraction     │     │ • Parameters: k1 = 1.5, b = 0.75         │  │
│  │ • CSV: Header-to-cell token binding  │     │ • Dynamic Corpus avgdl Calculation       │  │
│  │ • JSON: Key-expansion formatting     │     │ • IDF: ln((N - n(q) + 0.5)/(n(q)+0.5)+1) │  │
│  │ • Regex Sentence Chunker (200-600c)  │     │ • Minimum Confidence Guard: >= 0.30      │  │
│  │ • Stopword Filter (map[string]struct)│     │ • Keyword Coverage Guard: >= 40%         │  │
│  └──────────────────┬───────────────────┘     └────────────────────┬─────────────────────┘  │
│                     │                                              │                        │
│                     └───────────────────────┬──────────────────────┘                        │
│                                             ▼                                               │
│  ┌───────────────────────────────────────────────────────────────────────────────────────┐  │
│  │ Thread-Safe Memory Store (internal/engine/store.go)                                   │  │
│  │ • In-Memory sync.RWMutex Storage (Sub-millisecond slices)                             │  │
│  │ • Atomic Disk Flush to backend/data/store.json (Temp write + atomic rename)           │  │
│  │ • Startup Auto-Seeding from 'text files/' folder                                      │  │
│  └───────────────────────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Storage & Concurrency Architecture

### In-Memory Concurrency (`sync.RWMutex`)
Readr holds all parsed document chunks and query history directly in RAM. Go's memory layout ensures that iterating over chunks during BM25 evaluation takes microseconds without heap thrashing.

```go
type Store struct {
    mu        sync.RWMutex
    dataPath  string
    documents map[string]DocumentMeta
    chunks    []DocumentChunk
    history   []ChatHistoryItem
}
```

* **Read Operations (`RLock()`)**: Multiple WebSocket clients can query the store simultaneously without contention. Goroutines read slices concurrently during searches.
* **Write Operations (`Lock()`)**: Document ingestion, deletion, and history appending acquire an exclusive lock, perform in-memory mutations, and trigger an atomic disk flush.

### Write-Through Atomic Disk Persistence
To prevent corrupt files during unexpected crashes or power loss, persistence uses a **temp-file-write-and-atomic-rename** strategy:

1. Data is serialized to formatted JSON in memory.
2. Written to a temporary file: `data/store.json.tmp.<nanoseconds>`.
3. Flushed and closed on disk.
4. Atomically renamed to `data/store.json` using `os.Rename`.

```go
// Atomic disk flush guarantee
tmpFile := fmt.Sprintf("%s.tmp.%d", s.dataPath, time.Now().UnixNano())
if err := os.WriteFile(tmpFile, dataBytes, 0644); err != nil {
    return err
}
return os.Rename(tmpFile, s.dataPath) // Atomic replace on POSIX & Windows
```

---

## 4. Ingestion & Document Parsing Pipeline

Readr normalizes diverse file types into token-dense textual representations before chunking:

### 1. Pure-Go PDF Parsing (`github.com/ledongthuc/pdf`)
* Decodes incoming Base64 binary streams into raw byte buffers.
* Extracts plain text page by page with zero external C/C++ dependencies (`poppler`, `pdf2text`, or `ghostscript` are **not** needed).
* Runs seamlessly in cross-platform environments.

### 2. CSV Column-to-Cell Token Binding
In traditional naive chunkers, CSV rows are parsed as comma-separated values (`Elena Rostova, Zurich, 195000`). If a user queries *"What is Elena's salary?"*, the word *"salary"* is absent from the row.
Readr solves this by **binding every column header to its cell value**:
```text
Employee ID: EMP-1042, Full Name: Elena Rostova, Job Title: Principal Systems Architect, Office Location: Zurich, Annual Salary: 195000
```
Headers undergo automatic CamelCase and snake_case splitting (`AnnualSalary` $\rightarrow$ `Annual Salary`), turning column headers into searchable tokens.

### 3. Structured JSON Key Expansion
JSON arrays and objects are flattened into descriptive key-value pairs with normalized labels (`max_concurrent_websockets` $\rightarrow$ `max concurrent websockets: 250000`), allowing users to search by field names or values.

### 4. Sentence & Paragraph Chunker
* Normalizes line endings (`\r\n` $\rightarrow$ `\n`).
* Splits text on paragraph breaks (`\r?\n+`).
* When paragraphs exceed 600 characters, splits along terminal sentence boundaries (`[.?!](\s+|$)`).
* Avoids PCRE lookbehinds (which Go's standard library does not support) while ensuring sentences remain intact between 200 and 600 characters.

---

## 5. Mathematical BM25 Ranking Engine

Readr uses Okapi BM25, the industry-standard probabilistic term-frequency ranking algorithm:

$$\text{Score}(D, Q) = \sum_{q \in Q} \text{IDF}(q) \cdot \frac{f(q, D) \cdot (k_1 + 1)}{f(q, D) + k_1 \cdot \left(1 - b + b \cdot \frac{|D|}{\text{avgdl}}\right)}$$

### Parameters & Constants
* **$k_1 = 1.5$**: Controls term frequency saturation. Prevents a chunk that repeats a single word 20 times from dominating the score.
* **$b = 0.75$**: Controls document length penalization relative to the corpus average document length ($avgdl$).
* **Corpus Statistics**: Dynamic calculation of total chunks $N$, chunk length $|D|$, and average length $avgdl$:
  $$\text{avgdl} = \frac{1}{N} \sum_{i=1}^N |D_i|$$

### Inverse Document Frequency (IDF)
Readr uses the standard Lucene/Robertson smoothed IDF formula:

$$\text{IDF}(q) = \ln\left(\frac{N - n(q) + 0.5}{n(q) + 0.5} + 1\right)$$

where $n(q)$ is the number of chunks containing query term $q$. If a term appears in every document, its IDF approaches zero; if it appears in only one chunk, its IDF is high.

### Deterministic Confidence Guards (Zero AI Slop)
To guarantee that the engine **never returns an irrelevant chunk**, two guards are enforced before a match is accepted:

1. **Keyword Match Ratio Guard ($\ge 40\%$)**:
   $$\frac{|\text{Matched Query Keywords}|}{|\text{Unique Non-Stopword Query Keywords}|} \ge 0.40$$
   At least 40% of the query's distinct keywords must appear verbatim in the candidate passage.
2. **Minimum Confidence Score Threshold ($\ge 0.30$)**:
   $$\text{BM25 Score} \ge 0.30$$

If no chunks satisfy both guards, the engine strictly emits:
```json
{
  "status": "not_found",
  "primaryMatch": null,
  "allMatches": []
}
```

---

## 6. Model Context Protocol (MCP) & Full-Duplex WebSockets

### Why Full-Duplex WebSockets Replaces SSE
* **Server-Sent Events (SSE)**: Unidirectional. The browser must open a new HTTP connection for every search query, sending request headers and incurring TCP handshake latency.
* **Full-Duplex WebSockets (`ws://localhost:3030/ws`)**: A single, persistent TCP connection stays open indefinitely. Lightweight JSON frames fly back and forth with near-zero overhead.

### Goroutine Client Hub Lifecycle
* Each connected client is managed by two dedicated Go routines:
  * **`readPump()`**: Reads incoming WebSocket frames up to 10MB, enforces pong deadlines (60s), and dispatches messages.
  * **`writePump()`**: Periodically sends ping frames (every 25s) and writes queued JSON-RPC responses.

### Protocol Framing (JSON-RPC 2.0 / MCP Standard)

#### 1. Real-Time Streaming Progress Frame
Before executing the BM25 search, the server immediately notifies the client:
```json
{
  "jsonrpc": "2.0",
  "method": "notifications/progress",
  "params": {
    "progressToken": "req-1730528145-a9b8c",
    "progress": 50,
    "message": "Scanning chunks..."
  }
}
```

#### 2. Final Result Frame
```json
{
  "jsonrpc": "2.0",
  "id": "req-1730528145-a9b8c",
  "result": {
    "query": "Which cluster supports 250000 concurrent websockets in Northern Virginia",
    "status": "found",
    "primaryMatch": {
      "chunkId": "cluster-chk-1",
      "filename": "infrastructure.json",
      "chunkIndex": 1,
      "content": "cluster id: us-east-prod-1, max concurrent websockets: 250000, region: Northern Virginia...",
      "score": 14.52,
      "matchPercentage": 86.0,
      "matchedKeywords": ["cluster", "250000", "concurrent", "websockets", "northern", "virginia"]
    },
    "allMatches": [...],
    "totalChunksSearched": 18,
    "latencyMs": 0.45
  }
}
```

#### 3. MCP Tool & Resource Schemas
Readr provides discovery endpoints:
* **`tools/list`**: Returns machine-readable schemas for `ask_document`, `ingest_document`, and `delete_document`.
* **`resources/list`**: Exposes `readr://documents` (library metadata) and `readr://history` (query logs).

---

## 7. Frontend Architecture (Next.js 16 + React 19)

### Monochromatic Monospace Design System
* **Color Discipline**: Strict `#000000` / `#ffffff` foundations with subtle `#262626` / `#e5e5e5` borders. No purple/blue AI gradients.
* **Typography**: Clean pairing of `Geist Sans` with monospace accents (`font-mono`) for queries, latency tags, match percentages, and code blocks.
* **Theme Engine**: Class-based dark mode (`.dark` on `html`) toggleable in the top navbar and persisted via `localStorage` with an inline anti-FOUC script in `layout.tsx`.

### WebSocket Multiplexing Client (`lib/wsClient.ts`)
Handles connection lifecycle outside the React render loop:
* Maintains a `Map<string, PendingRequest>` indexed by JSON-RPC `id`.
* Automatic exponential backoff reconnect:
  $$\text{Delay} = \min(1000 \cdot 2^n, 30000)\text{ ms}$$
* Intercepts `notifications/progress` frames to drive the UI progress bar.

### Verbatim Presentation & 22ms Typewriter Stream
* Answers are rendered as frameless natural prose (no clunky nested card containers).
* When a match arrives, text streams in at **22ms per word token** with a blinking terminal cursor (`typewriter-cursor`).
* **Citation Badge**: Displays file name, chunk index, match percentage, and matched keyword tags.
* **Expandable `[N more sources ▼]` Drawer**: Renders conflicting or alternative excerpts side-by-side with full source attribution.

---

## 8. Step-by-Step Lifecycle of a User Query

1. **User Types Query**: The user enters `"What are the liquidated damages for leaking the BM25 retrieval code?"` and presses `Enter`.
2. **WebSocket Dispatch**: `McpWebSocketClient.sendRequest` generates a unique ID (`req-1730528145-a9b8c`), registers a promise in `pendingRequests`, and writes a JSON-RPC 2.0 frame to the WebSocket.
3. **Go Server Receives Frame**: `readPump` reads the frame; `dispatcher.go` routes it to `handleAskDocument`.
4. **Intermediate Progress Emitted**: The server immediately transmits `{"method": "notifications/progress", "message": "Scanning chunks..."}`.
5. **Frontend Displays Progress**: The Next.js UI intercepts the notification and renders an animated progress banner with a pulsing indicator.
6. **BM25 Math Evaluation**: Go evaluates query tokens against all in-memory chunks:
   * Strips stopwords (`"what"`, `"are"`, `"the"`, `"for"`).
   * Evaluates keywords: `[liquidated, damages, leaking, bm25, retrieval, code]`.
   * Calculates term frequencies, document frequencies, and length normalization.
   * Ranks candidates and verifies confidence guards ($\text{score} \ge 0.30$, $\text{coverage} \ge 40\%$).
7. **Final JSON-RPC Response**: The server writes the winning excerpt and Top-K matches back to the WebSocket channel.
8. **Frontend Typewriter Rendering**: `McpWebSocketClient` matches the response ID, resolves the promise, and triggers the 22ms typewriter stream to display the verbatim source passage.

---

## 9. Verification & Operational Reference

### Quick Commands

```powershell
# 1. Start Go Backend (:3030)
cd c:\Users\HP\Documents\readr-golang\backend
go run ./cmd/server

# 2. Run Backend Unit & Race Test Suite
cd c:\Users\HP\Documents\readr-golang\backend
go test -v -race ./...

# 3. Start Frontend (:3000)
cd c:\Users\HP\Documents\readr-golang\frontend
npm run dev

# 4. Verify Frontend Production Build
cd c:\Users\HP\Documents\readr-golang\frontend
npm run build
```

### Git Repository
* **Remote**: [https://github.com/LupoNetn/reader-golang.git](https://github.com/LupoNetn/reader-golang.git)
* **Branch**: `main`
* **Status**: 100% committed, verified, and pushed.
