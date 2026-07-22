# herdr-hint — 実装計画（Living Document）

> タスク（worktree）をまたいで参照する**生きた計画書**。方針・アーキテクチャ・タスク分割・
> 未決事項をここで管理する。決定の履歴と日々の進捗は [`docs/progress.md`](./progress.md) に記録する。
> 作成時点の設計スナップショット（[`../HANDOFF.md`](../HANDOFF.md)：本家 README/トップページの調査メモ）は
> これとは分けて扱う。
>
> **役割分担**: スナップショット([`../HANDOFF.md`](../HANDOFF.md))=作成時点の記録（不変）／
> plan.md=最新の計画（更新する）／ progress.md=決定ログと進捗（追記する）。

最終更新: 2026-07-22

---

## 1. 目的（要約）

herdr（ターミナルマルチプレクサ）用プラグイン。Vimium 風の hint ラベルを **エージェント**に付け、
キー入力で任意のエージェントへジャンプする。本家 [maedana/herdr-hint](https://github.com/maedana/herdr-hint) を
ベースに `shunichi` 用に作る。

### オリジナルとの差分（本プロジェクトの仕様）

本プロジェクト固有の差分は次の 3 点（本家 = maedana/herdr-hint に対して）:

1. **選択対象はエージェントのみ**（本家はタブ＋エージェント。本プロジェクトはタブをジャンプ対象から外す）。
2. **各エージェントに種別（claude / codex など）を表示する**（本家の agent 行に種別表示は無い）。
3. **エージェントをワークスペース別にグループ表示する**（本家の agent 一覧はフラット。本家がグループ化
   しているのはタブのみ。エージェントの workspace 別グループは本プロジェクト固有）。

表示項目のうち リポジトリ / ブランチ / ステータス / ターミナルタイトル は本家踏襲（本家 agent 行と同じ）。
種別の追加（差分 2）とグループ化（差分 3）だけが本家と異なる。

## 2. アーキテクチャ（要約）

**実装スタック: Go + Bubble Tea**（本家は Rust + crossterm だが、当プロジェクトは Go + Bubble Tea を採用。
Why は progress.md 参照）。herdr とは専用 SDK ではなく **CLI(JSON) 経由**で通信するため、言語非依存で移植可能。

**2 コマンド構成**（単一バイナリのサブコマンド。0006）: `herdr-hint`（hint = ラベル方式、entrypoint `jump`）と
`herdr-hint filter`（filter = プロジェクト名 fuzzy 絞り込み＋カーソル選択、entrypoint `filter`）。
`internal/herdr`（クライアント・focus・git_context）と grouping（`internal/ui.Arrange`）を共有する。
以下は hint 側の設計（filter も同じデータ・focus 経路を使う）。

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
- **実機の JSON スキーマは確認済み**（`HERDR_ENV=1`。`workspace list` / `agent list` の実出力を取得し
  §3.1 のフィールドを確定。代表 JSON と確認コマンドは progress.md 決定ログに記録）。
- **focus の識別子は `pane_id`**（herdr 0.7.5 対応・0007）。`herdr agent focus <pane_id>` を使う。
  - herdr 0.7.5 で `agent focus` の target が terminal_id → pane_id に変更され、terminal_id は受け付け
    られなくなった（help / 実測 `agent get` / CHANGELOG の 3 点で確認。progress.md 決定ログ参照）。
    このため `min_herdr_version` を 0.7.5 に設定する。
  - 実機 0.7.5 で filter→Enter→focus 成功（EXIT=0）を確認済み（0007）。
- プラグイン定義 `herdr-plugin.toml` の `id` は本家 `maedana.hint` から `shunichi.hint` 等へ変更する。
  `min_herdr_version` は 0.7.5。
- 表示順・ラベル上限・入力状態・focus タイミング・フォールバック表示は §3.2 に確定（テスト対象）。

### 3.1 本家実装の要点（Go 翻案の元。scratchpad に upstream-{lib,main}.rs 取得済み）

- **データ取得**: `herdr workspace list`（`result.workspaces[]`: `workspace_id` / `label` …）、
  `herdr agent list`（`result.agents[]`: `pane_id` / `name` / `agent` / `agent_status` / `cwd` /
  `focused` / `terminal_title_stripped` / `workspace_id` …。`terminal_id` も返るが 0.7.5 で focus には
  使えない）。本家はさらに `tab list` も取るが本プロジェクトは不要。
- **agent → 表示item への変換**（本家 parse_agents を Go 翻案。0007 で pane_id に更新）:
  `target_id = pane_id`（focus に使う）、`display_name = name || agent || pane_id`、`status = agent_status`、
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
  半画面スクロール、Esc でキャンセル。選択後 raw mode を戻して `herdr agent focus <pane_id>`（0007）。

### 3.2 確定仕様（実装前に決めた挙動。テスト対象）

本家に無い/曖昧な点をレビュー指摘を受けて確定する。以下は実装の受入条件かつテスト対象。

- **表示順（決定的にする。JSON 順に依存しない）**:
  - ワークスペースは `herdr workspace list` の `number` 昇順。
  - ワークスペース内のエージェントは `target_id`（= `pane_id`）の昇順（安定・決定的な全順序。0007 で pane_id に）。
  - **orphan エージェント**（`workspace_id` が workspace list に無い / 空）は、既知ワークスペース群の後ろに
    `?` グループとしてまとめて表示。
  - **エージェント 0 件のワークスペースは表示しない**（見出しだけの空グループを出さない）。
  - ラベルはこの確定順で先頭から割り当てる（順が決定的なので毎回同じラベルになる）。
- **ラベルの上限と入力状態**:
  - 割当は総数で決まる: `n<=26` → 1 文字（a..z）、`27<=n<=676` → 2 文字（aa..zz）、
    **`n>=677` は表現不能**。677 以上は先頭 676 件のみラベル付与し、超過分は「ラベル無し（選択不可）」として
    その旨を画面下部に明示する（クラッシュさせない）。
  - 入力状態の挙動: **Esc / Ctrl+C** はキャンセルして何もせず終了。**Backspace** は 2 文字モードで 1 文字目入力後に
    取り消せる。**a–z 以外**のキーは無視。**2 文字モードで 1 文字目が候補の接頭辞に無い**場合は即無効入力として
    無視（確定させない）。**resize** は再描画する。
  - テスト対象件数: 0 / 26 / 27 / 676 / 677。
- **focus の実行タイミングと失敗処理**:
  - Bubble Tea の model は「選択された target(`pane_id`)」を返すだけにし、**program 終了・端末復元後**に
    main が `herdr agent focus <pane_id>` を実行する（raw mode 中に外部コマンドを叩かない。本家踏襲）。
  - focus が失敗したら **stderr にメッセージを出し、非 0 exit code** で終える（無言で失敗しない）。
- **フォールバック表示（「情報欠落」と「空文字」を区別）**:
  - 種別（agent）: 値があれば生値を表示、欠落/空なら `unknown`。
  - ターミナルタイトル: 空/欠落なら代替名（display_name）または em dash `—`。
  - git context（repo:branch）: 算出不能なら `—`。

## 4. タスク分割

調査（本家ソース取得・実機 JSON スキーマ確認）は完了し §3・§3.1 に反映済み。実装は agent-tasks で管理する。
0002 と 0003 は 0001 完了後に並行可能。各タスクは**自己完結した PR** になるよう受入条件を持つ（0001 は
ビルド/テスト成功のスケルトン、0002/0003 は未接続でも公開契約＋テストが完結、0004 で初めて E2E 接続、
0005 で manifest/install/実機確認）。詳細な受入条件・テストケースは各 agent-tasks タスク本文に記載。

| ID | タイトル | 依存 |
|----|----------|------|
| 0001 | Go プロジェクト初期化・骨組み・ビルド基盤（go.mod / main スケルトン / herdr-plugin.toml / Bubble Tea 導入） | — |
| 0002 | herdr クライアント・データモデル・git_context（workspace/agent list 取得・パース・focus・repo:branch 算出）＋テスト | 0001 |
| 0003 | ラベル割当・入力解決の純ロジック（エージェント対象・動的 1/2 文字・resolve）＋テスト | 0001 |
| 0004 | Bubble Tea TUI（ワークスペース別グループ表示・種別/status/context/title 行・キー操作・スクロール・focus 連携） | 0002, 0003 |
| 0005 | 実機動作確認・README・仕上げ（ローカル herdr へ link して動作確認、空状態/スクロール境界の調整） | 0004 |
| 0006 | インクリメンタル絞り込み式ジャンプを別コマンドとして追加（`herdr-hint filter`。fuzzy・グループ維持・↑↓/C-n/C-p・Enter） | 0002-0004 |
| 0007 | herdr 0.7.5 対応（`agent focus` の target を terminal_id → pane_id に） | 0002 |

0001-0006 は実装・レビュー・マージ済み。0007 はレビュー中（進捗は progress.md）。

## 5. 未決事項（実装時に決める）

> 解決済み（§3.1 / §3.2 へ移動）: 識別子＝pane_id（0007）/ git_context 方式 / 種別・タイトル・git の
> フォールバック / 表示順・0 件 workspace / ラベル上限・入力状態 / focus タイミング / 実機 focus 成功確認（0007）。

- （現時点で未決事項なし。新たな herdr 仕様変更や要望が出たらここに追記する。）
