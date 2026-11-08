# Sample Model API Documentation

This directory contains the API documentation and OpenAPI specifications for the **Sample** domain model in `tc-template`.

## Contents

- **`openapi.json`**: OpenAPI 3.0.3 specification in JSON format (served at `/docs/sample/openapi.json` in Swagger UI).
- **`openapi.yaml`**: OpenAPI 3.0.3 specification in YAML format.
- **`sample-crud.md`**: Detailed architectural guide, error envelope definitions, and cURL examples for Sample CRUD operations.

---

## Endpoints Summary

Base URL: `http://localhost:8081/api/v1`

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `POST` | `/samples` | Create a new sample record | Yes (`X-API-Key` if configured) |
| `GET` | `/samples` | List sample records with pagination and sorting | Yes (`X-API-Key` if configured) |
| `GET` | `/samples/{id}` | Retrieve a sample record by ID | Yes (`X-API-Key` if configured) |
| `PUT` | `/samples/{id}` | Update an existing sample record by ID | Yes (`X-API-Key` if configured) |
| `DELETE` | `/samples/{id}` | Delete a sample record by ID | Yes (`X-API-Key` if configured) |

---

## Swagger UI Access

When running the local development stack via Docker Compose (`../tools/docker-compose.yml`), this specification is available in Swagger UI under:
- Catalog Name: `Sample API`
- URL: `http://localhost:8080/docs/sample/openapi.json`
