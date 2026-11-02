# Readr Deterministic Retrieval Architecture

## 1. Engine Specifications
The Readr passage retrieval system replaces opaque probabilistic vector databases with an exact mathematical BM25 ranking engine written in pure Golang. The scoring parameters are locked to k1 = 1.5 and b = 0.75, with length normalization calculated dynamically against the corpus average document length (avgdl).

## 2. Ingestion Pipeline and Chunker
Incoming text streams are stripped of standard English stopwords using a fast in-memory map filter. Sentences are delimited along terminal punctuation boundaries (. ? !). When any individual paragraph exceeds 600 characters, it undergoes recursive sentence splitting to ensure chunks remain between 200 and 600 characters without stripping source words.

## 3. WebSocket Protocol Specs
Communication between the Next.js 16 frontend and the Go engine runs over a single full-duplex WebSocket connection mounted at `ws://localhost:3030/ws`. Heartbeat frames are exchanged via 25-second ping/pong cycles. All requests follow JSON-RPC 2.0 specifications. Prior to running passage ranking, the engine emits an intermediate `notifications/progress` frame with a progressToken and a status message reading "Scanning chunks...".

## 4. Minimum Confidence Guards
To prevent hallucinated or spurious results, two mathematical guards are strictly enforced:
- Minimum confidence score threshold: 0.30
- Minimum keyword match ratio: 0.40 (at least 40% of non-stopword query tokens must appear verbatim in candidate passages)
If no candidate chunk meets both criteria, the engine deterministically returns status `not_found`.

