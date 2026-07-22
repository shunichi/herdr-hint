# herdr-hint

[herdr](https://github.com/maedana/herdr)（ターミナルマルチプレクサ）用プラグイン。
任意のエージェントの pane へ素早くジャンプする。[maedana/herdr-hint](https://github.com/maedana/herdr-hint)
をベースにした個人用クローンで、**Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea)** で実装。

2 つのピッカーを持つ（同一プラグインの別 entrypoint）:

- **filter**（既定・`prefix+f`）: プロジェクト名でインクリメンタルに絞り込み、カーソルで選んでジャンプ。
- **hint**（entrypoint `jump`）: Vimium 風のラベルを付けて 1〜2 文字で選ぶ方式（残置。キーバインドは未割当）。

## 本家との違い

- **選択対象はエージェントのみ**（本家はタブ＋エージェント）。
- **各エージェントに種別（claude / codex など）を表示する**。
- **エージェントをワークスペース別にグループ表示する**（本家の agent 一覧はフラット）。
- インクリメンタル絞り込みの filter コマンドを追加。

表示項目（種別 / ステータス / リポジトリ:ブランチ / ターミナルタイトル）は共通。

## filter（既定コマンド）

入力欄にプロジェクト名の一部を打つと、**fuzzy マッチ**でワークスペースを絞り込む。候補はワークスペース別に
グループ表示され、カーソル行（`>`）のエージェントにジャンプする。

| キー | 動作 |
|------|------|
| 文字入力 | プロジェクト名でインクリメンタル絞り込み（fuzzy） |
| `↑` / `Ctrl+P` | カーソルを上へ |
| `↓` / `Ctrl+N` | カーソルを下へ |
| `Enter` | カーソル位置のエージェントの pane へ移動 |
| `Backspace` | 絞り込み文字を 1 つ削除 |
| `Esc` / `Ctrl+C` | キャンセル |

## hint（ラベル方式・残置）

各行に `[ラベル]` を付け、ラベル文字（26 以下は 1 文字 `a`〜、27 以上は 2 文字 `aa`〜）でジャンプ。
`Ctrl+D`/`Ctrl+U` スクロール、`Esc`/`Ctrl+C` キャンセル。ラベル上限は 676 件（超過分は選択不可、画面下部に表示）。
現在キーバインドは未割当。`herdr plugin pane open --plugin shunichi.hint --entrypoint jump` で起動できる。

## インストール

```sh
herdr plugin install shunichi/herdr-hint
```

キーバインド（`~/.config/herdr/config.toml`。既定は filter を `prefix+f` に割当）:

```toml
[[keys.command]]
key = "prefix+f"
type = "shell"
command = "herdr plugin pane open --plugin shunichi.hint --entrypoint filter"
description = "filter jump"
```

反映は `herdr server reload-config`。hint を使いたい場合は別キーに `--entrypoint jump` を割り当てる。

## 開発

`make` ターゲットで一通り操作できる:

```sh
make build      # go build -o herdr-hint .
make test       # go test ./...
make vet        # go vet ./...
make check      # vet + test
make install    # このワーキングコピーを herdr に登録（herdr plugin link。冪等）
make uninstall  # herdr から登録を外す
make clean      # バイナリ削除
```

`make install` はローカル開発リンク（`herdr plugin link "$(pwd)"`）で、実行後にキーバインド設定例を表示する
（キーバインドは各自の `~/.config/herdr/config.toml` に追記し、`herdr server reload-config` で反映）。
公開版を入れる場合は上記「インストール」の `herdr plugin install shunichi/herdr-hint`。

- herdr とは専用 SDK ではなく **CLI(JSON) 経由**で通信する（`herdr workspace list` /
  `herdr agent list` で取得、`herdr agent focus <id>` でフォーカス）。
- herdr バイナリは環境変数 `HERDR_BIN_PATH`、無ければ PATH 上の `herdr` を使う。
- サブコマンド: `herdr-hint`（hint）/ `herdr-hint filter`（filter）。
- 構成: `main.go`（entrypoint 分岐）、`internal/herdr`（CLI クライアント・git_context）、
  `internal/label`（hint のラベル割当・解決）、`internal/ui`（hint のグループ化・描画・model）、
  `internal/filter`（filter の絞り込み・カーソル・model。grouping は `internal/ui` を再利用）。
