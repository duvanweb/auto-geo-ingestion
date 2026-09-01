# Architecture — Auto-Geo Fleet Tracking System

## Overview

Sistema de tracking vehicular en tiempo real compuesto por dos microservicios Go y un dashboard React. Detecta alertas de vehículos detenidos, previene duplicados con Redis y mantiene resiliencia mediante circuit breakers.

```
                    ┌─────────────────────────────────────┐
                    │         auto-geo-front (React)        │
                    │  Dashboard: vehicle list + Leaflet map │
                    │  TanStack Query polling cada 5s       │
                    └──────────┬──────────────────────────┘
                               │ HTTP
              ┌────────────────┴──────────────────────┐
              │                                       │
              ▼                                       ▼
  ┌───────────────────────┐         ┌────────────────────────────┐
  │  auto-geo-ingestion   │         │     auto-geo-alerts        │
  │  :8080                │         │     :8081                  │
  │                       │         │                            │
  │ CRUD /v1/vehicles     │         │ GET  /v1/vehicles          │
  │ POST /v1/locations    │──CB──►  │ GET  /v1/alerts            │
  │                       │  HTTP   │ POST /v1/detections        │
  └────────┬──────────────┘         └──────────┬─────────────────┘
           │ CB1                               │
           ▼                                   ▼
    ┌──────────┐                       ┌──────────────┐
    │  Redis   │                       │  PostgreSQL  │
    │  :6379   │                       │  :5432       │
    └──────────┘                       └──────────────┘
```

---

## Services

### auto-geo-ingestion (:8080)

**Responsabilidades**: CRUD de vehículos + ingesta de coordenadas GPS.

| Endpoint | Descripción |
|---|---|
| `POST /api/v1/vehicles` | Crear vehículo |
| `GET /api/v1/vehicles` | Listar vehículos activos |
| `GET /api/v1/vehicles/{id}` | Obtener vehículo por ID |
| `PUT /api/v1/vehicles/{id}` | Actualizar vehículo |
| `DELETE /api/v1/vehicles/{id}` | Eliminar vehículo (soft delete + saga) |
| `GET /api/health` | Health check |
| `POST /api/v1/locations` | Ingestar coordenada GPS |

**Respuestas POST /v1/locations**:
- `201 Created` — guardado en DB y notificado a alerts
- `409 Conflict` — coordenada duplicada dentro de la ventana de deduplicación (30s)
- `202 Accepted` — DB circuit breaker abierto, encolado en memoria

**Dependencias**: PostgreSQL (persistencia), Redis (anti-duplicado), auto-geo-alerts (notificación asíncrona).

---

### auto-geo-alerts (:8081)

**Responsabilidades**: Detectar vehículos detenidos y exponer estado de la flota.

| Endpoint | Descripción |
|---|---|
| `POST /api/v1/detections` | Recibir coordenada de ingestion |
| `GET /api/v1/vehicles` | Estado actual de todos los vehículos |
| `GET /api/v1/alerts` | Alertas activas (no resueltas) |
| `GET /api/health` | Health check |

**Lógica de detección**:
1. Si mismas coordenadas (ε = 0.00001°) y tiempo > 60s → alerta `vehicle_stopped`
2. Si coordenadas cambiaron → resolver alerta activa, estado → `moving`

**Dependencias**: PostgreSQL únicamente.

---

### auto-geo-front (:3000)

Dashboard React SPA con polling cada 5s.

| Componente | Descripción |
|---|---|
| `MapView` | Mapa Leaflet centrado en Colombia, marcadores por vehículo coloreados por status |
| `VehicleList` | Tabla con ID, status badge, lat/lng, última actualización |
| `AlertBanner` | Banner superior con conteo de alertas activas |
| `StatusBadge` | moving=verde, stopped=amarillo, alert=rojo |

---

## Architecture Pattern

### Hexagonal Architecture (Ports & Adapters)

```
cmd/api/
  main.go          ← bootstrap FX, graceful shutdown
  module.go        ← assembles all FX modules

internal/
  core/
    domain/        ← pure entities (Vehicle, Location, Alert)
    ports/
      repositories/  ← repository interfaces (input to infra)
      services/      ← service interfaces (input to core)
      resources/     ← external resource interfaces (Cache)
    vehicle/       ← VehicleService implementation
    location/      ← LocationIngestionService implementation
    alert/         ← AlertService implementation

  infrastructure/
    postgres/      ← DB connection pool, repository implementations
    redis/         ← Redis client, LocationCache implementation
    http/          ← outbound HTTP client (alerts), circuit breaker
    api/
      controllers/ ← HTTP handlers
      dtos/        ← request/response shapes
      router/      ← Chi router, middleware, lifecycle hooks
```

