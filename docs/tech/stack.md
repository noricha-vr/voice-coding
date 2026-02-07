# 技術選定（Phase 4）

更新日: 2026-02-07
決定: **A. Go中心（Desktop MVPを先に完成）**

## 選定方針
- 最優先は `精度 > 貼り付け成功率 > 速度`
- デスクトップ（macOS/Windows/Linux）を先に統一体験で完成させる
- モバイル（iOS/Android）は Desktop MVP達成後に検証・分離実装
- UIは Native-first（OS標準UIを優先）

## フロントエンド（デスクトップUI）
- 基本方針:
  - 常駐UI: メニューバー/トレイ（ネイティブ）
  - 設定UI: OS標準コントロール中心
- 実装方式:
  - 共通ロジックは Go
  - UIとOS連携は Platform Adapter で分離（`darwin/windows/linux`）
- 補足:
  - Flutter は今回は採用しない（OS低レイヤー連携が主課題のため）

## バックエンド
- 構成: なし（ローカル実行アプリ）
- API方式:
  - 音声→文字起こしは Gemini API へ直接リクエスト
  - REST/GraphQL サーバーはMVPでは持たない

## データストア
- MVP:
  - ローカルファイル（設定JSON、履歴JSON、音声WAV）
- 将来候補:
  - 履歴検索/統計が必要になったら SQLite を検討

## インフラ
- アプリ配布:
  - macOS: `.app` / Homebrew
  - Windows: `msi` または `exe` インストーラ
  - Linux: `deb` / `rpm` / `AppImage` のいずれか
- 実行基盤:
  - クラウド常駐は不要（クライアントアプリ）

## その他サービス
- 認証: なし（ローカル利用）
- 決済: なし（MVP時点）
- メール: なし（MVP時点）
- 監視:
  - ローカルログ（必須）
  - 将来、クラッシュ収集が必要なら Sentry 等を検討

## Go採用理由
1. 既存Python実装の責務（録音/文字起こし/貼り付け/常駐）を小さく分割して移植しやすい
2. 単一バイナリ配布がしやすく、環境依存を減らせる
3. Desktop先行戦略に対して、OS連携を明示的に実装しやすい

## Flutter非採用（現時点）の理由
1. 今回の主課題はUI共通化より、ホットキー/クリップボード/貼り付けなどOS連携の安定化
2. Flutterでも最終的にプラットフォーム別実装が必要
3. Native-first要件（OSデフォルト見た目）との整合が取りにくい

## モバイル（後段）方針
- iOS/Androidは「全アプリ横断の入力体験」を再現するため、IME/キーボード拡張前提で別設計
- Desktopと同一コードベースに無理に統合せず、コアの音声処理部分のみ共有を検討

## リスクと対策
- リスク: OSごとの差分実装が増える
  - 対策: Core/Adapter分離、共通インターフェース化
- リスク: 貼り付け操作の失敗率
  - 対策: 貼り付け結果検知・リトライ・Fail Loudログ
- リスク: モバイル方式の不確実性
  - 対策: Desktop MVP達成後にPoCフェーズを独立実施

## 未確定事項
- Go UIライブラリ/OS APIバインディングの具体候補名（PoCで確定）

## 合意済みゲート
- GoのOS連携ライブラリ最終決定は、主要3機能（ホットキー/トレイ/貼り付け）のPoC結果で行う。

## 調査参照（公式）
- Go release policy / release history: https://go.dev/doc/devel/release
- Gemini API quickstart (Go SDK): https://ai.google.dev/gemini-api/docs/quickstart
- Gemini API libraries (Go SDK配布): https://ai.google.dev/gemini-api/docs/libraries
- macOS NSStatusBar: https://developer.apple.com/documentation/AppKit/NSStatusBar
- Windows Notification Area: https://learn.microsoft.com/en-us/windows/win32/uxguide/winenv-notification
- Android IME実装: https://developer.android.com/develop/ui/views/touch-and-input/creating-input-method
- iOS Custom Keyboard: https://developer.apple.com/library/archive/documentation/General/Conceptual/ExtensibilityPG/CustomKeyboard.html
