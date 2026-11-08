# Sample CRUD API Documentation

This document describes the RESTful HTTP API and gRPC interface for the **Sample** domain in the `tc-template` service.

---

## 1. Overview & Architecture

- **Service Name:** `tc-template`
- **Default HTTP Port:** `8081` (configurable via `configs/config.yaml`)
- **Default gRPC Port:** `7000` (configurable via `configs/config.yaml`)
- **Base HTTP Path:** `/api/v1`
- **Content-Type:** `application/json`

```
Client
  │
  ├── [HTTP / REST] :8081/api/v1/samples  ──► chi router ──► HTTP Handler ──┐
  │                                                                         ├──► Sample Usecase ──► Postgres Repository (DB)
  └── [gRPC]        :7000/sample.v1.*     ──► gRPC Server ─► gRPC Handler ──┘
```

---

## 2. Authentication

The API uses API Key authentication via the `X-API-Key` request header on protected routes.

| Header | Type | Required | Description |
|---|---|---|---|
| `X-API-Key` | String | Conditional | API key configured in `app.secret` |

### Authentication Rules
- **Development / Bypass Mode:** When `app.secret` is empty, `"SECRET_KEY"`, or `"CHANGE_ME"`, authentication is bypassed and all requests are accepted without `X-API-Key`.
- **Production Mode:** When `app.secret` contains an actual secret, missing or invalid `X-API-Key` headers return HTTP `401 Unauthorized`.

---

## 3. Standard Response Format

All HTTP endpoints use a standardized envelope response format.

### 3.1 Success Envelope (Single Entity / Action)
```json
{
  "success": true,
  "message": "sample created",
  "data": { ... }
}
```

### 3.2 Success Envelope (Paginated List)
```json
{
  "success": true,
  "message": "ok",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 42,
    "total_pages": 5
  }
}
```

### 3.3 Error Envelope
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "name": "failed on 'min' validation"
    }
  }
}
```

### 3.4 Error Codes Reference

| Error Code | HTTP Status | Description |
|---|---|---|
| `BAD_REQUEST` | 400 | Malformed JSON body or invalid syntax |
| `UNAUTHORIZED` | 401 | Missing or invalid `X-API-Key` header |
| `FORBIDDEN` | 403 | Insufficient permissions for operation |
| `NOT_FOUND` | 404 | Requested sample resource does not exist |
| `CONFLICT` | 409 | Resource conflict or optimistic locking version mismatch |
| `VALIDATION_ERROR` | 422 | Field validation failure |
| `INTERNAL_ERROR` | 500 | Unexpected internal server error |
| `SERVICE_UNAVAILABLE` | 503 | Database or downstream dependency is unavailable |
| `TIMEOUT` | 504 | Request deadline exceeded |

---

## 4. Data Model: Sample

| Field | Type | Storage / DB | Description | Constraints |
|---|---|---|---|---|
| `id` | String | UUID | Unique identifier | Auto-generated UUID v4 |
| `name` | String | VARCHAR(100) | Name of the sample | Required, 2-100 characters |
| `description` | String | TEXT | Detailed description | Optional, max 500 characters |
| `status` | String | VARCHAR(20) | Operational status | Allowed: `active`, `inactive` (default: `active`) |
| `created_at` | String (RFC 3339) | TIMESTAMPTZ | Creation timestamp | Set automatically on insert |
| `updated_at` | String (RFC 3339) | TIMESTAMPTZ | Last update timestamp | Set automatically on insert/update |

---

## 5. Endpoints Specification

### 5.1 Create Sample

Creates a new sample entity.

- **Method:** `POST`
- **Path:** `/api/v1/samples`
- **Auth:** `X-API-Key` (if enabled)
- **Headers:** `Content-Type: application/json`

#### Request Body Schema
| Field | Type | Required | Default | Validation Rules | Description |
|---|---|---|---|---|---|
| `name` | string | Yes | — | `min=2, max=100` | Sample name |
| `description` | string | No | `""` | `max=500` | Optional description |
| `status` | string | No | `"active"` | `oneof=active inactive` | Status of the sample |

#### Request Example
```bash
curl -X POST http://localhost:8081/api/v1/samples \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "name": "Rock Festival 2026 Ticket",
    "description": "VIP backstage access pass",
    "status": "active"
  }'
```

#### Responses

**`201 Created` - Success**
```json
{
  "success": true,
  "message": "sample created",
  "data": {
    "id": "7f000001-9310-1df0-8193-10bb4e9e0000",
    "name": "Rock Festival 2026 Ticket",
    "description": "VIP backstage access pass",
    "status": "active",
    "created_at": "2026-11-08T09:30:00Z",
    "updated_at": "2026-11-08T09:30:00Z"
  }
}
```

**`400 Bad Request` - Malformed JSON**
```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "invalid request body"
  }
}
```

**`422 Unprocessable Entity` - Validation Failure**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "name": "failed on 'min' validation",
      "status": "failed on 'oneof' validation"
    }
  }
}
```

---

### 5.2 List Samples

Retrieves a paginated list of samples ordered by creation time descending (`created_at DESC`).

- **Method:** `GET`
- **Path:** `/api/v1/samples`
- **Auth:** `X-API-Key` (if enabled)

#### Query Parameters
| Parameter | Type | Required | Default | Bounds | Description |
|---|---|---|---|---|---|
| `page` | integer | No | `1` | `min: 1` | Page number |
| `per_page` | integer | No | `10` | `min: 1, max: 100` | Items per page (values > 100 clamp to 100) |

#### Request Example
```bash
curl -X GET "http://localhost:8081/api/v1/samples?page=1&per_page=10" \
  -H "X-API-Key: YOUR_API_KEY"
```

