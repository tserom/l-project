ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
BIN_DIR := $(ROOT)/bin
FRONT_DIST := $(ROOT)/apps/stock-front/dist
EMBED_DIST := $(ROOT)/apps/stock-manage/internal/static/dist

.PHONY: build-front build-center build-manage build-sales-manage build-all pack-mac pack-windows dev-sales-front dev-sales-manage build-sales-front pack-sales-front-windows

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

build-sales-manage:
	mkdir -p $(BIN_DIR)
	cd apps/sales-manage && go build -o $(BIN_DIR)/sales-manage ./cmd/server

build-all: build-center build-manage build-sales-manage

pack-mac:
	bash scripts/pack/pack-mac.sh

pack-windows:
	bash scripts/pack/pack-windows.sh

dev-sales-manage:
	cd apps/sales-manage && go run ./cmd/server

dev-sales-front:
	cd apps/sales-front && pnpm install && pnpm dev

build-sales-front:
	cd apps/sales-front && pnpm install && pnpm build

pack-sales-front-windows:
	bash scripts/pack/pack-sales-front-windows.sh
