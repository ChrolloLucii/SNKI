# Zal MVP — Team Sports Slots

> A service for solo players to quickly find and join a team sports game slot in a gym with transparent rules and predictable costs.

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [MVP Scope](#2-mvp-scope)
3. [Architecture](#3-architecture)
4. [How to Run Locally (Docker Compose)](#4-how-to-run-locally-docker-compose)
5. [How to Run Migrations](#5-how-to-run-migrations)
6. [How to Seed Demo Data](#6-how-to-seed-demo-data)
7. [API Endpoints Summary](#7-api-endpoints-summary)
8. [Contribution Workflow](#8-contribution-workflow)
9. [Issue & Board Workflow](#9-issue--board-workflow)

---

## 1. Project Overview

**Zal** helps solo sports players find open game slots at local gyms, join them, and pay — all from their phone.

Key problems solved:
- Hard to find pick-up games at nearby gyms
- Unclear pricing and cancellation rules
- No-shows and last-minute cancellations ruin games

Users open the PWA on their phone, pick a sport, district, and time — then browse available slots, read the rules, and reserve a spot with one tap.

---

## 2. MVP Scope

### ✅ Must-Have
- Slot catalog with filters (sport, district, date range)
- Slot details page ("slot card") with rules, pricing, cancellation policy
- Join / reserve a spot (concurrency-safe, no overbooking)
- Fake payment flow (sets participation status to PAID)
- My Participations page
- Admin API to create slots + seed data for demo

### 🟡 Nice-to-Have
- Basic in-app notifications
- Admin web page for creating slots

### ❌ Out of Scope for MVP
- Real payment provider (Stripe, etc.)
- Push notifications
- Microservices / Kafka / gRPC
- iOS / Android native apps (PWA only)

---

## 3. Architecture

```
┌───────────────────────────────────────────────────────┐
│                     Browser / Phone                   │
│                                                       │
│   React PWA (Vite + TS)   http://localhost:5173       │
│   ┌───────────────────────────────────────────────┐   │
│   │  /slots (list+filters)                        │   │
│   │  /slots/:id (slot card)                       │   │
│   │  /me/participations                           │   │
│   └───────────────────────────────────────────────┘   │
└────────────────────┬──────────────────────────────────┘
                     │ REST/JSON
┌────────────────────▼──────────────────────────────────┐
│   Go REST API (chi router)   http://localhost:8080    │
│                                                       │
│   /health                                             │
│   /slots            GET  list + filters               │
│   /slots/:id        GET  details                      │
│   /slots/:id/join   POST reserve (concurrency safe)   │
│   /slots/:id/pay    POST fake payment                 │
│   /me/participations GET  my bookings                 │
│   /admin/slots      POST create (admin token)         │
└────────────────────┬──────────────────────────────────┘
                     │ SQL
┌────────────────────▼──────────────────────────────────┐
│   PostgreSQL 15    postgresql://localhost:5432/zal    │
│                                                       │
│   tables: users, slots, participants, payments        │
└───────────────────────────────────────────────────────┘
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for more detail.

---

## 4. How to Run Locally (Docker Compose)

**Prerequisites:** [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or Docker + Docker Compose v2).

```bash
# 1. Clone the repo
git clone https://github.com/<your-org>/zal-mvp.git
cd zal-mvp

# 2. Copy environment variables
cp .env.example .env

# 3. Start everything (DB + API + Web)
docker compose up --build
```

Services will be available at:

| Service | URL |
|---------|-----|
| Web PWA | http://localhost:5173 |
| API     | http://localhost:8080 |
| DB      | postgresql://postgres:postgres@localhost:5432/zal |

To stop: `docker compose down`  
To reset DB: `docker compose down -v`

---

## 5. How to Run Migrations

Migrations run automatically on API startup via `golang-migrate`.

To run them manually:

```bash
# Inside the running api container
docker compose exec api ./migrate -path /app/migrations -database "$DATABASE_URL" up

# Or with migrate CLI installed locally
migrate -path apps/api/migrations -database "postgresql://postgres:postgres@localhost:5432/zal?sslmode=disable" up
```

Migration files live in `apps/api/migrations/`.

---

## 6. How to Seed Demo Data

```bash
# Seed via the admin API (requires ADMIN_TOKEN from .env)
curl -X POST http://localhost:8080/admin/slots \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: secret" \
  -d '{
    "sport": "football",
    "district": "Центральный",
    "venue_name": "Стадион Динамо",
    "address": "ул. Ленина 1",
    "starts_at": "2026-04-01T18:00:00Z",
    "duration_minutes": 60,
    "capacity": 10,
    "min_players": 6,
    "deadline_at": "2026-04-01T16:00:00Z",
    "expected_price": 500,
    "max_price": 700,
    "rules_text": "Casual game, no tackles."
  }'
```

A seed script (for CI / local demo) lives at `apps/api/cmd/seed/main.go`.

---

## 7. API Endpoints Summary

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | — | Health check |
| GET | `/slots` | — | List slots (`?sport=&district=&date_from=&date_to=`) |
| GET | `/slots/{slotId}` | — | Slot details |
| POST | `/slots/{slotId}/join` | User token | Reserve a spot (409 if full) |
| POST | `/slots/{slotId}/pay` | User token | Fake payment → PAID |
| GET | `/me/participations` | User token | My bookings |
| POST | `/admin/slots` | Admin token | Create a slot |

Full OpenAPI spec: [docs/API.md](docs/API.md)

---

## 8. Contribution Workflow

We use **trunk-based development** with short-lived feature branches.

```
main  ←── feature/<ticket>-short-desc  (PR + review)
      ←── fix/<ticket>-short-desc
      ←── tech/<ticket>-short-desc
```

### Step-by-step

1. Pick an issue from the **Ready** column on the project board.
2. Create a branch: `git checkout -b feature/42-slot-filters`
3. Make your changes, commit often with clear messages:  
   `git commit -m "feat(slots): add district filter"`
4. Push and open a Pull Request using the [PR template](.github/pull_request_template.md).
5. Request a review from at least one teammate.
6. Merge to `main` once approved (squash merge preferred).

### Commit Message Convention

```
<type>(<scope>): <short description>

Types: feat, fix, refactor, test, docs, chore
Scopes: slots, participation, payments, admin, web, db, infra
```

---

## 9. Issue & Board Workflow

We use GitHub Issues + a GitHub Project board with these columns:

| Column | Meaning |
|--------|---------|
| **Backlog** | Ideas and future tasks — not yet prioritized |
| **Ready** | Clearly scoped, ready to pick up |
| **In Progress** | Someone is actively working on it |
| **In Review** | PR is open, waiting for review |
| **Done** | Merged to main |

### For non-technical teammates

- **To report a bug:** open an issue → choose _Bug Report_ template.
- **To request a feature:** open an issue → choose _Feature Request_ template.
- **To track a tech task:** use _Tech Task_ template.
- Assign labels (see below) and move the card to the right column.

### Labels

| Label | Meaning |
|-------|---------|
| `type:feature` | New feature or user story |
| `type:bug` | Something broken |
| `type:tech` | Refactor, infra, tooling |
| `type:docs` | Documentation |
| `priority:must` | Blocker for MVP |
| `priority:should` | Important, but not blocking |
| `area:web` | Frontend |
| `area:api` | Backend |
| `area:db` | Database / migrations |
| `good first issue` | Good for newcomers |
