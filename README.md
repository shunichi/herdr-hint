# herdr-hint

[herdr](https://github.com/maedana/herdr)（ターミナルマルチプレクサ）用プラグイン。
Vimium 風の hint ラベルを **エージェント**に付け、キー入力で任意のエージェントへジャンプする。
[maedana/herdr-hint](https://github.com/maedana/herdr-hint) をベースにした個人用クローンで、
**Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea)** で実装している。

## 本家との違い

- **選択対象はエージェントのみ**（本家はタブ＋エージェント。タブはジャンプ対象から外している）。
- **各エージェントに種別（claude / codex など）を表示する**。
- **エージェントをワークスペース別にグループ表示する**（本家の agent 一覧はフラット）。

表示項目（リポジトリ:ブランチ / ステータス / ターミナルタイトル）は本家踏襲。

## 表示と操作

各行は `[ラベル] 種別 ステータス リポジトリ:ブランチ タイトル` の形式で、フォーカス中のエージェントには
`*` が付く。ワークスペースごとにグループ化して表示する。

| キー | 動作 |
|------|------|
| ラベル文字（`a`〜、多い場合は 2 文字） | そのエージェントにフォーカスを移す |
| `Ctrl+D` / `Ctrl+U` | 半画面スクロール |
| `Esc` / `Ctrl+C` | キャンセル |
| `Backspace` | 2 文字入力の 1 文字目を取り消す |

- ラベルはエージェント数が 26 以下なら 1 文字（`a`〜`z`）、27 以上なら 2 文字（`aa`〜`zz`）。
- **制約**: ラベル上限は 676 件。677 件以上のエージェントがある場合、超過分はラベルが付かず選択できない
  （画面下部にその旨を表示する）。

## インストール

```sh
herdr plugin install shunichi/herdr-hint
```

キーバインド（`~/.config/herdr/config.toml`）:

```toml
[[keys.command]]
key = "prefix+f"
type = "shell"
command = "herdr plugin pane open --plugin shunichi.hint --entrypoint jump"
description = "hint jump"
```

反映は `herdr server reload-config`。

## 開発

```sh
go build -o herdr-hint .   # ビルド（herdr-plugin.toml の build と同じ）
go test ./...              # テスト
go vet ./...

herdr plugin link /path/to/herdr-hint   # ローカル開発リンク
```

- herdr とは専用 SDK ではなく **CLI(JSON) 経由**で通信する（`herdr workspace list` /
  `herdr agent list` で取得、`herdr agent focus <id>` でフォーカス）。
- herdr バイナリは環境変数 `HERDR_BIN_PATH`、無ければ PATH 上の `herdr` を使う。
- 構成: `main.go`（エントリ）、`internal/herdr`（CLI クライアント・git_context）、
  `internal/label`（ラベル割当・解決）、`internal/ui`（グループ化・描画・Bubble Tea model）。
