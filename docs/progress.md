# herdr-hint — 進捗管理・決定ログ（Living Document）

> タスク（worktree）をまたいで参照する進捗と決定の記録。計画そのものは [`docs/plan.md`](./plan.md)。
>
> - **進捗サマリ**: 各タスクの現在状態（正は agent-tasks。ここは俯瞰用のスナップショット）。
> - **決定ログ**: 実装中に確定した設計判断（Why 付き。未決は plan.md §5 に残す）。
> - **次アクション**: 直近やること。

最終更新: 2026-07-18

---

## 進捗サマリ

正確な状態は `agent-tasks --project herdr-hint` で確認。ここはマイルストーン俯瞰。

| ID | タイトル | 状態 | メモ |
|----|----------|------|------|
| —  | プロジェクト初期セットアップ（docs/・CLAUDE.md・AGENTS.md・agmsg） | 完了 | agent-project-init skill で実施 |
| —  | 実装（Go + Bubble Tea） | 未着手 | plan.md §4 のタスク分割参照 |

## 決定ログ

実装中に確定した設計判断を、Why 付きで新しい順に追記する。

- **2026-07-18 役割構成**: 実装役 = claude、レビュー役 = codex（agmsg team `herdr-hint`）。
  Why: agent-project-init skill の既定構成。状況に応じて入れ替え／両役 Claude も可。
- **2026-07-18 実装スタックを Go + Bubble Tea に決定**: 本家 herdr-hint は Rust + crossterm だが、
  当プロジェクトは Go + Bubble Tea を採用。Why: プロジェクト共通の技術選定既定（TUI は Go + Bubble Tea）に従う。
  herdr とは CLI(JSON) 経由通信で言語非依存のため、Rust 実装を移植可能。

## 次アクション

- [ ] 本家 `src/lib.rs` / `src/main.rs` の実コードを取得して仕様を確定する
- [ ] 実機（`HERDR_ENV=1`）で `herdr workspace/tab/agent list` の JSON スキーマを確認する
- [ ] Go プロジェクト初期化（go.mod・Bubble Tea 導入）し、agent-tasks にタスクを採番して登録する
