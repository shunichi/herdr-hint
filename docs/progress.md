# herdr-hint — 進捗管理・決定ログ（Living Document）

> タスク（worktree）をまたいで参照する進捗と決定の記録。計画そのものは [`docs/plan.md`](./plan.md)。
>
> - **進捗サマリ**: 各タスクの現在状態（正は agent-tasks。ここは俯瞰用のスナップショット）。
> - **決定ログ**: 実装中に確定した設計判断（Why 付き。未決は plan.md §5 に残す）。
> - **次アクション**: 直近やること。

最終更新: 2026-07-22

---

## 進捗サマリ

正確な状態は `agent-tasks --project herdr-hint` で確認。ここはマイルストーン俯瞰。

| ID | タイトル | 状態 | メモ |
|----|----------|------|------|
| —  | プロジェクト初期セットアップ（docs/・CLAUDE.md・AGENTS.md・agmsg） | 完了 | agent-project-init skill で実施 |
| —  | 調査（本家ソース＋実機 JSON）・仕様確定・タスク登録 | 完了 | §3.1/§3.2・0001-0005 |
| —  | 実装前レビュー（codex） | 完了 | 指摘反映 → 承認 |
| 0001-0005 | 実装（Go + Bubble Tea・hint） | **完了** | 全 PR (#1-#5) codex 承認 → マージ。build/test/vet green |
| 0006 | filter コマンド追加（別 entrypoint） | **完了** | PR #6 マージ（ユーザー指示。codex 未返信） |
| 0007 | herdr 0.7.5 対応（focus を pane_id に） | 実装済み・レビュー中 | PR #7。実機 0.7.5 で focus 成功確認 |

## 決定ログ

実装中に確定した設計判断を、Why 付きで新しい順に追記する。

- **2026-07-22 herdr 0.7.5 対応: focus を terminal_id → pane_id に（0007）**: herdr 0.7.5 で
  `agent focus` の target が terminal_id → pane_id に変更（terminal_id は受け付け不可）。確認は 3 点:
  (1) `agent` help「targets accept unique agent names and pane ids」、(2) 実測 `agent get <terminal_id>`→
  agent_not_found / `<pane_id>`→OK、(3) CHANGELOG 0.7.5「accept only a unique live agent name or the pane ID」。
  `internal/herdr` の TargetID を pane_id に（hint/filter とも TargetID 経由で両対応）、`min_herdr_version`
  を 0.7.5 に。Why: 0.7.5 で現行 focus が破綻するため必須。実機 0.7.5 で focus 成功（EXIT=0）確認。
- **2026-07-22 filter コマンドを別コマンドとして追加（0006）**: hint（ラベル方式）は残し、プロジェクト名を
  fuzzy 絞り込みしてカーソル選択する filter を、同一プラグインの別 entrypoint として追加（案 A1・単一バイナリ
  + `herdr-hint filter`）。grouping は `internal/ui.Arrange` を再利用し hint 用コードは不変。Why: ラベルを
  目視で探すのがやりづらいため（ユーザー要望）。PR #6 はユーザー指示でマージ（codex レビューは未返信）。

- **2026-07-19 実装完了（全 5 タスク）**: 0001-0005 を各 PR (#1-#5) で実装し、codex レビュー承認 →
  squash マージ。最終タスクはユーザー承認も取得。main で 4 パッケージすべて build/test/vet green、
  実 TTY での E2E（描画→ラベル→focus）と plugin install を確認済み。herdr-hint 完成。

- **2026-07-18 実装前レビュー(codex)承認**: 修正版 docs(commit aeb140c)とタスク 0001-0005 を codex が承認。
  実装着手 OK。軽微な確認事項（再レビュー不要）を反映済み: (a) 決定的ソートは全順序のタイブレーク
  （workspace number→workspace_id、terminal_id→元位置）を 0004 で入れる、(b) 2 文字目不一致時はバッファ
  クリアして次の 1 文字を先頭入力に、(c) 677 件は実機困難なら fixture/自動テストで代替可(0005)。
- **2026-07-18 実装前レビュー(codex)を反映**: codex レビューで docs の確定/未決の矛盾・仕様の不足を
  指摘され修正。(1) §3.1 と §3/§5/次アクションの確定状態を同期、(2) 「ワークスペース別グループは本家踏襲」の
  誤りを訂正（本家 agent はフラット＝本プロジェクト固有差分）、(3) 表示順・ラベル上限・入力状態・focus
  タイミング・フォールバック表示を §3.2 に確定。Why: 実装前に受入条件を明確化し手戻りを防ぐため。
- **2026-07-18 調査完了（本家ソース＋実機 JSON）**: 本家 `src/lib.rs`(544行)/`src/main.rs`(113行) を
  取得・精読（scratchpad に upstream-{lib,main}.rs）。実機で JSON スキーマ確認済み:
  `herdr workspace list` → `result.workspaces[]`（`workspace_id`/`label`/`number`/`focused`/… ）、
  `herdr agent list` → `result.agents[]`（`terminal_id`/`agent`/`agent_status`/`cwd`/`focused`/
  `terminal_title_stripped`/`workspace_id`/`pane_id`/…。`agent` は種別=claude/codex）。focus は
  `herdr agent focus <terminal_id>`（コマンド仕様は確定、実 focus は他 pane を奪うため未実行→0004/0005 で確認）。
- **2026-07-18 仕様: 選択対象はエージェントのみ＋種別表示＋workspace 別グループ**: 本家（タブ＋エージェント・
  agent はフラット・種別非表示）に対し、本プロジェクトは (1) ジャンプ対象をエージェントのみに、
  (2) 各行に種別（claude/codex 等）を追加表示、(3) エージェントを workspace 別にグループ表示、の 3 点を変える。
  表示項目 repository/branch/status/terminal title と、グループ化の見た目自体は本家のタブ表示に倣う。
  terminal title は `terminal_title_stripped`（codex は空になり得る→フォールバックは §3.2）。
  Why: ユーザー要望（タブへのジャンプ不要、種別が分かると識別しやすい、workspace 単位で見たい）。
- **2026-07-18 相手役の spawn は herdr を使う**: レビュー役(codex)などを spawn するときは agmsg の
  tmux spawn ではなく herdr のネイティブ pane（`herdr pane split` + `pane run`）を使う。agmsg は
  チーム登録(join)と受信のみ。Why: HERDR_ENV=1 環境では herdr の pane 管理と統合させるため。
- **2026-07-18 役割構成**: 実装役 = claude、レビュー役 = codex（agmsg team `herdr-hint`）。
  Why: agent-project-init skill の既定構成。状況に応じて入れ替え／両役 Claude も可。
- **2026-07-18 実装スタックを Go + Bubble Tea に決定**: 本家 herdr-hint は Rust + crossterm だが、
  当プロジェクトは Go + Bubble Tea を採用。Why: プロジェクト共通の技術選定既定（TUI は Go + Bubble Tea）に従う。
  herdr とは CLI(JSON) 経由通信で言語非依存のため、Rust 実装を移植可能。

## 次アクション

- [x] 本家 `src/lib.rs` / `src/main.rs` の実コードを取得して仕様を確定（§3.1 に反映）
- [x] 実機で `herdr workspace/agent list` の JSON スキーマを確認（上記決定ログ）
- [x] 実装タスク 0001-0005 を agent-tasks に登録
- [x] 実装前レビュー(codex)を受け、docs を修正（§3.2 追加・矛盾解消）
- [x] 修正版 docs を codex に再レビュー → 承認
- [x] 0001 → 0002 → 0003 → 0004 → 0005 を実装・codex 承認・マージ（全完了）
