# Mermaid диаграммы

## Общая архитектура

```mermaid
flowchart TD
    subgraph Client["Браузер / Телефон"]
        PWA["React PWA (Vite + TS)\nhttp://localhost:5173\n\n/slots\n/slots/:id\n/me/participations"]
    end

    subgraph API["Go REST API (chi router)\nhttp://localhost:8080"]
        E1["GET /health"]
        E2["GET /slots"]
        E3["GET /slots/{id}"]
        E4["POST /slots/{id}/join"]
        E5["POST /slots/{id}/pay"]
        E6["GET /me/participations"]
        E7["POST /admin/slots"]
    end

    subgraph DB["PostgreSQL 15\npostgresql://localhost:5432/zal"]
        T1["users"]
        T2["slots"]
        T3["participants"]
        T4["payments"]
    end

    PWA -->|"REST / JSON"| API
    API -->|"SQL"| DB
```

## Поток данных: присоединение к слоту

```mermaid
sequenceDiagram
    participant B as Браузер
    participant A as API handler (participation.Join)
    participant D as База данных

    B->>A: POST /slots/{id}/join<br/>X-User-Token: &lt;token&gt;
    A->>D: BEGIN TRANSACTION
    A->>D: SELECT COUNT(*) FROM participants<br/>WHERE slot_id = $1 AND status IN ('RESERVED','PAID')<br/>FOR UPDATE
    D-->>A: count
    alt count >= slot.capacity
        A-->>B: 409 Conflict
    else Есть свободные места
        A->>D: INSERT INTO participants (slot_id, user_id, status='RESERVED')
        A->>D: COMMIT
        A-->>B: 201 Created { participant_id, status: "RESERVED" }
    end
```
