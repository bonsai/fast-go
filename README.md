# fast-go

A colorful terminal speed test inspired by [Fast.com](https://fast.com/).

Pop-style TUI with dynamic mood color schemes — every run feels different.

## Demo

### NORMAL mode

```
⚡ FAST-GO                                       ◆ NEON · NORMAL
────────────────────────────────────────────────────────────────

                █████  ██   ██ ███████    ██████
                    ██ ██   ██ ██              ██
                 ████  ███████ ██████      █████
               ██           ██      ██         ██
               ███████      ██ ██████  ██ ██████

                      Mbps  ·  ↓ DOWNLOAD

            ▃▃▄▄▅▅▅▆▆▆▇▇█▃▃▄▄▅▅▅▆▆▆▇▇█▃▃▄▄▅▅▅▆▆▆▇▇█▃

   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌  73%
                  ⠋ 18.0 MB / 25.0 MB · 7.0 s

    r retry · m mood · v mini · d details · s save · q quit
```

### MINI mode

```
       ⚡ FAST-GO ◆ OCEAN
╭────────────────────────────────────╮
│ ↓ 245.3 Mbps            ▅▅▅▆▆▆▇▇█▃ │
│ ━━━━━━━━━━━━━━━╌╌╌╌╌╌  73% · 7.0 s │
╰────────────────────────────────────╯
 r retry · v normal · s save · q quit
```

## Features

- **Real speed measurement** — downloads from OVH/Tele2/Hetzner test files, parallel connections
- **Running average** — displays cumulative average throughput over the test duration
- **Two view modes** — NORMAL (big block digits + sparkline) and MINI (compact widget panel), toggle with `v`
- **Big block digits** — speed readout rendered as 5-row block numerals with a vertical color gradient
- **Speed sparkline** — live history of throughput samples
- **Gradient gauge** — progress bar that fades between palette colors
- **7 mood color schemes** — VIVID, PASTEL, NEON, MONO, RETRO, OCEAN, SUNSET
- **Animated counter** — smooth ease-out counting from 0 to measured speed
- **Keyboard controls** — retry, switch mood, show details, save results
- **Zero external API dependency** — no fast.com API, no third-party services

## Install

```bash
go install github.com/bonsai/fast-go@latest
```

Or build from source:

```bash
git clone https://github.com/bonsai/fast-go.git
cd fast-go
go build -o fast-go.exe .
```

## Usage

```bash
fast-go.exe         # NORMAL mode
fast-go.exe -mini   # start in MINI mode
```

### Controls

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `r` | Retry test |
| `v` | Toggle view mode (NORMAL ⇄ MINI) |
| `m` | Cycle mood (VIVID → PASTEL → NEON → …) |
| `d` | Toggle details (NORMAL mode) |
| `s` | Save result to `fast-results.jsonl` |

## Moods

| Mood | Vibe |
|------|------|
| VIVID | Primary colors, energetic |
| PASTEL | Soft, cute |
| NEON | Dark bg, fluorescent |
| MONO | Monochrome gradient, chic |
| RETRO | Vintage, warm |
| OCEAN | Cool blue tones |
| SUNSET | Orange → purple gradient |

Each mood generates random hues within its range, so the same mood looks different every time.

## How it works

1. Pings known test file URLs (OVH, Tele2, Hetzner) to find a reachable server
2. Starts 4 parallel HTTP connections to download a large file
3. Every 200ms, calculates cumulative average: `(total bytes × 8) / elapsed seconds`
4. Smoothly animates the displayed number using exponential easing
5. Stops automatically when speed stabilizes (<3% variance) or after 15s

No external API calls — everything is self-contained.

## License

MIT
