# AGENTS.md — fast-go

## What is this?

A colorful terminal speed test (inspired by Fast.com) with a pop-style TUI built with Bubble Tea. Downloads from OVH/Tele2/Hetzner test files using parallel HTTP connections, measures throughput in real-time, and displays results with animated counters and gauges.

## Project structure

```
fast-go/
├── main.go                         # Entry point — creates tea.Program with AltScreen
├── internal/
│   ├── tui/                        # All rendering & user interaction
│   │   ├── model.go                # Bubble Tea Model + Init/Update/View (all in one file)
│   │   ├── animation.go            # AnimFloat (exponential ease), EaseOutCubic
│   │   └── theme.go                # Mood enum, Palette struct, HSL color generator
│   ├── speedtest/
│   │   ├── tester.go               # Speed test engine — parallel HTTP workers, avg calculation
│   │   └── result.go               # SpeedResult + SpeedSample types
│   └── util/
│       └── format.go               # FormatMbps, FormatBytes, FormatDuration
├── go.mod / go.sum
├── PRD.md                          # Full product requirements doc (Japanese) — useful reference
└── README.md
```

**Note:** The PRD lists `view.go` and `update.go` as separate files but they were merged into `model.go`. There are no separate view/update files.

## Essential commands

| Command | Purpose |
|---------|---------|
| `go build -o fast-go.exe .` | Build |
| `go run .` | Run |
| `go vet ./...` | Static analysis |
| `go build ./...` | Verify all packages compile |

No Makefile, no test files, no CI config, no linter config.

## Architecture & data flow

```
main.go
  └─ tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
       │
       ├─ Init() → starts tick (30fps) + speed test goroutine
       │
       ├─ Update(msg)
       │    ├─ tea.WindowSizeMsg → resize
       │    ├─ tea.KeyMsg → q/r/m/d/s commands
       │    ├─ startTestMsg → captures cancel func + channel
       │    ├─ speedSampleMsg → updates actualSpeed, progress, bytesRecv, elapsed
       │    └─ tickMsg → drives AnimFloat interpolation (30fps)
       │
       └─ View() → renders mood label, big number (double-border box), gauge bar, progress, status line
            └─ uses lipgloss for styling, AnimFloat for smooth display values
```

The speed test runs in a **background goroutine** (`speedtest.Run`) that:
1. Pings each URL via HEAD to find a reachable server
2. Spawns 4 parallel HTTP GET workers reading 64KB chunks
3. Reports progress on a channel every 200ms
4. Auto-stops at 15s max, or earlier if <3% variance over last samples
5. Closes the channel when done; model reads via `nextSampleCmd`

## Key gotchas & non-obvious patterns

- **No view.go / update.go**: Unlike typical Bubble Tea layouts, everything lives in `model.go` — View() and Update() are methods on the Model struct in that one file. Don't create separate files unless refactoring.
- **Upload & latency fields exist but aren't measured**: `SpeedResult` has `UploadMbps`, `LatencyMs`, `LoadedLatencyMs`, `BytesSent` — but the tester only measures download. These are placeholders for future use.
- **Display speed != actual speed**: `displaySpeed` is an `AnimFloat` that smoothly eases toward `actualSpeed` using exponential smoothing (`smoothFactor = 4.0`). The raw measurement goes through `AnimFloat.Update(dt)` before rendering.
- **Dual tick system**: `tickMsg` runs at 30fps for animation interpolation, while speed samples arrive asynchronously via `nextSampleCmd`. They're independent — ticks don't block on network.
- **Random mood on each run**: `InitialModel()` calls `RandomMood()` which picks one of 7 moods. Each mood's `GeneratePalette()` produces random HSL values within mood-specific ranges. The same mood looks different every time.
- **`r` (retry) fully resets everything**: The retry handler manually zeroes all fields, generates a new random mood, and returns `m.startTestCmd()`. It cancels the old context first.
- **JSON save format**: Pressing `s` writes to `fast-result-YYYYMMDD-HHMMSS.json` with a hand-rolled `fmt.Fprintf` (no JSON encoder). The file is `*.json` in `.gitignore`.
- **Pickiest URLs**: Test URLs are hardcoded in `tester.go:11-15`. `pickURL` does a HEAD request and picks the first that returns 200. If none reachable, the test errors with "no reachable test URL".
- **Windows binary in .gitignore**: `fast-go.exe` and `fast-go.exe~` are gitignored along with `*.exe`.
- **Progress cap at 25MB**: `progress = min(totalBytes / 25MB, 1.0)` — even if the server sends more, progress never exceeds 100%.
- **Go 1.25.5**: The module uses a very recent Go version. `golang.org/x/exp` appears as an indirect dependency in go.sum but isn't imported in any source file.
- **View modes (`v` key)**: Three view modes — `WIDE` (80% of terminal width), `DS` (~50 cols), `GB` (~34 cols). Mode name shown next to mood label. Uses `m.viewWidth()` helper that caps `m.width` per-mode. Defaults to WIDE.

## Naming & conventions

- **Package naming**: `tui`, `speedtest`, `util` — all lowercase, no underscores.
- **Exported types**: PascalCase (`Mood`, `Palette`, `State`, `ViewMode`, `SpeedResult`, `SpeedSample`, `AnimFloat`).
- **Unexported types**: camelCase (`displaySpeed`, `displayGauge`, `stopTest`, `speedCh`).
- **Mood enum**: `iota` const group starting at 0, with explicit `moodCount` terminal sentinel for bounds.
- **Model fields**: All unexported (lowercase) — typical Bubble Tea style where the model is package-internal.
- **Format helpers**: `FormatMbps`, `FormatBytes`, `FormatDuration` — all return strings with units baked in.

## Style guide

- Uses `lipgloss` for all styling (colors, borders, alignment, width).
- Direct `fmt.Fprintf` to `strings.Builder` in View() — no template rendering.
- HSL colors generated via a custom `hsl()` function that converts to hex strings like `#FF006E` (no lipgloss color adapter needed).
- Error strings are plain lowercase ("no reachable test URL") — no capitalization.
- Comments are sparse; PRD.md serves as the design doc.

## PRD reference

The `PRD.md` (in Japanese) contains detailed specs including the mood system design, animation specifications, and future phase plans. Key unimplemented features from Phase 3:
- Multi-server measurement
- History sparklines
- speedtest.net/ookla compatibility
- Theme customization via config file
- iperf3 mode
- Sound effects (beep)
