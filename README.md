# VoiceCode

日本語エンジニア向け音声コーディング入力システム。ホットキーで録音し、Gemini API で文字起こしして、クリップボード経由でペーストする。プログラミング用語を修正不要の精度で一発変換する。

## 必要環境

- macOS (darwin/arm64 or amd64)
- Go 1.22+
- PortAudio (`brew install portaudio`)
- Google AI Studio API キー

## セットアップ

```bash
# 依存インストール
brew install portaudio

# API キー設定
export GOOGLE_API_KEY="your-api-key"

# ビルド
go build -o voicecode ./cmd/voicecode
```

## macOS 権限設定（初回のみ）

GUI モードで使うには以下の権限が必要。

### アクセシビリティ（Cmd+V ペーストに必須）

**システム設定 → プライバシーとセキュリティ → アクセシビリティ** で、`voicecode` を実行するターミナルアプリ（Terminal.app / iTerm2 / Warp 等）を許可する。許可後はターミナルを再起動。

この権限がないと、文字起こしは成功するがペーストされない。

### マイク（録音に必須）

初回録音時に macOS がマイクアクセスの許可を求める。許可する。

### トラブルシューティング

文字起こしは成功するがペーストされない場合:

```bash
# クリップボードにセットされているか確認
./voicecode &
# 録音→文字起こし完了後に:
pbpaste
```

- `pbpaste` で結果が出る → アクセシビリティ権限の問題。上記の手順で許可する
- `pbpaste` で結果が出ない → クリップボード書き込みの問題。issue を報告してください

## 使い方

### GUI モード（メニューバー常駐）

```bash
./voicecode
```

ホットキー（デフォルト: F15）で録音開始/停止。文字起こし結果が自動でペーストされる。F15 キーがないキーボードの場合は `~/.voicecode/settings.json` で `hotkey` を変更する（例: `"f13"`, `"f14"`）。

### CLI モード（WAV ファイル文字起こし）

```bash
./voicecode transcribe <wav-file>
```

## 設定

設定ファイル: `~/.voicecode/settings.json`

```json
{
  "hotkey": "f15",
  "restore_clipboard": true,
  "max_recording_duration": 120,
  "push_to_talk": false
}
```

| 項目 | デフォルト | 説明 |
|------|-----------|------|
| `hotkey` | `f15` | トリガーキー |
| `restore_clipboard` | `true` | ペースト後にクリップボードを復元 |
| `max_recording_duration` | `120` | 最大録音秒数（10-300） |
| `push_to_talk` | `false` | キー押下中のみ録音 |

### ユーザー辞書

`~/.voicecode/dictionary.txt` にタブ区切りで変換ルールを定義:

```
Kubernetes	クバネティス,クーバネティス
Docker	ドッカー
```

## 環境変数

| 変数 | 必須 | 説明 |
|------|------|------|
| `GOOGLE_API_KEY` | Yes | Google AI Studio API キー |
| `VOICECODE_GEMINI_MODEL` | No | モデル指定（デフォルト: auto） |
| `VOICECODE_THINKING_LEVEL` | No | Thinking レベル: minimal/low/medium/high |
| `VOICECODE_ENABLE_PROMPT_CACHE` | No | プロンプトキャッシュ（デフォルト: true） |
| `VOICECODE_PROMPT_CACHE_TTL` | No | キャッシュ TTL（デフォルト: 3600s） |

## アーキテクチャ

```
cmd/voicecode/          エントリポイント（CLI + GUI）
internal/
  core/                 プラットフォーム非依存ロジック
    transcriber/        Gemini API 連携（モデル解決・リトライ・キャッシュ）
    prompt/             システムプロンプト・ユーザー辞書
    audio/              WAV 読み書き
    history/            履歴保存（WAV + JSON）
    settings/           設定管理
  platform/             OS 固有アダプタ（Interface + darwin 実装）
    recorder/           PortAudio 録音（16kHz mono）
    clipboard/          テキスト読み書き + Cmd+V シミュレーション
    hotkey/             グローバルホットキー
    sound/              効果音（afplay）
    overlay/            オーバーレイ表示
    tray/               メニューバーアイコン
  app/                  オーケストレーター
assets/                 埋め込みアイコン
```

### 処理フロー

```
ホットキー押下 → 録音開始（WAV 16kHz mono）
ホットキー再押下 → 録音停止 → Gemini API 文字起こし
→ クリップボードにセット → Cmd+V ペースト → 履歴保存
```

## テスト

```bash
# ユニットテスト
go test ./...

# 統合テスト（実 API 使用）
go test -tags=integration ./internal/core/transcriber/ -v
```

## 依存ライブラリ

| ライブラリ | 用途 |
|-----------|------|
| `google.golang.org/genai` | Gemini API クライアント |
| `fyne.io/systray` | システムトレイ |
| `golang.design/x/hotkey` | グローバルホットキー |
| `golang.design/x/clipboard` | クリップボード操作 |
| `github.com/gordonklaus/portaudio` | 録音（PortAudio バインディング） |

## ライセンス

MIT
