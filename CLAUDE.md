# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

日本語エンジニア向け音声コーディング入力システム。ホットキーで録音 → Gemini API で文字起こし → クリップボード経由でペースト。プログラミング用語を修正不要の精度で一発変換する。

- ベースプロジェクト（Python参照実装）: `/Users/ms25/project/vibescribe`
- 本リポジトリは Go への再構築用。現在は設計ドキュメントのみ（実装未着手）

## 技術スタック

| 項目 | 選定 |
|------|------|
| 言語 | Go（単一バイナリ配布） |
| UI | Native-first（Flutter不採用） |
| API | Gemini API 直接リクエスト（バックエンドサーバーなし） |
| データ | ローカルファイル（JSON/WAV）、将来 SQLite 検討 |
| 対象 | macOS / Windows / Linux（デスクトップ先行） |

## アーキテクチャ

### Core/Adapter 分離パターン

```
Go Core Logic（共通）
  ├── Audio Recorder    ... 録音 → WAV
  ├── Transcriber       ... Gemini API 呼び出し
  └── Text Postprocess  ... システムプロンプト + 用語辞書

OS Platform Adapters（darwin/windows/linux）
  ├── Global Hotkey     ... F15 等のホットキー監視
  ├── Clipboard/Paste   ... クリップボード操作 + Cmd+V
  └── Tray/Menu UI      ... メニューバー/通知領域
```

### 処理フロー

ホットキー押下 → 録音(WAV 16kHz mono) → Gemini API → テキスト → クリップボード → ペースト → 履歴保存

### 設定ファイル

- アプリ設定: `~/.voicecode/settings.json`
- 履歴: `~/.voicecode/history/`
- API キー: `~/.voicecode/.env`（`GOOGLE_API_KEY`）

## ドキュメント構成

| パス | 内容 |
|------|------|
| `docs/overview.md` | サービス概要・価値提案 |
| `docs/foundation/goals.md` | ゴール・MVP定義・KPI |
| `docs/requirements/features.md` | 機能要件（MVP優先度付き） |
| `docs/tech/stack.md` | 技術選定理由 |
| `docs/tech/architecture.md` | システム構成・通信方式 |
| `docs/tech/er-diagram.md` | データ設計 |
| `docs/api/endpoints.md` | Gemini API 仕様 |
| `docs/design/sitemap.md` | 画面一覧・遷移図 |

## 開発方針

- PoC で 3 機能を先に検証: ホットキー監視、クリップボード貼り付け、Gemini 連携
- Go の OS 連携ライブラリは PoC 結果で最終決定
- 優先順位: 精度 > 貼り付け成功率 > 速度
- KPI: 文字起こし成功率 98%、貼り付け成功率 99%、応答時間 6 秒以内

## 環境変数

| 変数 | 必須 | 説明 |
|------|------|------|
| `GOOGLE_API_KEY` | 必須 | Google AI Studio API キー |
| `VOICECODE_GEMINI_MODEL` | 任意 | モデル指定（デフォルト: auto） |
| `VOICECODE_ENABLE_PROMPT_CACHE` | 任意 | プロンプトキャッシュ（デフォルト: true） |
