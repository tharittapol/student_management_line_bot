# CLAUDE.md — student_management_line_bot

LINE Bot backend for Thai student lesson scheduling. Written in Go 1.22, deployed on Google Cloud Run, uses Google Sheets as the primary database. All user-facing text is in Thai.

---

## Repository Layout

```
main.go                          # All HTTP handlers, LINE webhook logic, command parsing, scheduling (2430 lines)
class_schedule_sheets_store.go   # Primary storage backend — reads/writes the 3-tab Google Sheet
google_sheets_store.go           # Alternative normalized-schema Sheets store (rarely used in prod)
main_test.go                     # Unit tests with MockLessonStore
go.mod / go.sum
Dockerfile                       # Multi-stage Alpine build, runs as uid 65532
docker-compose.yml               # Local dev: app + ngrok tunnel
.env.example                     # All supported env vars with comments
db/
  schema.sql                     # PostgreSQL schema (reference only, not used in prod)
  README.md                      # Google Sheets tab structure docs (Thai)
```

All application logic lives in `main.go` — there are no sub-packages. Keep it that way unless explicitly asked to refactor.

---

## Storage Backends

`newLessonStore()` picks a backend at startup:

| Priority | Condition | Backend |
|----------|-----------|---------|
| 1 | `GOOGLE_SHEET_ID` is set | `ClassScheduleSheetsLessonStore` ← **prod/daily use** |
| 2 | `DATABASE_URL` is set | `PostgresLessonStore` |
| 3 | Neither | `MockLessonStore` ← used in unit tests |

The **ClassScheduleSheetsLessonStore** (in `class_schedule_sheets_store.go`) reads and writes three tabs:

- `Overview` — master student/enrollment data (columns A–N)
- `ตารางเรียน` — multi-week schedule grid
- `สัปดาห์นี้` — per-session rows for the current week (columns A–U, data starts row 3)

The `สัปดาห์นี้` tab is the one mutated by commands. **When writing back, the store copies formatting and data validation from an adjacent row** to preserve checkbox styles, bold text, background colors, and dropdown validators. This logic is fragile — do not change it without testing against a real sheet.

---

## Development Workflow

```bash
# Unit tests (primary dev loop — no real Sheets needed)
go test ./...

# Local server with real Sheets
cp .env.example .env   # fill in real values
go run .

# Local server with ngrok tunnel (for LINE webhook testing)
docker compose up --build
```

The test suite uses `MockLessonStore` and covers command parsing, date/time parsing, message formatting, LINE group ID validation, and HTTP handlers. Run `go test ./...` before every commit.

---

## HTTP Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/line/webhook` | LINE webhook — verifies HMAC-SHA256 signature, routes commands |
| `GET` | `/healthz` | Liveness probe — returns `200 ok` |
| `GET` | `/readyz` | Readiness probe — returns `200 ok` |
| `POST` | `/tasks/notify-daily` | Triggered by Cloud Scheduler at 09:00 Asia/Bangkok; requires `X-Task-Token` header or `?token=` query param |

LINE webhook signature verification uses `LINE_CHANNEL_SECRET` and HMAC-SHA256. **Do not weaken or bypass this check.**

The task endpoint uses `crypto/subtle.ConstantTimeCompare` against `NOTIFY_TASK_TOKEN`. Keep it that way.

---

## LINE Bot Commands

