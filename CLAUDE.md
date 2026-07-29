# herdr-hint — プロジェクトガイド（CLAUDE.md）

herdr（ターミナルマルチプレクサ）用プラグイン。Vimium 風の hint ラベルで任意のタブ／エージェントへ
ジャンプする。**実装スタックは Go + Bubble Tea**（herdr とは CLI(JSON) 経由で通信）。
作成時点の調査スナップショットは [`HANDOFF.md`](./HANDOFF.md)、最新の計画は [`docs/plan.md`](./docs/plan.md)。

## ドキュメント運用

- [`docs/plan.md`](./docs/plan.md) — 最新の実装計画（アーキテクチャ・確定事実・タスク分割・未決事項）。
- [`docs/progress.md`](./docs/progress.md) — 進捗管理・決定ログ（Why 付き）。
- [`HANDOFF.md`](./HANDOFF.md) — 作成時点の調査スナップショット（**不変**。本家 README/トップページの読み取り）。
- タスク（agent-tasks の worktree）をまたいで参照する情報は `docs/` に集約し、**実装の節目で更新する**。

## タスク管理（agent-tasks）

このプロジェクトの開発タスクは **agent-tasks skill** で管理する（中央ストア `~/agent-tasks-store`、
project キー = リポジトリ名 `herdr-hint`）。リポジトリ内にはタスク状態を置かない。

- 一覧: `agent-tasks`（現在 project）。詳細: `agent-tasks show herdr-hint <ID>`。
- 着手は `start`（git worktree で並行開発）、完了は `done`、レビュー待ちは `done --review`（status=review）、保留は `block`。
  操作手順は agent-tasks skill に従う。
- タスクファイル更新後は `agent-tasks sync <ID>`（scoped sync）でストアを同期する。
- タスク分割の全体像は [`docs/plan.md`](./docs/plan.md) §4 を参照。

## レビュー役 / ユーザー レビュー運用（必須）

このプロジェクトは agmsg の team `herdr-hint` に **実装役** と **レビュー役** が参加している。
実装役はレビュー役とユーザーのレビューを受けながら進める。

**現在の役割構成**: 実装役 = `claude`（type: `claude-code`）／レビュー役 = `codex`（type: `codex`）。
役割は固定ではない。状況に応じて実装役とレビュー役を入れ替えたり、両役を Claude（別 pane の独立エージェント）が
担ったりできる。構成を変えたらこの節を更新する。

**往復の手順は agmsg-review skill に従う。** 依頼文の書き方、相手の起こし方（codex は send だけでは
気づかないので `herdr agent prompt <pane_id> '$agmsg' --wait --until working` が必須。`pane run` は
bracketed-paste で Enter が食われるため使わない）、指摘の裏取り、承認が出るまでのループとラウンド上限
（既定 3 回）はすべてあちらにある。ここには重複して書かない。以下はこのプロジェクト固有の決めごとで、
agmsg-review の既定より優先される。

- **いつレビューを受けるか**:
  - タスクの**重要な設計判断・実装ポイント**に差しかかったとき（方針を固める前）。
  - **タスクを完了（done / review）にする前**（コミット・PR 化の前が望ましい）。
- **レビュー役が不在なら agmsg-review の手順で spawn する**（herdr の pane を開く。agmsg の
  `spawn.sh` は使わない）。team は `herdr-hint`、name は `codex`、type は `codex`。
- **タスク完了前にはユーザーのレビューも受ける**: レビュー役の承認が得られたら、done / review にする前に
  ユーザーに成果を提示してレビュー・承認を得る。
- レビューのやり取り・結論は `docs/progress.md` の決定ログに要点を残す。

## ビルド / 実行 / 構成

- **言語 / TUI**: Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea)（スタイリング Lip Gloss、
  コンポーネント Bubbles を併用可）。
- **ビルド**: `go build -o herdr-hint .`（`herdr-plugin.toml` の build command もこれに合わせる）。
- **herdr との通信**: CLI(JSON) 経由。取得 `herdr workspace/tab/agent list`、操作 `herdr tab/agent focus <id>`。
  herdr バイナリは env `HERDR_BIN_PATH`、無ければ `herdr`。
- **プラグイン定義**: `herdr-plugin.toml`（`id` は本家 `maedana.hint` から `shunichi.hint` 等へ変更）。
- 詳細な構成は [`docs/plan.md`](./docs/plan.md) §2 を参照。

### 技術選定の既定

- **TUI ツールは Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) を採用する**。
  スタイリングは Lip Gloss、コンポーネントは Bubbles を併用してよい。別スタックにする明確な理由が
  あるときのみ、その理由を docs に書いた上で外れる（本プロジェクトはこの既定に従う）。

### 依存ライブラリの方針

- **基本は最新安定版を使う**。追加/更新時は最新の安定版（プレリリース `-alpha/-beta/-rc` を除く）を
  確認して採用する（Go なら `go list -m -versions <module>`）。
  メジャーバージョンでモジュールパス/パッケージ名が変わる場合があるので注意する。
- 最新安定版を**あえて使わない**場合は、その理由を依存定義の近辺のコメントか docs に明記する。
