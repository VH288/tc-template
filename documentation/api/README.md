# Multi-Model API Documentation

Welcome to the `tc-template` API documentation directory. This directory is organized into isolated per-model subdirectories, providing modular API specifications and zero-configuration Swagger UI cataloging.

---

## Architecture & Directory Structure

API specifications and documentation are separated by domain model to prevent file bloat, avoid merge friction, and allow independent evolution of each domain service:

```
documentation/api/
├── README.md                  # Root documentation index (this file)
├── swagger-config.json        # Dynamic Swagger catalog (loaded by Swagger UI via CONFIG_URL)
├── sample/                    # Sample domain model documentation
│   ├── README.md              # Model overview and endpoint summary
│   ├── openapi.json           # OpenAPI 3.0.3 specification (JSON)
│   ├── openapi.yaml           # OpenAPI 3.0.3 specification (YAML)
│   └── sample-crud.md         # Detailed CRUD guide, envelopes, and cURL examples
└── othermodel/                # Example second model documentation
    ├── README.md              # Model overview and endpoint summary
    ├── openapi.json           # OpenAPI 3.0.3 specification (JSON)
    └── openapi.yaml           # OpenAPI 3.0.3 specification (YAML)
```

---

## Domain Model Documentation Index

| Model | Catalog Name | Spec Path | Documentation |
|---|---|---|---|
| **Sample** | `Sample API` | [`sample/openapi.json`](./sample/openapi.json) | [Overview](./sample/README.md) · [CRUD Guide](./sample/sample-crud.md) |
| **OtherModel** | `OtherModel API` | [`othermodel/openapi.json`](./othermodel/openapi.json) | [Overview](./othermodel/README.md) |

---

## Swagger UI Zero-Config Integration

The local Docker development stack (`../tools/docker-compose.yml`) configures Swagger UI with:

1. **Directory-level volume mount**:
   ```yaml
   volumes:
     - ../tc-template/documentation/api:/usr/share/nginx/html/docs:ro
   ```
2. **Dynamic configuration catalog**:
   ```yaml
   environment:
     CONFIG_URL: /docs/swagger-config.json
   ```

Swagger UI reads `swagger-config.json` and renders a **"Select a definition"** dropdown in the top bar. Each model's spec is served statically by Nginx at `http://localhost:8080/docs/<model>/openapi.json`.

Developers **never** need to edit `docker-compose.yml`, mount additional files, or restart containers when adding or updating model specifications.

---

## Adding a New Domain Model (Developer Workflow)

Adding documentation for a new model takes three simple steps:

### Step 1: Create the Model Folder
Create a new directory under `documentation/api/` with the name of your model:
```bash
mkdir -p documentation/api/<new-model>
```

### Step 2: Add OpenAPI Specification
Place your OpenAPI 3.0.3 specification file in the directory:
- `documentation/api/<new-model>/openapi.json` (or `openapi.yaml`)
- Optional: Add a `documentation/api/<new-model>/README.md` summarizing endpoints.

Ensure the `info.title` property in the spec reflects your desired catalog name (e.g., `"Ticket API"`).

### Step 3: Synchronize the Swagger Catalog
Run the synchronization script via Makefile:
```bash
make doc-sync
```

This automatically scans all model directories, extracts spec titles, and regenerates `documentation/api/swagger-config.json`. Refresh your browser at `http://localhost:8080` to view the new model in Swagger UI.

---

## Common Commands

```bash
# Synchronize Swagger UI catalog with all domain model specs
make doc-sync

# Start the HTTP server (default :8081) and gRPC server (default :7000)
make run-api

# Run all unit and integration tests
make test
```
