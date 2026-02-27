BINARY  := pc_cleaner
VERSION := 0.1.0
MODULE  := github.com/seomini/pc_cleaner

LDFLAGS := -ldflags "-X main.version=$(VERSION)"
INSTALL_DIR := /usr/local/bin

# 기본 타겟
.DEFAULT_GOAL := build

# ── 빌드 ──────────────────────────────────────────────────────
.PHONY: build
build:
	go build $(LDFLAGS) -o $(BINARY) .

# ── 크로스컴파일 ───────────────────────────────────────────────
.PHONY: build-all
build-all:
	@mkdir -p bin
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)_darwin_amd64 .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)_darwin_arm64 .
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)_linux_amd64  .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)_windows_amd64.exe .
	@echo "✓ 전체 빌드 완료 → bin/"

# ── 클린 ──────────────────────────────────────────────────────
.PHONY: clean
clean:
	rm -f $(BINARY) $(BINARY).exe $(BINARY)_linux $(BINARY)_darwin
	rm -rf bin/
	@echo "✓ 빌드 산출물 삭제 완료"

# ── 설치 (/usr/local/bin) ─────────────────────────────────────
.PHONY: install
install: build
	install -m 755 $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "✓ 설치 완료: $(INSTALL_DIR)/$(BINARY)"

# ── 설치 삭제 ─────────────────────────────────────────────────
.PHONY: uninstall
uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY)
	@echo "✓ 삭제 완료: $(INSTALL_DIR)/$(BINARY)"

# ── 실행 (dry-run) ────────────────────────────────────────────
.PHONY: run
run: build
	./$(BINARY) --dry-run

# ── 테스트 ────────────────────────────────────────────────────
.PHONY: test
test:
	go test ./... -v

# ── 도움말 ────────────────────────────────────────────────────
.PHONY: help
help:
	@echo ""
	@echo "  make build      바이너리 빌드"
	@echo "  make build-all  전체 플랫폼 크로스컴파일 → bin/"
	@echo "  make clean      빌드 산출물 삭제"
	@echo "  make install    /usr/local/bin 에 설치"
	@echo "  make uninstall  설치된 바이너리 삭제"
	@echo "  make run        빌드 후 dry-run 실행"
	@echo "  make test       유닛 테스트 실행"
	@echo ""
