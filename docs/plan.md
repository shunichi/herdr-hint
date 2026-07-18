# herdr-hint — 実装計画（Living Document）

> タスク（worktree）をまたいで参照する**生きた計画書**。方針・アーキテクチャ・タスク分割・
> 未決事項をここで管理する。決定の履歴と日々の進捗は [`docs/progress.md`](./progress.md) に記録する。
> 作成時点の設計スナップショット（[`../HANDOFF.md`](../HANDOFF.md)：本家 README/トップページの調査メモ）は
> これとは分けて扱う。
>
> **役割分担**: スナップショット([`../HANDOFF.md`](../HANDOFF.md))=作成時点の記録（不変）／
> plan.md=最新の計画（更新する）／ progress.md=決定ログと進捗（追記する）。

最終更新: 2026-07-18

---

## 1. 目的（要約）

herdr（ターミナルマルチプレクサ）用プラグイン。Vimium 風の hint ラベルを全タブ・全エージェントに付け、
キー入力で任意の対象へジャンプする。本家 [maedana/herdr-hint](https://github.com/maedana/herdr-hint) 相当を
`shunichi` 用に作る。

## 2. アーキテクチャ（要約）

**実装スタック: Go + Bubble Tea**（本家は Rust + crossterm だが、当プロジェクトは Go + Bubble Tea を採用。
Why は progress.md 参照）。herdr とは専用 SDK ではなく **CLI(JSON) 経由**で通信するため、言語非依存で移植可能。

想定コンポーネント（本家 lib.rs/main.rs の関数群を Go に翻案）:

- **herdr CLI クライアント**: `herdr workspace list` / `tab list` / `agent list` を JSON 取得しパース。
  操作は `herdr tab focus <id>` / `herdr agent focus <id>`。herdr バイナリは env `HERDR_BIN_PATH`、無ければ `herdr`。
- **ラベル割当**: タブ＋エージェントに a-z / aa-zz を割当。26 件以下は 1 文字、27 件以上は 2 文字（動的）。
- **Bubble Tea model（TUI）**: ワークスペース別 2 カラム表示。キー入力ループで
  Ctrl+D/Ctrl+U スクロール、ラベル入力で対象解決、Esc キャンセル。
- **git context**: エージェント行に repository / branch / status / terminal title を表示。

ディレクトリ構成（予定）:

```
main.go              # エントリポイント（Bubble Tea program 起動）
internal/            # herdr クライアント・ラベル割当・model（分割は着手時に確定）
herdr-plugin.toml    # プラグイン定義（id を shunichi.hint 等に変更）
go.mod / go.sum
```

## 3. 確定した事実 / 前提

- 実装スタックは Go + Bubble Tea（[[decision-2026-07-18-tech-stack]] 相当。progress.md 決定ログ参照）。
- herdr との通信は CLI(JSON) 経由。取得: `workspace/tab/agent list`。操作: `tab/agent focus <id>`。
- このマシンは `HERDR_ENV=1`（herdr 内で動作中）なので、`herdr ... list` の実 JSON スキーマを実機で確認できる。
- `herdr agent list` の各要素の `agent` フィールドは**種別**（`claude`/`codex`）。ラベル対象の識別には
  `pane_id` / `terminal_title(_stripped)` / `cwd` 等を使う（HANDOFF.md 段階では未確認、実機確認が必要）。
- プラグイン定義 `herdr-plugin.toml` の `id` は本家 `maedana.hint` から `shunichi.hint` 等へ変更する。

## 4. タスク分割

（agent-tasks で管理。着手時に ID を採番して追記する）

| ID | タイトル | 依存 |
|----|----------|------|
|    | 本家 src/lib.rs・src/main.rs の実コードを取得し仕様確定 | — |
|    | 実機で `herdr workspace/tab/agent list` の JSON スキーマを確認 | — |
|    | Go プロジェクト初期化（go.mod・Bubble Tea 導入） | — |
|    | herdr CLI クライアント（list 取得・focus 操作） | 上記 |
|    | ラベル割当ロジック（1/2 文字・動的） | — |
|    | Bubble Tea model（2 カラム表示・キー入力・スクロール・解決） | 上記 |
|    | herdr-plugin.toml（id・build・panes）を Go 向けに作成 | — |

## 5. 未決事項（実装時に決める）

- `internal/` のパッケージ分割粒度。
- ラベル対象（タブ／エージェント）を herdr の JSON からどのフィールドで一意識別するか（実機確認後に確定）。
- Bubble Tea でのポップアップ描画と herdr の pane placement(`popup`) の兼ね合い。
- テスト戦略（ラベル割当・入力解決は純ロジックとして単体テストしやすい）。
