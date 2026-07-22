BINARY := herdr-hint
PLUGIN_ID := shunichi.hint

.PHONY: build test vet check install uninstall clean

build:
	go build -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

check: vet test

## install: このワーキングコピーを herdr に登録する。
## `herdr plugin link` はビルドしない（build するのは GitHub からの `plugin install`）ので、
## 先に build してバイナリを plugin_root に用意してから link する。無いと pane コマンド
## `${HERDR_PLUGIN_ROOT}/herdr-hint` が起動できず popup が一瞬で閉じる。
## 既存の同 id 登録があれば best-effort で外してから link し直す（冪等）。
install: build
	-herdr plugin unlink $(PLUGIN_ID) 2>/dev/null
	-herdr plugin uninstall $(PLUGIN_ID) 2>/dev/null
	herdr plugin link "$$(pwd)"
	@echo
	@echo "installed $(PLUGIN_ID). キーバインド例（~/.config/herdr/config.toml）:"
	@echo '  command = "herdr plugin pane open --plugin $(PLUGIN_ID) --entrypoint filter"'
	@echo "設定変更後は: herdr server reload-config"

## uninstall: herdr から登録を外す（link/install どちらでも best-effort）。
uninstall:
	-herdr plugin unlink $(PLUGIN_ID) 2>/dev/null
	-herdr plugin uninstall $(PLUGIN_ID) 2>/dev/null

clean:
	rm -f $(BINARY)
