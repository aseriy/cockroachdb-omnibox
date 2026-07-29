# Visualize — Design

A self-contained log viewer for the omnibox demo: a single Go binary that tails
application log files and serves a browser UI showing them as parallel lanes.

## Principles

- **Decoupled from log producers.** The visualizer knows nothing about the
  applications that write the logs. Its only contract is the log line format
  (below). Event types, message contents, and their meanings are opaque;
  anything the operator knows about their logs is expressed in `visualize.yml`,
  never in code.
- **Config-driven.** All deployment knowledge — lanes, colours, display
  defaults — lives in `visualize.yml`. The UI derives everything from the
  backend's two endpoints.
- **Single artifact.** Backend and frontend ship as one binary; the compiled
  frontend is embedded and served from it. No separate web server, port, or
  deployment step.

## Log record contract

A record is one line of the form:

    date time LEVEL [EVENT] message

with a Python-logging-style asctime timestamp (comma milliseconds), e.g.:

    2026-07-27 03:02:14,813 INFO [DATAPOINT] Logged at ...: {...}

The backend parses records positionally into
`{ts, epoch, level, type, msg}` where `epoch` is the timestamp in Unix
milliseconds and `type` is the bracketed event with brackets removed.

Filtering at the source:

- Only records that parse cleanly are streamed; non-matching lines
  (including any multi-line spillover) are silently dropped.
- Only `INFO` records are streamed. Today's producers log nothing else;
  the filter makes the contract explicit.

Consequently every record delivered to the UI is well-formed and carries a
valid `epoch`.

## Backend

Go, three packages:

- `config` — reads `visualize.yml`, no validation or clamping. Misconfigured
  values (e.g., an excessive `buffer_size`) are the operator's to correct;
  practical ranges are documented in the config file itself.
- `tail` — follows one log file: seeds the last `buffer_size` records, then
  streams new ones, surviving file rotation.
- `main` — HTTP server on `listen`:
  - `GET /` — the embedded frontend.
  - `GET /api/config` — the parsed config as JSON.
  - `GET /api/stream/{lane}` — Server-Sent Events; one parsed record per
    frame. Unknown lane → 404. On connect, replays the seed history, then
    follows live.

## Configuration (`visualize.yml`)

| Key             | Meaning                                                        |
|-----------------|----------------------------------------------------------------|
| `listen`        | Backend address, e.g. `localhost:8080`.                        |
| `buffer_size`   | Records kept per lane — backend seed and browser buffer alike. Documented practical range; no enforcement. |
| `scheme`        | `light` or `dark`; initial UI theme.                           |
| `default_color` | Colour of the type token for types absent from `messages`.     |
| `messages`      | Map of event type → type-token colour.                         |
| `lanes`         | Roster of available lanes: `name` + log file `path`. Any number; the deployment decides. |
| *display defaults* | Initial mode, initial orientation (mode 1), and mode 2 bucket `resolution` — exact keys finalized in the config implementation phase. |

## Frontend

Vue 3, built with Vite; `dist/` embedded into the Go binary via `go:embed`.

### Lane slots

The UI has exactly three lane **slots**. Each slot has a dropdown selecting
any lane from the config roster — any 1–3 lanes, any order — or **off**
(blank pane). Defaults: the first three lanes in config order; fewer than
three in the config fills what's available; an empty roster leaves all slots
blank. Each active slot owns one `/api/stream/{lane}` connection.

### Record rows

A row is one record: the type token (no brackets), coloured per
`messages`/`default_color`, followed by the message in the scheme's normal
foreground. `ts`, `epoch`, and `level` are never displayed — timestamps
exist only to order and align records. One record per line, clipped at the
pane edge; no wrapping. Rows are rendered virtualized, so buffer sizes well
beyond the default stay smooth.

### Mode 1 — independent lanes

Three panes (columns or rows; orientation switchable), each showing its
lane's records in arrival order. Each pane scrolls independently like a
terminal: scrollback through the full buffer, auto-follow when at the tail,
per-pane jump-to-tail.

### Mode 2 — timeline

Vertical columns only. A borderless three-column table; each row is one
**occupied** time bucket. Records are assigned to buckets by rounding their
timestamp to the configured `resolution`; buckets where no displayed lane
logged anything do not render at all — the timeline is ordinal, never
scaled. A lane with no events in an occupied bucket renders an empty cell,
preserving horizontal co-occurrence. Multiple events from one lane in one
bucket stack within the cell in epoch order; coarser resolutions trade
timeline length for denser cells. The table scrolls as one unit — one
scroll position, one tail-follow state, one global jump-to-tail. Scrollable
extent is whatever the lane buffers hold (up to 3 × `buffer_size` records).

Switching a slot's lane in mode 2 clears that column and populates it from
the switch moment forward; the replaced lane's history does not linger and
the new lane's pre-switch history is not backfilled. (In mode 1 a switch
simply reseeds the independent pane.)

### Control bar

Mode toggle, orientation toggle (mode 1 only), light/dark toggle (seeded
from `scheme`), and **RESET**.

### Reset and failure

Disconnection is expected to be rare (same executable, same container), so
recovery is coarse and total: any dropped stream triggers a full reset —
all slots disconnect, reconnect, and reseed. RESET triggers the same path
manually. A lane that cannot be opened at all stays silently blank; RESET
is the operator's retry.

## Delivery

Implemented in phases with a review checkpoint after each; design and
subsequent phases adjust at checkpoints as needed. Phase 1 finalizes and
implements the configuration.
