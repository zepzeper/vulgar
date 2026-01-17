# Vulgar Module Implementation Plan

## Overview

| Category | Total | Implemented | Stubs |
|----------|-------|-------------|-------|
| Core | 9 | **9** | 0 |
| Stdlib | 28 | **20** | 8 |
| Integrations | 26+ | **~8** | 18 |
| AI | 5 | **0** | 5 |

---

## ✅ Implemented Modules (No Action Needed)

### Core (All Implemented)
`http`, `json`, `log`, `fs`, `path`, `time`, `env`, `crypto`, `uuid`

### Stdlib (20 Implemented)
`timer`, `cron`, `event`, `shell`, `process`, `filewatch`, `yaml`, `xml`, `csv`, `regex`, `strings`, `url`, `html`, `template`, `validator`, `mathx`, `queue`, `workflow`, `retry`, `compress/*`

### Integrations (Implemented)
`gsheets`, `gdrive`, `gcalendar`, `github`, `gitlab`, `codeberg`, `slack`, `ssh`, `smtp`

---

## 🔴 Stub Modules (Need Implementation)

### Priority 1: High-Value, Common Use Cases

| Module | Category | Effort | Why Priority |
|--------|----------|--------|--------------|
| `ai.openai` | AI | Medium | ChatGPT is extremely popular for automation |
| `ai.anthropic` | AI | Medium | Claude API, growing popularity |
| `ai.ollama` | AI | Medium | Local LLM, privacy-focused users |
| `integrations.postgres` | Database | Medium | Most common production DB |
| `integrations.sqlite` | Database | Low | Local/embedded DB, great for testing |
| `integrations.redis` | Cache | Medium | Ubiquitous caching/queuing |

### Priority 2: Cloud & DevOps

| Module | Category | Effort | Why Priority |
|--------|----------|--------|--------------|
| `integrations.s3` | Storage | Medium | AWS S3 / MinIO, very common |
| `integrations.docker` | DevOps | Medium | Container automation |
| `integrations.k8s` | DevOps | High | Kubernetes orchestration |
| `integrations.webhook` | Web | Low | Inbound HTTP triggers |
| `integrations.websocket` | Web | Medium | Real-time communication |

### Priority 3: Messaging & Notifications

| Module | Category | Effort | Why Priority |
|--------|----------|--------|--------------|
| `integrations.notify.discord` | Messaging | Low | Already have slack pattern |
| `integrations.notify.telegram` | Messaging | Low | Already have slack pattern |
| `integrations.kafka` | Messaging | High | Enterprise event streaming |
| `integrations.rabbitmq` | Messaging | Medium | Message queuing |
| `integrations.nats` | Messaging | Medium | Cloud-native messaging |

### Priority 4: Additional Databases & Services

| Module | Category | Effort | Why Priority |
|--------|----------|--------|--------------|
| `integrations.mongodb` | Database | Medium | Document DB |
| `integrations.graphql` | API | Medium | GraphQL client |
| `integrations.stripe` | Payments | Medium | Payment processing |
| `integrations.twilio` | SMS | Low | SMS/Voice |
| `integrations.airtable` | Productivity | Low | Spreadsheet-like DB |
| `integrations.notion` | Productivity | Medium | Knowledge base |

### Priority 5: Utilities & Specialized

| Module | Category | Effort | Why Priority |
|--------|----------|--------|--------------|
| `stdlib.jwt` | Auth | Low | Token handling |
| `stdlib.cache` | Perf | Low | In-memory caching |
| `stdlib.secrets` | Security | Low | Secret management |
| `stdlib.parallel` | Perf | Medium | Concurrent execution |
| `stdlib.metrics` | Observability | Low | Prometheus metrics |
| `stdlib.health` | Observability | Low | Health checks |
| `stdlib.trace` | Observability | Medium | Distributed tracing |
| `stdlib.osinfo` | System | Low | OS information |
| `integrations.dns` | Network | Low | DNS lookups |
| `integrations.ftp` | Network | Low | FTP transfers |
| `ai.huggingface` | AI | Medium | ML models |
| `ai.localai` | AI | Medium | Local AI server |

---

## Recommended Implementation Order

```
Phase 1: Core Workflows          Phase 2: Cloud/DevOps
─────────────────────────        ─────────────────────────
1. sqlite (quick win)            7. s3 (storage)
2. postgres (production DB)      8. docker (containers)
3. redis (caching)               9. webhook (triggers)
4. openai (AI automation)        10. websocket (real-time)
5. anthropic (Claude)
6. ollama (local LLM)

Phase 3: Messaging               Phase 4: Everything Else
─────────────────────────        ─────────────────────────
11. discord (notifications)      15. mongodb
12. telegram (notifications)     16. stripe
13. jwt (auth)                   17. graphql
14. cache (performance)          18. remaining modules...
```

---

## Effort Estimates

| Effort | Lines of Code | Time |
|--------|--------------|------|
| Low | 100-200 | 1-2 hours |
| Medium | 200-400 | 2-4 hours |
| High | 400+ | 4-8 hours |
