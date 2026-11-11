---
name: fast-go
description: >
  Fast.com 風のカラフルなターミナル速度測定 TUI（Go / Bubble Tea）を
  ビルド・実行・拡張するスキル。OVH/Tele2/Hetzner のテストファイルを
  並列 HTTP でダウンロードし回線速度をリアルタイム表示する。
  起動キーワード：「速度測定」「回線速度」「fast-go」「speed test」
  「ネット速度」「fast.com」などで発動。
  CLI: `go run .`（`build.ps1` で .exe ビルド）
---

# fast-go 速度測定 TUI

## 使い方

```bash
go run .        # 実行（Bubble Tea AltScreen で動く）
go build .      # ビルド → build.ps1 は `go build -o fast-go.exe .`
go vet ./...    # 静的解析
go build ./...  # 全パッケージのコンパイル確認
```

概要は `GUIDE.md` に記載。テストファイルは none（`tui/view_test.go` のみ存在）。

## 主要ファイル

- `GUIDE.md` — アーキテクチャ・データフロー・コマンド一覧
- `fast-results.jsonl` — 測定結果ログ（download_mbps / lat / lon / timestamp）
- `build.ps1` — Windows ビルド
- `internal/tui/`（モデル・レンダー・テーマ）、`internal/speedtest/`（計測エンジン）、
  `internal/util/`（フォーマット・位置情報）
- `.github/workflows/build.yml` — CI ビルド

## 関連

- speedtest（同系のカラフル速度測定。こちらは単体技能として運用）
- wifi-mac-rotate-skill（TSUBAME 速度計測）