#### Responses

**`200 OK` - Success**
```json
{
  "success": true,
  "message": "ok",
  "data": [
    {
      "id": "7f000001-9310-1df0-8193-10bb4e9e0000",
      "name": "Rock Festival 2026 Ticket",
      "description": "VIP backstage access pass",
      "status": "active",
      "created_at": "2026-11-08T09:30:00Z",
      "updated_at": "2026-11-08T09:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

---

### 5.3 Get Sample by ID

Retrieves a single sample entity by its unique ID.

- **Method:** `GET`
- **Path:** `/api/v1/samples/{id}`
- **Auth:** `X-API-Key` (if enabled)

#### Path Parameters
| Parameter | Type | Required | Description |
|---|---|---|---|
| `id` | string (UUID) | Yes | Unique sample identifier |

#### Request Example
```bash
curl -X GET http://localhost:8081/api/v1/samples/7f000001-9310-1df0-8193-10bb4e9e0000 \
  -H "X-API-Key: YOUR_API_KEY"
```

#### Responses

**`200 OK` - Success**
```json
{
  "success": true,
  "message": "ok",
  "data": {
    "id": "7f000001-9310-1df0-8193-10bb4e9e0000",
    "name": "Rock Festival 2026 Ticket",
    "description": "VIP backstage access pass",
    "status": "active",
    "created_at": "2026-11-08T09:30:00Z",
    "updated_at": "2026-11-08T09:30:00Z"
  }
}
```

**`404 Not Found` - Resource Not Found**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "resource not found"
  }
}
```

---

### 5.4 Update Sample

Updates an existing sample entity. This replaces `name`, `description`, and `status`.

- **Method:** `PUT`
- **Path:** `/api/v1/samples/{id}`
- **Auth:** `X-API-Key` (if enabled)
- **Headers:** `Content-Type: application/json`

#### Path Parameters
| Parameter | Type | Required | Description |
|---|---|---|---|
| `id` | string (UUID) | Yes | Unique sample identifier |

#### Request Body Schema
| Field | Type | Required | Default | Validation Rules | Description |
|---|---|---|---|---|---|
| `name` | string | Yes | — | `min=2, max=100` | Updated sample name |
| `description` | string | No | `""` | `max=500` | Updated description |
| `status` | string | Yes | — | `oneof=active inactive` | Updated status (`active` or `inactive`) |

#### Request Example
```bash
curl -X PUT http://localhost:8081/api/v1/samples/7f000001-9310-1df0-8193-10bb4e9e0000 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "name": "Rock Festival 2026 Ticket - Sold Out",
    "description": "VIP backstage access pass - allocation exhausted",
    "status": "inactive"
  }'
```

#### Responses

**`200 OK` - Success**
```json
{
  "success": true,
  "message": "sample updated",
  "data": {
    "id": "7f000001-9310-1df0-8193-10bb4e9e0000",
    "name": "Rock Festival 2026 Ticket - Sold Out",
    "description": "VIP backstage access pass - allocation exhausted",
    "status": "inactive",
    "created_at": "2026-11-08T09:30:00Z",
    "updated_at": "2026-11-08T09:45:00Z"
  }
}
```

**`400 Bad Request` - Malformed JSON**
```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "invalid request body"
  }
}
```

**`404 Not Found` - Resource Not Found**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "resource not found"
  }
}
```

**`409 Conflict` - Optimistic Concurrency Conflict**
```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "resource already exists"
  }
}
```

**`422 Unprocessable Entity` - Validation Failure**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "status": "failed on 'required' validation"
    }
  }
}
```

---

### 5.5 Delete Sample

Deletes a sample entity by its ID.

- **Method:** `DELETE`
- **Path:** `/api/v1/samples/{id}`
- **Auth:** `X-API-Key` (if enabled)

#### Path Parameters
| Parameter | Type | Required | Description |
|---|---|---|---|
| `id` | string (UUID) | Yes | Unique sample identifier |

#### Request Example
```bash
curl -X DELETE http://localhost:8081/api/v1/samples/7f000001-9310-1df0-8193-10bb4e9e0000 \
  -H "X-API-Key: YOUR_API_KEY"
```

#### Responses

**`200 OK` - Success**
```json
{
  "success": true,
  "message": "sample deleted",
  "data": null
}
```

**`404 Not Found` - Resource Not Found**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "resource not found"
  }
}
```

---

## 6. gRPC Interface Reference

For high-throughput or internal service-to-service communication, the Sample CRUD is also exposed via gRPC on port `7000`.

- **Package:** `sample.v1`
- **Service:** `SampleService`
- **Proto definition:** `internal/proto/sample/v1/sample.proto`

### Available RPC Methods

```protobuf
service SampleService {
  rpc CreateSample(CreateSampleRequest) returns (SampleResponse);
  rpc GetSample(GetSampleRequest) returns (SampleResponse);
  rpc ListSamples(ListSamplesRequest) returns (ListSamplesResponse);
  rpc UpdateSample(UpdateSampleRequest) returns (SampleResponse);
  rpc DeleteSample(DeleteSampleRequest) returns (DeleteSampleResponse);
}
```

### gRPC Status Code Mapping

| gRPC Status Code | Trigger Condition |
|---|---|
| `codes.InvalidArgument` | Invalid request parameters or empty ID on update/delete |
| `codes.NotFound` | Sample does not exist |
| `codes.AlreadyExists` | Resource already exists or version conflict |
| `codes.Unauthenticated` | Unauthorized request |
| `codes.PermissionDenied` | Forbidden action |
| `codes.Unavailable` | Underlying database unavailable |
| `codes.Internal` | Unexpected internal repository/usecase error |
