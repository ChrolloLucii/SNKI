# API Reference

Base URL (local): `http://localhost:8080`

---

## Authentication

MVP uses a simple header-based demo auth:

| Header | Used by | Value |
|--------|---------|-------|
| `X-User-Token` | User endpoints | Any token that maps to a user row |
| `X-Admin-Token` | Admin endpoints | Value of `ADMIN_TOKEN` env var |

---

## Endpoints

### GET `/health`

Health check. No auth required.

**Response `200`**
```json
{ "status": "ok" }
```

---

### GET `/slots`

List available game slots.

**Query Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `sport` | string | e.g. `football`, `basketball` |
| `district` | string | City district |
| `date_from` | RFC3339 | Start of date range |
| `date_to` | RFC3339 | End of date range |

**Response `200`**
```json
[
  {
    "id": "uuid",
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
    "rules_text": "Casual game, no tackles.",
    "status": "OPEN",
    "spots_left": 4
  }
]
```

---

### GET `/slots/{slotId}`

Get slot details.

**Response `200`** — same shape as list item above.  
**Response `404`** — slot not found.

---

### POST `/slots/{slotId}/join`

Reserve a spot. Requires `X-User-Token` header.

Concurrency-safe: uses a DB transaction with `SELECT … FOR UPDATE`.  
Returns `409` if the slot is full or user already joined.

**Response `201`**
```json
{
  "participant_id": "uuid",
  "slot_id": "uuid",
  "user_id": "uuid",
  "status": "RESERVED",
  "reserved_at": "2026-03-21T10:00:00Z"
}
```

**Error responses**

| Status | Reason |
|--------|--------|
| `401` | Missing `X-User-Token` |
| `404` | Slot not found |
| `409` | Slot full or already joined |

---

### POST `/slots/{slotId}/pay`

Fake payment — sets participant status to `PAID` and creates a payment record.  
Requires `X-User-Token` header.

**Response `200`**
```json
{
  "payment_id": "uuid",
  "participant_id": "uuid",
  "status": "PAID",
  "amount": 500,
  "currency": "RUB",
  "provider": "FAKE"
}
```

**Error responses**

| Status | Reason |
|--------|--------|
| `401` | Missing `X-User-Token` |
| `404` | Participation not found |
| `409` | Already paid |

---

### GET `/me/participations`

List current user's participations. Requires `X-User-Token` header.

**Response `200`**
```json
[
  {
    "participant_id": "uuid",
    "slot": { /* same shape as GET /slots/{id} */ },
    "status": "PAID",
    "reserved_at": "2026-03-21T10:00:00Z",
    "paid_at": "2026-03-21T10:05:00Z"
  }
]
```

---

### POST `/admin/slots`

Create a new slot. Requires `X-Admin-Token` header.

**Request Body**
```json
{
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
}
```

**Response `201`** — created slot object.

**Error responses**

| Status | Reason |
|--------|--------|
| `401` | Missing or invalid `X-Admin-Token` |
| `422` | Validation error |

---

## Error Shape

All error responses use the same JSON envelope:

```json
{
  "error": "human readable message"
}
```
