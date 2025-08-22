---

## Introduction

In 2025, the intersection of **Go, AI, and Retrieval-Augmented Generation (RAG)** offers an opportunity to build practical, high-performance applications that bridge traditional software engineering and cutting-edge AI. My goal is to design and implement a hands-on project that helps me master these technologies while producing something useful and production-ready.

---

## Context

* **Why Go?**
  Go is fast, concurrent, and cloud-native friendly. It’s increasingly popular for building scalable backends and microservices. By using Go, I strengthen my systems engineering skills while staying close to production-grade practices.

* **Why RAG + LLMs?**
  LLMs are powerful but need grounding in private or domain-specific knowledge. A Retrieval-Augmented Generation system combines embeddings, vector search, and LLMs to deliver accurate, source-backed answers. This pattern is the backbone of modern AI-powered assistants and copilots.

* **Why this project?**
  Instead of just experimenting with APIs, I want to **design, build, and deploy** an end-to-end AI system:

  * Ingest and embed documents (PDF, Markdown, text).
  * Store embeddings in a vector database.
  * Query via natural language, retrieve relevant chunks, and generate answers with citations.
  * Scale to real-world usability with reranking, caching, monitoring, and deployment.

This approach forces me to learn the **full stack of AI engineering**: ingestion pipelines, embeddings, retrieval, LLM orchestration, evaluation, observability, and deployment.

---

# Week 1 — Scaffold & Ingestion (MVP data in, vectors out)

**Objectives**

* Set up a clean Go service.
* Ingest PDFs/Markdown → chunk → embed → store in a vector DB.

**Deliverables**

* Repo scaffold (see structure below).
* `/ingest` CLI that loads files from `data/` and upserts vectors.
* Vector DB running locally (Qdrant via Docker or pgvector in Postgres).

**Tasks**

* Init: `go mod init`, `.env` (API keys), Makefile.
* Libraries: HTTP (net/http or Fiber), PDF parsing (pdfcpu or uniDoc), tokenizer (simple char/word split), embeddings (OpenAI/Cohere/Gemini SDK), vector store (Qdrant/pgvector client).
* Chunking: 500–1,000 tokens; store text + metadata (doc\_id, page, chunk\_id).
* Upsert: batch inserts; keep original full text for grounding.

**Success criteria**

* Run `make ingest` → prints counts of chunks and vectors.
* 100% of sample docs embedded and visible in the DB.

**Stretch**

* Doc dedup on checksum.
* Basic ingestion metrics (chunks/sec, token count, cost estimate).

---

# Week 2 — Query & RAG Answering (question in, grounded answer out)

**Objectives**

* Implement the full RAG pipeline.
* Expose simple HTTP API.

**Deliverables**

* Endpoints:

  * `POST /query {query: string, top_k?: int}` → `{answer, citations:[{doc_id,page,chunk_id}]}`
  * `POST /upload` (optional) to ingest single files via API.
* Prompt template with system guardrails + citations.

**Tasks**

* Query flow: embed query → vector search (top\_k=5–8) → (optional) score threshold → build prompt → call LLM → return JSON.
* Add **citations** by including chunk metadata in the prompt and response.
* Add **simple UI** (static HTML or tiny React page) that streams tokens.

**Success criteria**

* Ask 5–10 questions about your docs → relevant answers with 2–4 citations.
* P95 latency under \~1.5–2.5s (hosted LLM) on small corpora.

**Stretch**

* Server-sent events (SSE) streaming.
* Basic “answer + source snippets” rendering in the UI.

---

# Week 3 — Quality, Reliability & Cost

**Objectives**

* Make results better and cheaper; make behavior observable.

**Deliverables**

* Reranking + improved chunking.
* Caching + evaluation harness.

**Tasks**

* **Rerank** top 30 results down to 5–8 (Cohere Rerank or similar).
* **Chunking iteration:** overlap 50–100 tokens; consider semantic splitting by headings.
* **Caching:** store query→results and embeddings in Postgres/Redis.
* **Eval harness:** a YAML with {question, expected\_sources} and a small Go test that:

  * measures hit\@k (did we retrieve the right doc?),
  * counts hallucinations (no source match),
  * logs latency + token usage.
* **Observability:** request/response & token counts in logs; Prometheus counters.

**Success criteria**

