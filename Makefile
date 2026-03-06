.PHONY: help release test coverage lint fmt clean build install proto

# 默认目标
.DEFAULT_GOAL := help

# 颜色输出
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
CYAN := \033[0;36m
NC := \033[0m # No Color

## help: 显示帮助信息
help:  ## Show this help message
	@echo "$(BLUE)════════════════════════════════════════$(NC)"
	@echo "$(BLUE)      Fight the Landlord - Makefile     $(NC)"
	@echo "$(BLUE)════════════════════════════════════════$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-15s$(NC) %s\n", $$1, $$2}'

## lint: 运行 linter
lint:  ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	golangci-lint run

## test: 运行所有测试
test:  ## Run all tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -race ./...

## coverage: 生成测试覆盖率报告
coverage:  ## Generate test coverage report
	@echo "$(BLUE)Generating coverage report...$(NC)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

## proto: 重新生成 Protocol Buffer 和消息类型映射代码
proto:  ## Regenerate Protocol Buffer and message type mapping code
	@echo "$(BLUE)Regenerating Protocol Buffer code...$(NC)"
	@if ! command -v protoc >/dev/null 2>&1; then \
		echo "$(RED)Error: protoc not found$(NC)"; \
		echo "$(YELLOW)Install it with: brew install protobuf$(NC)"; \
		exit 1; \
	fi
	protoc --proto_path=. --go_out=. --go_opt=module=github.com/palemoky/fight-the-landlord internal/protocol/proto/*.proto
	@echo "$(GREEN)✓ Protocol Buffer code regenerated$(NC)"
	@echo "$(BLUE)Generating MessageType mapping code...$(NC)"
	@cd internal/protocol/convert/msgtype && go run gen.go
	@echo "$(GREEN)✓ MessageType mapping code generated$(NC)"

## release: 创建并推送版本标签
release:  ## Create and push version tag
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "$(RED)Error: Working directory has uncommitted changes$(NC)"; \
		echo "$(YELLOW)Please commit or stash your changes before releasing$(NC)"; \
		exit 1; \
	fi; \
	LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "none"); \
	echo "$(BLUE)════════════════════════════════════════$(NC)"; \
	echo "$(BLUE)         Release New Version$(NC)"; \
	echo "$(BLUE)════════════════════════════════════════$(NC)"; \
	echo "$(CYAN)Current latest tag: $(GREEN)$$LATEST_TAG$(NC)"; \
	echo "$(BLUE)════════════════════════════════════════$(NC)"; \
	printf "$(YELLOW)Enter new version: $(NC)"; \
	read -r VERSION; \
	if [ -z "$$VERSION" ]; then \
		echo "$(RED)Error: Version cannot be empty$(NC)"; \
		exit 1; \
	fi; \
	if ! echo "$$VERSION" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "$(RED)Error: Invalid version format '$$VERSION'$(NC)"; \
		echo "$(YELLOW)Expected format: v1.0.0$(NC)"; \
		exit 1; \
	fi; \
	if git tag | grep -q "^$$VERSION$$"; then \
		echo "$(RED)Error: Tag $$VERSION already exists$(NC)"; \
		exit 1; \
	fi; \
	echo ""; \
	echo "$(YELLOW)About to create and push tag: $(GREEN)$$VERSION$(NC)"; \
	printf "$(YELLOW)Continue? [y/N] $(NC)"; \
	read -r CONFIRM; \
	if [ "$$CONFIRM" != "y" ] && [ "$$CONFIRM" != "Y" ]; then \
		echo "$(YELLOW)Aborted$(NC)"; \
		exit 1; \
	fi; \
	if [ "$$LATEST_TAG" != "none" ]; then \
		NEW_VER=$$(echo $$VERSION | sed 's/^v//'); \
		CUR_VER=$$(echo $$LATEST_TAG | sed 's/^v//'); \
		NEW_MAJOR=$$(echo $$NEW_VER | cut -d. -f1); \
		NEW_MINOR=$$(echo $$NEW_VER | cut -d. -f2); \
		NEW_PATCH=$$(echo $$NEW_VER | cut -d. -f3); \
		CUR_MAJOR=$$(echo $$CUR_VER | cut -d. -f1); \
		CUR_MINOR=$$(echo $$CUR_VER | cut -d. -f2); \
		CUR_PATCH=$$(echo $$CUR_VER | cut -d. -f3); \
		if [ $$NEW_MAJOR -lt $$CUR_MAJOR ] || \
		   ([ $$NEW_MAJOR -eq $$CUR_MAJOR ] && [ $$NEW_MINOR -lt $$CUR_MINOR ]) || \
		   ([ $$NEW_MAJOR -eq $$CUR_MAJOR ] && [ $$NEW_MINOR -eq $$CUR_MINOR ] && [ $$NEW_PATCH -le $$CUR_PATCH ]); then \
			echo "$(RED)Error: New version $$VERSION must be greater than $$LATEST_TAG$(NC)"; \
			exit 1; \
		fi; \
	fi; \
	if git config user.signingkey >/dev/null 2>&1 && command -v gpg >/dev/null 2>&1; then \
		echo "$(BLUE)Creating GPG signed tag $$VERSION...$(NC)"; \
		if git tag -s $$VERSION -m "Release $$VERSION" 2>/dev/null; then \
			echo "$(GREEN)✓ Signed tag $$VERSION created (Verified ✓)$(NC)"; \
		else \
			echo "$(YELLOW)⚠ GPG signing failed, using regular tag...$(NC)"; \
			git tag -a $$VERSION -m "Release $$VERSION"; \
			echo "$(GREEN)✓ Tag $$VERSION created$(NC)"; \
		fi \
	else \
		echo "$(BLUE)Creating tag $$VERSION...$(NC)"; \
		git tag -a $$VERSION -m "Release $$VERSION"; \
		echo "$(GREEN)✓ Tag $$VERSION created$(NC)"; \
		echo "$(YELLOW)💡 Tip: Configure GPG key to show Verified badge$(NC)"; \
	fi; \
	echo "$(BLUE)Pushing tag to remote...$(NC)"; \
	git push origin $$VERSION; \
	echo "$(GREEN)✓ Release $$VERSION completed$(NC)"
