ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
BIN_DIR := $(ROOT)/bin
FRONT_DIST := $(ROOT)/apps/stock-front/dist
EMBED_DIST := $(ROOT)/apps/stock-manage/internal/static/dist

.PHONY: build-front build-center build-manage build-all pack-mac pack-windows

build-front:
	cd apps/stock-front && pnpm install && pnpm build
	mkdir -p $(EMBED_DIST)
	rm -rf $(EMBED_DIST)/*
	cp -r $(FRONT_DIST)/* $(EMBED_DIST)/

build-center:
	mkdir -p $(BIN_DIR)
	cd apps/stock-center && go build -o $(BIN_DIR)/stock-center ./cmd/server

build-manage: build-front
	mkdir -p $(BIN_DIR)
	cd apps/stock-manage && go build -o $(BIN_DIR)/stock-manage ./cmd/server

build-all: build-center build-manage

pack-mac:
	bash scripts/pack/pack-mac.sh

pack-windows:
	bash scripts/pack/pack-windows.sh
