# fast-go

A colorful terminal speed test inspired by [Fast.com](https://fast.com/).

Pop-style TUI with dynamic mood color schemes — every run feels different.

## Demo

```
◆ SUNSET ◆

╔══════════════════════════════════════════════════════════╗
║                    123.4 Mbps                            ║
╚══════════════════════════════════════════════════════════╝

  ████████████████████████████████████████████░░░░░░░  75%
  Testing... 18.8 MB / 25.0 MB  (7.5 s)

  ↓ 118.2 Mbps    ↑ --    ping --
```

## Features

- **Real speed measurement** — downloads from OVH/Tele2/Hetzner test files, parallel connections
- **Running average** — displays cumulative average throughput over the test duration
- **7 mood color schemes** — VIVID, PASTEL, NEON, MONO, RETRO, OCEAN, SUNSET
- **Animated counter** — smooth ease-out counting from 0 to measured speed
- **Progress gauge** — real-time fill bar with percentage
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
fast-go.exe
```

### Controls

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `r` | Retry test |
| `m` | Cycle mood (VIVID → PASTEL → NEON → …) |
| `d` | Toggle details |
| `s` | Save result to JSON |

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
