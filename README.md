# github-releases-notifier

![CI](https://github.com/reqlane/github-releases-notifier/actions/workflows/ci.yml/badge.svg)

A REST API application that that allows users to subscribe to email notifications about new releases of a chosen Github repository.

## Workflow

On startup server loads configuration, connects to db, runs migrations. All dependencies are wired up: database connection instance, Github API client, SMTP notifier, repository, handler and service, and scanner. Passes them down the chain via constructor injection. I tried to avoid loose coupling as much as possible using interfaces for scanner, notifier, repository and github client.

Two things work concurrently:

- **HTTP server** handles subscription management: subscribe, confirm, unsubscribe, get by email.
- **FixedRateScanner** runs in a goroutine, polling Github for new releases at a steady, controlled rate depending on github API token presense.

### Server

Consists of handler and service responsible for HTTP and business logic respectively. Across 4 endpoints that API handles, each requested scenario is handled correctly with help of domain error types. Application was designed to prevent any leaking responsibilities and mixing concerns, meaning repository is responsible for database operations only and no other place is. Same with service not knowing about http/db and handler not knowing about anything except http.

### Database

Database used is MariaDB

Two tables `repos` and `subscriptions`.
`repos` contain Github repositories that at least one user has subscribed to, and the last release tag the scanner detected. When a new subscription comes in for a repo that isn't tracked yet, a record is created. `repo` is indexed because every subscription lookup and scanner query searches by repo name.

`subscriptions` stores the relationship between email and repo, along with the confirmed state and tokens used for confirm/unsubscribe links in emails.
A subscription starts unconfirmed. `confirm_token` is nullable and is set to NULL after confirmation, making it usable once. `unsubscribe_token` is permanent and stays for the lifetime of the subscription. They are not encrypted with say argon2 because it could be possible only if we needed to send it once at the beginning, which is not the case. The unique constraint on `(email, repo_id)` prevents duplicate subscriptions at the database level, but application check is prioritized.

### Response format

Response on /subscriptions is in swagger specified format [Subscription]

Other responses return JSON in the format:

```json
{
  "status": "error",
  "message": "Validation failed",
  "details": {
    "email": "value is empty",
    "repo": "bad_repo is invalid github repo, must be in owner/repo format"
  }
}
```

`status` is always present - "success"/"error". `message` is also always included. `details` is only included if request body in correct format failed validation, where each key is the JSON field name that failed and value carries descriptive message.

Regarding POST /subscribe endpoint: each request triggers github API, and because of that, 429(403) is possible if app is out of limit. If such scenario occurs, the application returns 503 Service temporary unavailable, because returning 429 would be a mistake, user could do just one request and that's it.

### Github client

Responsible for all operations with Github API.
Rate limiting is handled gracefully according to guidelines: if Github returns a rate limit response 429 (or 403 without token), client blocks concurrent-safely and returns an error with reset time.

### The scanner

Fetches all tracked repos that have at least one confirmed subscriber, then starts each repo check in its own goroutine while sleeping between starts to maintain a fixed request rate (default 60 requests/min, 1/min if no Github API token is configured). When a new release of repo is detected, it updates its tag and sends email notifications to all confirmed subscribers of that repo.

If rate limit error received, goroutine sends a pause signal through a buffered channel, and the scanner empties it fully and sleeps until the reset time before continuing.

## Running with Docker

You can run application with Docker. It will manage MariaDB and the API separately and database data is stored on shutdown/restart:

```bash
docker compose up --build
```

The server will be available at `http://localhost:{SERVER_PORT}`

## Configuration

`.env.example` is example of how .env file in root directory should look like

| Variable | Description | Example |
|---|---|---|
| `DB_USER` | Database user | `root` |
| `DB_PASSWORD` | Database password | `pass` |
| `DB_NAME` | Database name | `github_releases_notifier` |
| `DB_HOST` | Database host | `127.0.0.1` |
| `DB_PORT` | Database port | `3306` |
| `SERVER_PORT` | Port the server listens on | `3000` |
| `SERVER_BASE_URL` | Public base URL of the server, used in emails | `http://localhost:3000` |
| `GITHUB_API_TOKEN` | (optional) GitHub personal access token, but without it the scanner is throttled to 1 req/min | `ghp_xxx...` |
| `SMTP_HOST` | SMTP server host | `smtp.gmail.com` |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_USERNAME` | SMTP login / sender address | `you@gmail.com` |
| `SMTP_PASSWORD` | SMTP password or app password | `abcd qwer tyiu fghj` |

## API

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/subscribe` | Subscribe to release notifications |
| `GET` | `/api/confirm/:token` | Confirm email subscription |
| `GET` | `/api/unsubscribe/:token` | Unsubscribe from release notifications |
| `GET` | `/api/subscriptions?email=` | Get subscriptions for an email |

## Tests

Unit tests cover all business logic across two core packages `service` and `scanner` with all dependencies replaced by mocks.

**Service** (`internal/api/service`):

- `Subscribe` — success, invalid email (5 edge cases), invalid repo format (5 edge cases), subscription already exists, Github repo not found, repo with no releases stilll succeeds, repo not yet tracked is created, race condition on repo creation handled correctly, confirmation email failure propagates
- `Confirm` — success on 3 valid tokens, invalid token format (5 edge cases), token not found
- `Unsubscribe` — success on 3 valid tokens, invalid token format (5 edge cases), tokjen not found
- `GetSubscriptions` — success, invalid email (5 edge cases), database error propagates

**Scanner** (`internal/scanner`):

- `scan` — no repos = no GitHub calls, database error = no GitHub calls
- `checkRepo` — new release updates tag and notifies all subscribers, same tag = no update and no notifications, repo with no releases at all = no update and no notifications, rate limit response sends correct pause signal to the pause channel, tag update failure = no notifications, fetching targets failure = no notifications, single notify failure does not stop remaining subscribers from being notified

```bash
go test -v ./...
```

## CI

Every push runs two jobs in parallel via GitHub Actions:

- **Lint** — runs `golangci-lint` linter for code
- **Test** — runs `go test -v ./...` all tests

> I might do extras despite deadline has passed, so if you review my code after that and reckon it unfair, the last commit before deadline is
> [f05212d](https://github.com/reqlane/github-releases-notifier/commit/f05212dc2068ada362c1a82031d948698f6ac6f8)