### Dependency Injection — Uber FX

Cada capa publica un `fx.Options` (module.go). El `cmd/api/module.go` los compone:

```go
// Patrón para registrar implementación como interface
fx.Annotate(NewVehicleService, fx.As(new(ports.VehicleService)))

// Desambiguación de tipos cuando hay wrapper (Circuit Breaker)
// NewLocationRepository retorna *LocationRepository (concreto)
// El wrapper CB se registra como ports.LocationRepository (interface)
```

---

## Resilience — Circuit Breaker

Implementado con `github.com/sony/gobreaker`. El servicio de ingesta **nunca debe caer** aunque la DB o el servicio de alertas fallen.

### CB1 — PostgreSQL (ingestion)

```
LocationIngestionService
    │
    ▼
LocationRepositoryWithCB  ←── gobreaker.CircuitBreaker
    │                              │
    ├── closed → inner.Save()      │ MaxRequests: 3
    │                              │ Timeout:     60s
    └── open → chan(1000) ◄────────┘ ConsecutiveFailures > 3

drainQueue goroutine (cada 5s)
    └── cuando circuit se cierra → vacía la cola → inner.Save()
```

**Comportamiento**:
- DB disponible → guarda directamente
- DB falla → encola en buffer de 1000 items (bounded channel)
- Buffer lleno → descarta con log (nunca propaga error al caller)
- DB recupera → goroutine drain vacía la cola

### CB2 — Alerts service (HTTP)

```
LocationController.Ingest()
    └── go alertsClient.NotifyLocation()  ← goroutine asíncrona
            │
            ▼
        AlertsClient  ←── gobreaker.CircuitBreaker
            │                   MaxRequests: 3
            ├── closed → POST /v1/detections   Timeout: 10s
            └── open   → log + return nil
```

**Comportamiento**: Si alerts falla, la ingesta continúa sin interrupción. El dashboard verá las ubicaciones aunque la detección de alertas esté temporalmente fuera.

---

## Anti-Duplicate Location Deduplication

```
POST /v1/locations
    │
    ▼
Redis GET location:vehicle:{id}:last
    │
    ├── not found OR coords differ (ε=0.00001°) → save to DB + update Redis (TTL 30s)
    │
    └── same coords within TTL → return ErrDuplicate (HTTP 409)
```

La clave Redis expira a los 30 segundos. Si el mismo GPS envía la misma posición más de una vez en ese ventana, la segunda se descarta silenciosamente.

---

## Vehicle Deletion — Saga Pattern

```
DELETE /api/v1/vehicles/{id}
    │
    ▼
1. Postgres: UPDATE vehicles SET deleted_at = NOW() WHERE id = ?
    │         (soft delete, transacción Postgres)
    │
    ▼
2. go redis.DeleteVehicleKeys(ctx, vehicleID)
    │   ← goroutine asíncrona (respuesta HTTP es inmediata)
    │
    ├── éxito → fin
    └── fallo → log de error (best-effort)
```

**Consistencia eventual**: El vehículo desaparece del listado en el próximo poll del frontend (máx 5s). Las claves Redis son efímeras con TTL de 30s, por lo que expiran naturalmente aunque DeleteVehicleKeys falle.

---

## Database Schemas

### auto-geo-ingestion

```sql
-- vehicles (soft delete)
CREATE TABLE vehicles (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    plate       VARCHAR(20) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ  -- NULL = activo
);

-- vehicle_locations
CREATE TABLE vehicle_locations (
    id         BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id),
    lat        DOUBLE PRECISION NOT NULL,
    lng        DOUBLE PRECISION NOT NULL,
    timestamp  TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vl_vehicle_id_ts ON vehicle_locations(vehicle_id, timestamp DESC);
```

### auto-geo-alerts

```sql
-- estado actual por vehículo (UPSERT)
CREATE TABLE vehicle_status (
    vehicle_id BIGINT PRIMARY KEY,
    status     VARCHAR(20) NOT NULL DEFAULT 'moving',
    last_lat   DOUBLE PRECISION,
    last_lng   DOUBLE PRECISION,
    last_seen  TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- alertas (resolved_at NULL = activa)
CREATE TABLE alerts (
    id          BIGSERIAL PRIMARY KEY,
    vehicle_id  BIGINT NOT NULL,
    alert_type  VARCHAR(50) NOT NULL,
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    started_at  TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- índice parcial para consultas de alertas activas
CREATE INDEX idx_alerts_active ON alerts(vehicle_id) WHERE resolved_at IS NULL;
```

---

## Environment Variables

### auto-geo-ingestion

