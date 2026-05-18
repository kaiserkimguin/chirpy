# Chirpy

A social media server built with Go. Users can register, authenticate, and post short messages called "chirps".

## Prerequisites

- **Go 1.26+**
- **PostgreSQL** — a running PostgreSQL instance
- **sqlc** — for generating type-safe Go code from SQL queries (`brew install sqlc`)
- **goose** — for running database migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

## Installation

1. **Clone the repository:**

```bash
git clone https://github.com/kaiserkimguin/chirpy.git
cd chirpy
```

2. **Install Go dependencies:**

```bash
go mod download
```

3. **Set up the database:**

Create a PostgreSQL database, then run the migrations:

```bash
goose postgres "YOUR_DB_URL" up
```

4. **Configure environment variables:**

Create a `.env` file in the project root with the following variables:

```
DB_URL=postgres://user:password@host:port/dbname?sslmode=disable
PLATFORM=dev
TOKEN_SECRET=your-secret-key
POLKA_KEY=your-polka-api-key
```

- `DB_URL` — PostgreSQL connection string
- `PLATFORM` — set to `dev` to enable admin reset endpoint
- `TOKEN_SECRET` — secret key used for signing JWT tokens
- `POLKA_KEY` — API key for validating Polka webhooks

5. **Generate sqlc code (if you modify SQL files):**

```bash
sqlc generate
```

6. **Build and run:**

```bash
go build -o chirpy .
./chirpy
```

The server will start on `http://localhost:8080`.

## Functionality

Chirpy is a minimal social media API that provides:

- **User management** — register and update user accounts
- **Authentication** — JWT-based authentication with access tokens (1 hour expiry) and refresh tokens (60 day expiry)
- **Chirps (posts)** — create, read, and delete short text posts (max 140 characters) with profanity filtering
- **Webhook integration** — supports Polka webhooks for user upgrade events
- **Admin dash** — internal metrics page and a dev-only data reset endpoint

## Environment Variables

| Variable       | Description                              |
|----------------|------------------------------------------|
| `DB_URL`       | PostgreSQL connection string (required)   |
| `PLATFORM`     | Environment mode (`dev` enables reset)   |
| `TOKEN_SECRET` | Secret for JWT signing (required)        |
| `POLKA_KEY`    | API key for Polka webhook verification   |

## API Endpoints

### Public Endpoints

| Method | Endpoint             | Description        |
|--------|----------------------|--------------------|
| GET    | `/api/healthz`      | Health check, returns `OK` |

### Frontend

| Method | Endpoint             | Description        |
|--------|----------------------|--------------------|
| GET    | `/app/`             | Serves the frontend (index.html) |

### User Management

| Method | Endpoint          | Description           | Auth Required |
|--------|-------------------|-----------------------|---------------|
| POST   | `/api/users`      | Create a new user     | No            |
| PUT    | `/api/users`      | Update current user's email and password | Yes (Bearer token) |

**Create user — `POST /api/users`**

Request body:
```json
{
  "email": "user@example.com",
  "password": "secretpass"
}
```

**Update user — `PUT /api/users`**

Headers:
```
Authorization: Bearer <access_token>
```

Request body:
```json
{
  "email": "newemail@example.com",
  "password": "newpassword"
}
```

### Authentication

| Method | Endpoint          | Description           |
|--------|-------------------|-----------------------|
| POST   | `/api/login`      | Authenticate and receive tokens |
| POST   | `/api/refresh`    | Refresh an expired access token using a refresh token as Bearer |
| POST   | `/api/revoke`     | Revoke a refresh token (pass refresh token as Bearer) |

**Login — `POST /api/login`**

Request body:
```json
{
  "email": "user@example.com",
  "password": "secretpass"
}
```

Response includes `token` (access JWT) and `refresh_token` in the user object.

**Refresh — `POST /api/refresh`**

Headers:
```
Authorization: Bearer <refresh_token>
```

**Revoke — `POST /api/revoke`**

Headers:
```
Authorization: Bearer <refresh_token>
```

### Chirps

| Method   | Endpoint                    | Description                | Auth Required |
|----------|-----------------------------|----------------------------|---------------|
| POST     | `/api/chirps`              | Create a new chirp         | Yes (Bearer token) |
| GET      | `/api/chirps`              | List all chirps            | No            |
| GET      | `/api/chirps?author_id=<uuid>` | List chirps by a specific author | No |
| GET      | `/api/chirps?sort=desc`     | Sort chirps by creation date descending | No |
| GET      | `/api/chirps/{chirpID}`     | Get a single chirp by ID   | No            |
| DELETE   | `/api/chirps/{chirpID}`     | Delete a chirp (owner only) | Yes (Bearer token) |

**Create chirp — `POST /api/chirps`**

Headers:
```
Authorization: Bearer <access_token>
```

Request body:
```json
{
  "body": "Hello, world!"
}
```

Chirps have a maximum length of **140 characters** and contain built-in profanity filtering.

### Webhooks

| Method | Endpoint                    | Description                |
|--------|-----------------------------|----------------------------|
| POST   | `/api/polka/webhooks`       | Handle Polka webhook events |

Headers:
```
Authorization: ApiKey <polka_api_key>
```

Request body for a user upgrade event:
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "uuid-here"
  }
}
```

### Admin

| Method | Endpoint               | Description                      |
|--------|------------------------|----------------------------------|
| GET    | `/admin/metrics/`      | View frontend visit count        |
| POST   | `/admin/reset/`        | Reset all users and metrics (dev only, requires `PLATFORM=dev`) |

## Project Structure

```
├── main.go                  # HTTP server, route definitions, handlers
├── types.go                 # Response types (User, Chirp, Token)
├── get_cleaned.go           # Chirp body sanitization
├── return_with.go           # JSON/error response helpers
├── index.html               # Frontend page served at /app/
├── sqlc.yaml                # sqlc configuration
├── sql/
│   ├── schema/              # Goose migration files
│   └── queries/             # SQL queries for sqlc code generation
├── internal/
│   ├── auth/                # Password hashing, JWT, refresh tokens
│   └── database/            # sqlc-generated query code
└── .env                     # Environment variables (not committed)
```

## Testing

```bash
go test ./...
```
