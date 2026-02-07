# API仕様（Phase 5）

更新日: 2026-02-07
方針: MVPでは独自バックエンドAPIを持たず、クライアントから Gemini API を直接利用する。

## API構成
- 独自API: なし（REST/GraphQLサーバーは導入しない）
- 外部API: Google Gemini API（音声文字起こし）

## 外部エンドポイント一覧

### 1. 音声文字起こし
- エンドポイント: `POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent`
- 用途: 音声データを日本語テキストへ変換し、開発用語補正を適用した結果を取得
- 認証: APIキー（`GOOGLE_API_KEY`）

#### リクエスト概要
- `model`: 例 `gemini-2.5-flash`
- `contents`:
  - 文字起こし指示テキスト
  - 音声パート（WAV/PCM相当）
- `generationConfig` / `thinkingConfig`:
  - 応答品質・速度・推論量の調整

#### レスポンス概要
- `candidates[].content.parts[].text`:
  - 文字起こし結果テキスト（主利用）
- エラー:
  - 401/403: 認証・権限エラー
  - 429: レート制限
  - 5xx: 一時的API障害

## クライアント内インターフェース（参考）
- `startRecording()`: 録音開始
- `stopRecording()`: 録音停止と音声確定
- `transcribe(audio)`: 外部APIへ送信して文字起こし
- `copyAndPaste(text)`: クリップボード設定に従って貼り付け実行
- `saveHistory(meta)`: ローカル履歴保存

## 非機能要件（API関連）
- タイムアウト: 10秒前後を初期値（要調整）
- リトライ:
  - 一時エラー（429/5xx/timeout）のみ限定リトライ
  - 恒久エラー（認証不備など）は即失敗
- ログ:
  - API呼び出し時間、成功/失敗、エラー種別を記録