* +10–20% improvement in hit\@k vs Week 2.
* Stable P95 latency and reduced API spend via caching.

**Stretch**

* Guardrails (max context size, profanity/PII filters).
* Automated nightly eval run with a small dashboard.

---

# Week 4 — Productionization & Deployment

**Objectives**

* Ship a deployable, secure, reproducible service.

**Deliverables**

* Docker images (app + vector DB if local).
* Auth (API key or JWT).
* Cloud deployment (e.g., Fly.io, AWS ECS/Fargate, or GCP Cloud Run).
* Runbook + README.

**Tasks**

* **Dockerize**: minimal image via `FROM golang:1.xx-alpine` builder + scratch/distroless.
* **Config** via env and `config/` with sane defaults.
* **Auth middleware**; rate limiting (token bucket).
* **CI**: lint, vet, unit tests, and a small load test (k6) on PR.
* **Cost & latency tuning**: batch upserts; reduce top\_k; compress prompts; enable response truncation where safe.

**Success criteria**

* One-command deploy.
* Smoke test: upload a doc, ask questions, see logs/metrics.
* Basic SLO: 99% uptime, P95 ≤ 2.5s on demo corpus.

**Stretch**

* Multi-tenant: workspace\_id on all storage rows.
* Background jobs for ingestion via queue (e.g., NATS).

---

## Recommended repo structure

```
rag-go/
  cmd/
    api/               # main.go (HTTP server)
    ingest/            # main.go (CLI for batch ingestion)
  internal/
    config/            # load env, structs
    docs/              # parsing, chunking
    embed/             # embeddings client
    store/             # vector DB + metadata store
    rag/               # retrieve, rerank, prompt, generate
    llm/               # LLM client (OpenAI/Ollama)
    eval/              # evaluation harness
    auth/              # middleware
  web/                 # tiny UI (optional)
  deployments/
    docker/            # Dockerfile(s), compose
    k8s/               # manifests (if used)
  data/                # sample documents
  Makefile
  .env.example
  README.md
```

---

## Concrete milestones (checklist)

* **M1 (end W1):** ingest CLI works; vectors present; metrics printed.
* **M2 (mid W2):** `/query` returns coherent answers; citations included.
* **M3 (end W2):** simple UI + streaming.
* **M4 (mid W3):** reranking + improved chunking; cache enabled.
* **M5 (end W3):** eval harness shows measurable gains.
* **M6 (end W4):** Dockerized, authenticated, deployed.

---

## Minimal API contracts

* `POST /query`

  * **Request:** `{ "query": "How do I reset SSO?", "top_k": 8 }`
  * **Response:** `{ "answer": "...", "citations": [{ "doc_id":"handbook.pdf","page":12,"chunk_id":"12-3"}], "latency_ms": 812 }`
* `POST /upload` (optional)

  * multipart file → returns `{doc_id, chunks, vectors}`

---

## Success metrics to track

* **Retrieval:** hit\@k, MRR.
* **Answering:** citation coverage (% answers with ≥1 good source).
* **Ops:** P50/P95 latency, error rate, \$/100 queries.
* **User:** feedback thumbs up/down (even for yourself).

---

## Suggested “golden path” stack (swappable later)

* **Embeddings:** OpenAI text-embedding-3-small (cheap/strong).
* **LLM:** Hosted GPT or local **Ollama** for dev.
* **Vector store:** **Qdrant** (Docker) or **pgvector** if you prefer Postgres.
* **Frameworks:** net/http + chi (or Fiber), no over-orchestration early.

---

## Summary

The project is an **AI-Powered Knowledge Search Assistant** built in Go, backed by RAG and LLMs. It will allow users to upload documents and query them conversationally, returning accurate answers with citations.

* **Learning Goals:**

  * Strengthen Go backend development skills.
  * Master embeddings, vector databases, and RAG pipelines.
  * Learn how to integrate and deploy LLM-powered services in production.
  * Gain experience with observability, caching, reranking, and scaling AI apps.

* **Outcome:**
  By the end of this project, I will have a **production-ready AI knowledge assistant**, a portfolio-worthy showcase of practical AI engineering skills, and a foundation for future work in multi-agent systems, multi-modal AI, and scalable AI-powered applications.


---

## Executing Process

1. To start executing store the data you want to ingest on the data folder.
2. Execute `go mod tidy`
3. Execute `go run ./cmd/ingest`