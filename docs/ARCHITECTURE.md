# Architecture

## Overview

Zal MVP is a monorepo with three runnable services wired together via Docker Compose:

```
┌────────────────────────────────────────────────────────┐
│                    Browser / Phone                     │
│                                                        │
│   React PWA (Vite + TS)    http://localhost:5173       │
│   ┌────────────────────────────────────────────────┐   │
│   │  /slots          — browse & filter game slots  │   │
│   │  /slots/:id      — slot details card           │   │
│   │  /me/participations — my bookings              │   │
│   └────────────────────────────────────────────────┘   │
└─────────────────────┬──────────────────────────────────┘
                      │ REST / JSON
┌─────────────────────▼──────────────────────────────────┐
│   Go REST API (chi router)   http://localhost:8080     │
│                                                        │
│   GET  /health                                         │
│   GET  /slots                 list + filters           │
│   GET  /slots/{id}            slot details             │
│   POST /slots/{id}/join       reserve (concurrency ✓)  │
│   POST /slots/{id}/pay        fake payment             │
│   GET  /me/participations     my bookings              │
│   POST /admin/slots           create slot (admin)      │
└─────────────────────┬──────────────────────────────────┘
                      │ SQL (pgx/v5)
┌─────────────────────▼──────────────────────────────────┐
│   PostgreSQL 15    localhost:5432 / db:5432 (Docker)   │
│                                                        │
│   users        — demo identity (token-based)           │
│   slots        — game slot catalog                     │
│   participants — reservations (RESERVED → PAID)        │
│   payments     — fake payment records                  │
└────────────────────────────────────────────────────────┘
```

## Directory Layout

```
zal-mvp/
├── apps/
│   ├── api/          # Go REST API
│   │   ├── cmd/api/  # Entrypoint (main.go)
│   │   ├── internal/ # Domain modules: slots, participation, payments, admin
│   │   └── migrations/
│   └── web/          # React + Vite PWA
│       └── src/
│           ├── routes/     # SlotsList, SlotDetails, MyParticipations
│           └── components/
├── infra/            # Shared infra config (reserved)
├── docs/             # ARCHITECTURE.md, API.md
├── .github/
│   ├── ISSUE_TEMPLATE/
│   └── pull_request_template.md
├── docker-compose.yml
└── .env.example
```

## Key Design Decisions

| Concern | Decision | Reason |
|---------|----------|--------|
| Auth | Demo token (no password) | Fastest MVP; avoids auth complexity |
| Overbooking | DB transaction + `SELECT … FOR UPDATE` + capacity check | Correct under concurrent load |
| Payments | Fake (status flip only) | Real payment out of scope for MVP |
| API style | REST + JSON | Simple, familiar to all |
| Migrations | golang-migrate, run on startup | Zero extra tooling needed in dev |
| Frontend routing | react-router v6 | Standard, well-documented |

## Data Flow: Join a Slot

```
Browser
  │  POST /slots/{id}/join
  │  Headers: X-User-Token: <token>
  ▼
API handler (participation.Join)
  │  BEGIN TRANSACTION
  │  SELECT COUNT(*) FROM participants
  │     WHERE slot_id = $1 AND status IN ('RESERVED','PAID')
  │     FOR UPDATE   ← prevents race condition
  │  if count >= slot.capacity → 409 Conflict
  │  INSERT INTO participants (slot_id, user_id, status='RESERVED')
  │  COMMIT
  ▼
Response 201 Created { participant_id, status: "RESERVED" }
```