| Variable | Default | Descripción |
|---|---|---|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASS` | `secret` | PostgreSQL password |
| `DB_NAME` | `auto_geo_ingestion` | Database name |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | `` | Redis password (vacío = sin auth) |
| `REDIS_DB` | `0` | Redis database index |
| `ALERTS_SERVICE_URL` | `http://localhost:8081` | URL del servicio de alertas |
| `SERVER_PORT` | `8080` | Puerto HTTP del servicio |

### auto-geo-alerts

| Variable | Default | Descripción |
|---|---|---|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASS` | `secret` | PostgreSQL password |
| `DB_NAME` | `auto_geo_alerts` | Database name |
| `SERVER_PORT` | `8081` | Puerto HTTP del servicio |

### auto-geo-front

| Variable | Descripción |
|---|---|
| `VITE_INGESTION_API_URL` | URL base del servicio de ingesta |
| `VITE_ALERTS_API_URL` | URL base del servicio de alertas |
| `VITE_POLL_INTERVAL_MS` | Intervalo de polling en ms (default 5000) |

---

## Technical Decisions

| Decisión | Elección | Razón |
|---|---|---|
| Comunicación inter-service | HTTP REST | Simple, testeable, sin infraestructura adicional (no MQ) |
| Circuit breaker | `sony/gobreaker` | Pure Go, sin dependencias externas, API clara |
| Anti-duplicate | Redis con TTL | O(1) lookup, expiración automática sin cron jobs |
| Soft delete | `deleted_at` TIMESTAMPTZ | Preserva historial de ubicaciones con FK intacta |
| DI framework | Uber FX | Lifecycle hooks, módulos composables, evita global state |
| Router | Chi | Lightweight, compatible con net/http estándar |
| JSON | `json-iterator` | Drop-in de encoding/json, 3-6x más rápido |
| Env loading | `caarlos0/env/v10` | Genérico, sin reflejo manual, struct tags |
| Frontend polling | TanStack Query | Stale-while-revalidate gratis, cache automático |
| Map | react-leaflet@4 | Gratuito, sin API key, compatible con React 18 |

---

## Local Development

### Prerrequisitos
- Go 1.22+
- Node 20+
- Docker + Docker Compose

### Con Docker Compose (recomendado)

```bash
cd G:\Cursos\auto-geo-ingestion

# Levantar todos los servicios
docker compose up -d

# Ver logs
docker compose logs -f ingestion
docker compose logs -f alerts

# Detener
docker compose down
```

### Sin Docker (desarrollo)

```bash
# 1. Infraestructura
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=secret postgres:16-alpine
docker run -d -p 6379:6379 redis:7-alpine

# 2. Migraciones
cd G:\Cursos\auto-geo-ingestion
migrate -path db/migration -database "postgres://postgres:secret@localhost:5432/auto_geo_ingestion?sslmode=disable" up

cd G:\Cursos\auto-geo-alerts
migrate -path db/migration -database "postgres://postgres:secret@localhost:5432/auto_geo_alerts?sslmode=disable" up

# 3. Servicios Go
cd G:\Cursos\auto-geo-ingestion && go run ./cmd/api/
cd G:\Cursos\auto-geo-alerts    && go run ./cmd/api/

# 4. Frontend
cd G:\Cursos\auto-geo-front && npm run dev
```

### Verificación end-to-end

```bash
# Crear vehículo
curl -X POST http://localhost:8080/api/v1/vehicles \
  -H "Content-Type: application/json" \
  -d '{"name":"Truck-01","plate":"ABC123","description":"Camión de prueba"}'

# Ingestar coordenada
curl -X POST http://localhost:8080/api/v1/locations \
  -H "Content-Type: application/json" \
  -d '{"vehicle_id":1,"lat":4.711,"lng":-74.072,"timestamp":"2026-09-01T00:00:00Z"}'

# Misma coordenada → 409 Conflict (anti-duplicate)
curl -X POST http://localhost:8080/api/v1/locations \
  -H "Content-Type: application/json" \
  -d '{"vehicle_id":1,"lat":4.711,"lng":-74.072,"timestamp":"2026-09-01T00:00:01Z"}'

# Ver estado de vehículos
curl http://localhost:8081/api/v1/vehicles

# Ver alertas activas
curl http://localhost:8081/api/v1/alerts
```

---

## GitHub Repositories

| Repo | Descripción |
|---|---|
| [duvanweb/auto-geo-ingestion](https://github.com/duvanweb/auto-geo-ingestion) | CRUD vehículos + ingesta GPS + docker-compose + CI/CD |
| [duvanweb/auto-geo-alerts](https://github.com/duvanweb/auto-geo-alerts) | Detección de alertas vehiculares |
| [duvanweb/auto-geo-front](https://github.com/duvanweb/auto-geo-front) | Dashboard React SPA |
