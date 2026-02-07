# アーキテクチャ設計（Phase 7）

更新日: 2026-02-07
方針: Go中心、Core/Adapter分離、Desktop先行。

## システム構成図
```mermaid
flowchart LR
    U[ユーザー] --> HK[グローバルホットキー]
    HK --> REC[Audio Recorder]
    REC --> WAV[一時WAV]
    WAV --> TR[Transcriber]
    TR --> GEM[Gemini API]
    GEM --> TR
    TR --> PP[Text Postprocess]
    PP --> CP[Clipboard/Paste Adapter]
    CP --> APP[対象アプリ]
    PP --> HIS[History Store(JSON/WAV)]
    CORE[Go Core Logic] --- REC
    CORE --- TR
    CORE --- PP
    ADP[OS Platform Adapters] --- HK
    ADP --- CP
    ADP --- UI[Tray/Menu + Settings UI]
```

## コンポーネント間通信方式
- Core内部: 関数呼び出し + 明示的インターフェース
- Core ↔ OS連携: Platform Adapter（`darwin/windows/linux`）による同期呼び出し
- Core ↔ Gemini: HTTPS API（JSON + 音声データ）
- 永続化: ローカルファイルI/O（JSON/WAV）

## スケーラビリティ考慮
- 現状は単一ユーザー・単一プロセスで十分
- 将来拡張:
  - 非同期キューで録音/文字起こし処理を分離
  - 履歴をSQLiteへ移行して検索性能を確保
  - モデル切替戦略を設定化して負荷/コスト制御

## セキュリティ考慮
- APIキーは環境変数またはローカル設定に保存し、ログへ出さない
- 権限不足時は即失敗（Fail Fast）し、明示的にユーザーへ通知
- クリップボード操作は最小権限・最小時間で行い、復元設定を提供
- ログに音声データ本体を保存しない（履歴ON時のみ明示保存）