All commands require a `/` prefix and must come from a group listed in `LINE_GROUP_IDS`. Exception: `/groupid` works in any group (returns the group's C-ID).

### Read-only

| Command | Response |
|---------|----------|
| `/ตารางเรียน` | Current week's schedule from `สัปดาห์นี้`, grouped by Thai weekday |
| `/ข้อมูลนักเรียน` | Active students from `Overview`, deduplicated by person |
| `/วิธีใช้งาน` or `/help` | Full Thai usage guide |
| `/groupid` or `/group-id` | Current LINE group ID |

### Write commands

All write commands follow the pattern: `/action nickname firstname [date time_range]`

| Command aliases | Required args | Optional | Action |
|-----------------|--------------|---------|--------|
| `/อัพเดท` `/อัปเดท` `/update` `/เลื่อน` | nickname firstname | date time | Change lesson date/time |
| `/คอนเฟิร์ม` `/confirm` `/ยืนยัน` | nickname firstname | date time | Set confirmed = TRUE |
| `/ไม่คอนเฟิร์ม` `/not-confirm` `/unconfirm` | nickname firstname | date time | Set confirmed = FALSE |
| `/ลา` `/leave` `/absent` | nickname firstname | — | Set learning status to `ลา` |
| `/เข้าเรียน` `/attend` `/present` | nickname firstname | — | Set learning status to `เข้าเรียนปกติ` |

**Accepted date/time formats:**

```
9/5 13:00-15:00          # short (current year assumed)
9/5/2570 13:00-15:00     # Buddhist year (2570 → CE 2027, offset -543)
2027-05-09 13:00-15:00   # ISO
9 พ.ค. 13:00-15:00       # Thai month abbreviation
```

### Adding a new command

1. Add the command string(s) to the action normalization block in `processStaffCommand()` in `main.go`.
2. Implement the handler logic (look up student by nickname+firstname, call the appropriate store method, compose a Thai response string).
3. Add the command to the `/วิธีใช้งาน` (help) text.
4. Add test cases in `main_test.go` using `MockLessonStore`.

---

## Thai Localization

- All display text, error messages, and command names are in Thai.
- Buddhist year = CE year + 543 (e.g., 2027 CE = 2570 BE). Conversion lives in `parseDateStr()` and formatting helpers. The offset math is easy to break silently — test with explicit year inputs.
- Thai weekday/month name arrays are hardcoded. Index order is: Sunday = 0.
- The status emoji convention: ✅ = confirmed, ⏳ = pending.

---

## Core Data Models

```go
// One lesson session for a student (maps to a row in สัปดาห์นี้)
type StudentLesson struct {
    ID             string
    Nickname       string        // Thai nickname
    FirstName      string
    FullName       string
    Course         string
    TotalHours     int
    CompletedHours int
    SessionHours   int
    NextStart      time.Time
    NextEnd        time.Time
    ScheduleText   string
    Confirmed      bool
    LearningStatus string        // "เข้าเรียนปกติ" | "ลา" | ""
    UpdatedAt      time.Time
}

// Aggregated view of a student across all courses (for /ข้อมูลนักเรียน)
type StudentScheduleSummary struct {
    Nickname        string
    FirstName       string
    FullName        string
    Course          string        // comma-separated if multiple courses
    TotalHours      int
    CompletedHours  int
    DefaultSchedule string
    PastLessons     string
    NextLessons     string
    ScheduleNotes   string
}
```

---

## Environment Variables

```env
# Required for production
LINE_CHANNEL_SECRET=...
LINE_CHANNEL_ACCESS_TOKEN=...
LINE_GROUP_IDS=Cxxx,Cyyy              # comma/semicolon/space separated
GOOGLE_SHEET_ID=...
GOOGLE_SERVICE_ACCOUNT_JSON_BASE64=... # base64-encoded service account JSON

# Optional / tuning
LINE_STAFF_GROUP_ID=...               # legacy single-group fallback
GOOGLE_SERVICE_ACCOUNT_JSON_PATH=...  # alternative: local path to JSON file
GOOGLE_SHEETS_INIT_SCHEMA=true        # validate tab names on startup
RUN_DAILY_ON_START=false              # send schedule once at startup
DISABLE_DAILY_SCHEDULER=false         # true = rely on Cloud Scheduler instead of in-process cron
NOTIFY_TASK_TOKEN=...                 # secret for /tasks/notify-daily
PORT=8080
TZ=Asia/Bangkok
```

---

## Message Formatting Rules

- LINE text messages are capped at 4500 characters (`lineTextMaxLength`).
- Long responses are auto-split and sent in batches of up to 5 messages (`lineMessageBatchLimit`).
- Split points prefer newlines to avoid cutting mid-line.

---

## Critical Implementation Notes

1. **Google Sheets format copying** — `class_schedule_sheets_store.go` copies row formatting (bold, background, checkbox, dropdown) from an adjacent row when inserting or updating a `สัปดาห์นี้` row. This is the most fragile part of the codebase. If you touch this logic, test with a real spreadsheet.

2. **Mutex protection** — all store implementations hold a `sync.Mutex` during reads and writes. Do not add concurrent Sheets calls outside this lock.

3. **Student matching** — `strings.EqualFold()` is used for nickname and firstname matching. New lookup code must also be case-insensitive.

4. **Buddhist year offset** — +543 CE→BE, −543 BE→CE. Always test with a year near a century boundary to catch off-by-one errors.

5. **LINE group ID validation** — valid IDs start with `C` and pass `isLikelyLineTargetID()`. The `/groupid` command is intentionally unrestricted (no group allowlist) so admins can discover a new group's ID before adding it.

6. **Google Sheets auth** — RSA JWT flow using a service account. The token is cached with a 1-minute refresh buffer. The service account email must have edit access to the spreadsheet.

---

## Deployment (Google Cloud Run)

```bash
# Build and push image
docker build -t gcr.io/PROJECT/student-bot .
docker push gcr.io/PROJECT/student-bot

# Deploy
gcloud run deploy student-bot \
  --image gcr.io/PROJECT/student-bot \
  --region asia-southeast1 \
  --set-env-vars "..." \
  --allow-unauthenticated
```

- Set `DISABLE_DAILY_SCHEDULER=true` on Cloud Run and create a Cloud Scheduler job to `POST /tasks/notify-daily` at `0 9 * * *` Asia/Bangkok with `X-Task-Token` header.
- Store secrets (tokens, service account JSON) in Google Secret Manager and mount as env vars.

---

## Testing

```bash
go test ./...         # run all tests
go test -v ./...      # verbose output
go test -run TestName # run a specific test
```

Tests are in `main_test.go` and use `MockLessonStore` (seeded with 10 students). No external services needed. Add test cases to `main_test.go` whenever you add a new command or change parsing logic.
