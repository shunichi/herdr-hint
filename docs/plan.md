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

herdr（ターミナルマルチプレクサ）用プラグイン。Vimium 風の hint ラベルを **エージェント**に付け、
キー入力で任意のエージェントへジャンプする。本家 [maedana/herdr-hint](https://github.com/maedana/herdr-hint) を
ベースに `shunichi` 用に作る。

### オリジナルとの差分（本プロジェクトの仕様）

- **選択対象はエージェントのみ**（本家はタブ＋エージェント。本プロジェクトはタブをジャンプ対象から外す）。
- **各エージェントに種別（claude / codex など）を表示する**（本家に無い追加項目）。
- 上記以外の表示（リポジトリ / ブランチ / ステータス / ターミナルタイトル）とワークスペース別
  グルーピングは本家踏襲。

## 2. アーキテクチャ（要約）

**実装スタック: Go + Bubble Tea**（本家は Rust + crossterm だが、当プロジェクトは Go + Bubble Tea を採用。
Why は progress.md 参照）。herdr とは専用 SDK ではなく **CLI(JSON) 経由**で通信するため、言語非依存で移植可能。

想定コンポーネント（本家 lib.rs/main.rs の関数群を Go に翻案）:

- **herdr CLI クライアント**: `herdr workspace list` / `agent list` を JSON 取得しパース（タブは対象外だが、
  ワークスペース情報は表示グルーピングに使う）。操作は `herdr agent focus <id>`。
  herdr バイナリは env `HERDR_BIN_PATH`、無ければ `herdr`。
- **ラベル割当**: **エージェント**に a-z / aa-zz を割当。26 件以下は 1 文字、27 件以上は 2 文字（動的）。
- **Bubble Tea model（TUI）**: ワークスペース別にエージェントをグルーピング表示。キー入力ループで
  Ctrl+D/Ctrl+U スクロール、ラベル入力で対象解決、Esc キャンセル。
- **エージェント行の表示項目**: 種別（claude / codex：追加項目）／ repository ／ branch ／ status ／
  terminal title。repository・branch は cwd から git で算出（本家 git_context 相当）、
  status は `agent_status`、種別は `agent` フィールド、title は `terminal_title_stripped`。

ディレクトリ構成（予定）:

```
main.go              # エントリポイント（Bubble Tea program 起動）
internal/            # herdr クライアント・ラベル割当・model（分割は着手時に確定）
herdr-plugin.toml    # プラグイン定義（id を shunichi.hint 等に変更）
go.mod / go.sum
```

## 3. 確定した事実 / 前提

- 実装スタックは Go + Bubble Tea（progress.md 決定ログ参照）。
- **ジャンプ対象はエージェントのみ**（タブは対象外）。操作は `herdr agent focus <id>`。
  ワークスペース別グルーピングのため `herdr workspace list` は取得する。
- **エージェント行の表示項目**: 種別（追加）／ repository ／ branch ／ status ／ terminal title。
  - 種別 = `agent` フィールド（`claude`/`codex` 等）、status = `agent_status`、
    terminal title = `terminal_title_stripped`（codex は空になり得る）。
  - repository / branch = cwd から git で算出（本家 git_context 相当）。
- このマシンは `HERDR_ENV=1`（herdr 内で動作中）なので、`herdr ... list` の実 JSON スキーマを実機で確認できる。
- ラベル対象（エージェント）の focus には `herdr agent focus <target>`。target は terminal id / 一意な
  agent 名 / agent ラベル等を受け付ける（識別に使うフィールドは実機確認後に確定）。
- プラグイン定義 `herdr-plugin.toml` の `id` は本家 `maedana.hint` から `shunichi.hint` 等へ変更する。

### 3.1 本家実装の要点（Go 翻案の元。scratchpad に upstream-{lib,main}.rs 取得済み）

- **データ取得**: `herdr workspace list`（`result.workspaces[]`: `workspace_id` / `label` …）、
  `herdr agent list`（`result.agents[]`: `terminal_id` / `name` / `agent` / `agent_status` / `cwd` /
  `focused` / `terminal_title_stripped` / `workspace_id` …）。本家はさらに `tab list` も取るが本プロジェクトは不要。
- **agent → 表示item への変換**（本家 parse_agents）: `target_id = terminal_id`（focus に使う）、
  `display_name = name || agent || terminal_id`、`status = agent_status`、
  `context = git_context(cwd)`（`"repo:branch"`）、`title = terminal_title_stripped`。
- **git_context**（本家）: `git -C <cwd> rev-parse --show-toplevel` / `--git-common-dir` /
  `--abbrev-ref HEAD` を実行。repo 名は common-dir が絶対パスならその親ディレクトリ名（worktree 対応）、
  無ければ toplevel の basename。branch は abbrev-ref HEAD。→ `"repo:branch"`。
- **ラベル割当**（本家 assign_labels）: 全item数が 26 以下なら 1 文字（a,b,…）、27 以上なら 2 文字
  （aa,ab,…）。`resolve_input` はラベル完全一致で item を返す。`uses_double_labels` は先頭itemのラベル長で判定。
- **本家と変える点（本プロジェクト固有）**:
  1. **タブを扱わない**（本家は tab+agent。2 カラムのタブ表示ロジックは不要）。
  2. **種別（agent フィールド）を行に表示する**（本家は agent 行に種別を出していない）。
  3. **エージェントを workspace 別にグループ表示する**（本家の agent 行はフラットな "Agents" 一覧。
     グループ化には agent の `workspace_id` と workspace の `label` を使う）。
- **キー操作**（本家 main）: ラベル文字入力で解決（2 文字モードは 2 文字読む）、Ctrl+D/Ctrl+U で
  半画面スクロール、Esc でキャンセル。選択後 raw mode を戻して `herdr agent focus <terminal_id>`。

## 4. タスク分割

調査（本家ソース取得・実機 JSON スキーマ確認）は完了し §3・§3.1 に反映済み。実装は agent-tasks で管理する
（登録後に ID を採番して追記）。依存はほぼ直列。

| ID | タイトル | 依存 |
|----|----------|------|
| 0001 | Go プロジェクト初期化・骨組み・ビルド基盤（go.mod / main スケルトン / herdr-plugin.toml / Bubble Tea 導入） | — |
| 0002 | herdr クライアント・データモデル・git_context（workspace/agent list 取得・パース・focus・repo:branch 算出）＋テスト | 0001 |
| 0003 | ラベル割当・入力解決の純ロジック（エージェント対象・動的 1/2 文字・resolve）＋テスト | 0001 |
| 0004 | Bubble Tea TUI（ワークスペース別グループ表示・種別/status/context/title 行・キー操作・スクロール・focus 連携） | 0002, 0003 |
| 0005 | 実機動作確認・README・仕上げ（ローカル herdr へ link して動作確認、空状態/スクロール境界の調整） | 0004 |

## 5. 未決事項（実装時に決める）

- `internal/` のパッケージ分割粒度。
- エージェントを herdr の JSON からどのフィールドで一意識別し、`agent focus` の target に何を渡すか（実機確認後に確定）。
- repository / branch の算出方法（cwd で git を呼ぶ／worktree の扱い）。
- 種別（claude/codex 等）の取りうる値の一覧と、未知種別・空タイトルの表示フォールバック。
- ワークスペース内でのエージェント表示順、およびエージェント 0 件のワークスペースの扱い。
- Bubble Tea でのポップアップ描画と herdr の pane placement(`popup`) の兼ね合い。
- テスト戦略（ラベル割当・入力解決は純ロジックとして単体テストしやすい）。
