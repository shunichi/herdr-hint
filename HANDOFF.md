# HANDOFF: herdr-hint 自分用クローン

このリポジトリは [maedana/herdr-hint](https://github.com/maedana/herdr-hint) に似たものを
`shunichi` 用に作るためのもの。**まだ実装は空**（README / .gitignore のみ）。
以降の実装は別の AI エージェントが担当する前提で、README 段階でわかっていることを記録する。

> 注意: 下記は本家 herdr-hint の README とトップページから読み取った内容。
> ソースコード（`src/lib.rs` / `src/main.rs`）の詳細な調査はまだ行っていない。
> 実装前に本家リポジトリのソースを直接確認すること。

## これは何か

herdr（ターミナルマルチプレクサ）用の **Rust 製プラグイン**。
Vimium 風の hint ラベルを表示し、キー入力で任意のタブ / エージェントへジャンプする。

- キーを押すと、全タブ・全エージェントに 1〜2 文字のラベル（a-z / aa-zz）を付けて一覧表示
- ラベルを入力するとその対象にフォーカスが移る

## 機能（本家 README より）

- タブはワークスペースごとにグルーピングし、2 カラムで表示
- エージェントは repository / branch / status / terminal title を表示
- ラベル長は動的: 26 件以下は 1 文字、27 件以上は 2 文字
- スクロール: Ctrl+D / Ctrl+U、キャンセル: Esc

## リポジトリ構成（本家）

```
src/
  main.rs            # 実行エントリポイント（TUI ループ）
  lib.rs             # パース・ラベル割当・レンダリングの中核 + テスト
Cargo.toml
Cargo.lock
herdr-plugin.toml    # プラグイン定義（メタ情報・build・panes）
.gitignore
```

### Cargo.toml 依存（本家）

- edition = "2024"
- `serde`（derive 有効）
- `serde_json`
- `crossterm` 0.28
- **herdr SDK crate への依存はなし**（herdr とは CLI 経由で通信する）

### herdr-plugin.toml（本家、ほぼ全文）

```toml
id = "maedana.hint"
name = "Hint"
version = "0.1.0"
min_herdr_version = "0.7.0"
description = "Vimium-style hint labels to jump to any tab or agent"
platforms = ["linux", "macos"]

[[build]]
command = ["cargo", "build", "--release"]

[[panes]]
id = "jump"
title = "Hint Jump"
placement = "popup"
width = "60%"
height = "50%"
command = ["sh", "-c", "${HERDR_PLUGIN_ROOT}/target/release/herdr-hint"]
```

> 自分用に作る際は `id` を `maedana.hint` から `shunichi.hint` などに変える。

## herdr との通信方法（重要）

プラグインは **herdr の CLI を叩いて** 情報取得・操作を行う（専用 SDK ではない）。

- herdr バイナリのパスは環境変数 `HERDR_BIN_PATH`、なければ `herdr` を使う
- 情報取得（JSON を stdout で受け取る）:
  - `herdr workspace list`
  - `herdr tab list`
  - `herdr agent list`
- 操作:
  - `herdr tab focus <id>`
  - `herdr agent focus <id>`
- プラグイン実行環境: `${HERDR_PLUGIN_ROOT}` にビルド成果物が置かれる

### main.rs の流れ（本家、読み取れた範囲）

1. `workspace list` / `tab list` / `agent list` を JSON で取得しパース
2. タブ＋エージェントにラベルを割当（`assign_labels`）
3. crossterm で raw mode に入りポップアップ描画（`render`）、カーソル非表示
4. キー入力ループ:
   - Ctrl+D / Ctrl+U でスクロール
   - ラベル文字（2 文字モードなら 2 文字）で対象を解決（`resolve_input`）
   - Esc でキャンセル
5. raw mode を戻し、選択された対象に `tab focus` / `agent focus`

### lib.rs の主な関数（本家、名前のみ判明）

- `parse_workspace_labels`, `parse_tabs`, `parse_agents`
- `assign_labels`, `uses_double_labels`, `resolve_input`
- `render`（ワークスペース別 2 カラム表示）
- `git_context`（git コマンドで repo 名・branch を取得）
- `HintKind`（Tab / Agent の enum）, `HintItem`

## インストール / 開発（本家 README より）

- インストール: `herdr plugin install <owner>/<repo>`
- キーバインド（`~/.config/herdr/config.toml`）:
  ```toml
  [[keys.command]]
  key = "prefix+f"
  type = "shell"
  command = "herdr plugin pane open --plugin maedana.hint --entrypoint jump"
  description = "hint jump"
  ```
  → `herdr server reload-config` で反映
- ローカル開発: `herdr plugin link /path/to/herdr-hint`

## 次にやること（提案）

1. 本家 `src/lib.rs` / `src/main.rs` の実際のコードを取得して仕様を確定
2. `herdr <workspace|tab|agent> list` の実際の JSON スキーマを、動作中の herdr で確認
   （このマシンは `HERDR_ENV=1` = herdr 内で動作中なので実データが取れる）
3. Cargo プロジェクト初期化 → `herdr-plugin.toml` の `id` を自分用に変更して実装

## メタ情報

- owner: `shunichi`（gh は SSH でログイン済み）
- リポジトリ: https://github.com/shunichi/herdr-hint （public）
- ローカル: `/home/shun/dev/src/github.com/shunichi/herdr-hint`
