# デザイン参考（Phase 3）

更新日: 2026-02-07
方針: 独自デザインを最小化し、各プラットフォームのデフォルトUIを優先する。

## ユーザー要望の整理
- 画面は「メニューバー/トレイ + 設定画面」が中心。
- 可能な限り各OSの標準デザインに合わせたい。
- ブランド演出よりも、違和感のないネイティブ体験を重視。

## 参考サイト（公式）
1. Apple Human Interface Guidelines: Designing for macOS
   - https://developer.apple.com/design/human-interface-guidelines/designing-for-macos
2. Apple AppKit: NSStatusBar
   - https://developer.apple.com/documentation/appkit/nsstatusbar
3. Microsoft Learn: Notification Area
   - https://learn.microsoft.com/en-us/windows/win32/uxguide/winenv-notification
4. GNOME HIG
   - https://developer.gnome.org/hig/
5. Android Developers: Create an input method
   - https://developer.android.com/develop/ui/views/touch-and-input/creating-input-method
6. Apple Docs: Custom Keyboard / Open Access
   - https://developer.apple.com/library/archive/documentation/General/Conceptual/ExtensibilityPG/CustomKeyboard.html
   - https://developer.apple.com/library/archive/documentation/General/Conceptual/ExtensibilityPG/CustomKeyboard.html#//apple_ref/doc/uid/TP40014214-CH16-SW19

## サイト分析（Web調査）

### 1) macOS（Apple HIG + NSStatusBar）
- レイアウト構成:
  - 常駐機能はメニューバーExtra中心、詳細設定は別ウィンドウに分離。
- カラースキーム:
  - システムの Light/Dark とアクセントカラーに追従。
- タイポグラフィ:
  - SF（San Francisco）を前提に、標準コントロールで可読性を担保。
- インタラクション:
  - クリックでメニュー表示、状態変化はアイコンで明確化。
- ナビゲーション:
  - 階層を浅くし、設定はカテゴリ分割より単純なフォーム優先。

### 2) Windows（Notification Area）
- レイアウト構成:
  - 常駐は通知領域アイコン + コンテキストメニュー。
- カラースキーム:
  - システムテーマ追従（ライト/ダーク/ハイコントラスト）。
- タイポグラフィ:
  - システムフォント（Segoe UI）と標準間隔に合わせる。
- インタラクション:
  - 左クリック/右クリック動作を明確に分離し、誤操作を防ぐ。
- ナビゲーション:
  - 設定画面は「主要設定を1画面」で完結させる。
- 注記:
  - Notification Areaページは旧版ガイド表記があるため、最終実装時は最新の Fluent/WinUI ガイドとの整合チェックが必要。

### 3) Linux（GNOME HIG）
- レイアウト構成:
  - シンプルな設定画面（余白広め、1カラム）を推奨。
- カラースキーム:
  - テーマ追従（ディストリビューション差分を吸収）。
- タイポグラフィ:
  - システム標準フォントを使用し、独自指定を最小化。
- インタラクション:
  - 直接的で予測可能な操作（トグル・保存）を優先。
- ナビゲーション:
  - 階層を増やさず、主要設定をトップレベルに置く。

### 4) モバイル将来対応（Android/iOS）
- レイアウト構成:
  - 全画面アプリではなく、IME/キーボード拡張に寄せる設計が前提。
- カラースキーム:
  - Material（Android）/Human Interface（iOS）の標準テーマ追従。
- タイポグラフィ:
  - Androidはシステムフォント、iOSはSF準拠。
- インタラクション:
  - 音声入力開始/停止の明確な状態表示が必須。
- ナビゲーション:
  - 「キーボード設定」画面と「権限説明」画面を最小構成で提供。

## デザイン決定（現時点）
- デスクトップ:
  - Native-first（OS標準コントロールのみでMVPを成立させる）
  - メニューバー/トレイはOSごとの慣習に合わせる
  - 独自装飾・独自フォントは採用しない
- モバイル:
  - Desktop MVP後に入力方式（IME/キーボード拡張）を検証して決定
