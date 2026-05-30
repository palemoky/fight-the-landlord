#!/bin/bash
# 斗地主客户端一键安装脚本
# 使用方法: curl -fsSL https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/install.sh | bash

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# 检测操作系统和架构
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            error "不支持的操作系统: $OS"
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            error "不支持的架构: $ARCH"
            ;;
    esac

    info "检测到系统: $OS-$ARCH"
}

# 获取最新版本
get_latest_version() {
    info "获取最新版本..."
    LATEST_VERSION=$(curl -fsSL https://api.github.com/repos/palemoky/fight-the-landlord/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        error "无法获取最新版本"
    fi

    info "最新版本: $LATEST_VERSION"
}

# 下载二进制文件
download_binary() {
    BINARY_NAME="fight-the-landlord-${OS}-${ARCH}"
    DOWNLOAD_URL="https://github.com/palemoky/fight-the-landlord/releases/download/${LATEST_VERSION}/${BINARY_NAME}"
    CHECKSUM_URL="${DOWNLOAD_URL}.sha256"

    info "下载客户端..."
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    # 下载二进制文件
    if ! curl -fsSL -o "$BINARY_NAME" "$DOWNLOAD_URL"; then
        error "下载失败: $DOWNLOAD_URL"
    fi

    # 下载校验和
    if ! curl -fsSL -o "${BINARY_NAME}.sha256" "$CHECKSUM_URL"; then
        warn "无法下载校验和文件，跳过校验"
    else
        info "验证文件完整性..."
        if ! sha256sum -c "${BINARY_NAME}.sha256" 2>/dev/null; then
            error "文件校验失败"
        fi
    fi

    info "下载完成"
}

# 安装二进制文件
install_binary() {
    info "安装客户端..."

    # 优先安装到 ~/.local/bin
    if [ -d "$HOME/.local/bin" ]; then
        INSTALL_DIR="$HOME/.local/bin"
    else
        # 尝试创建 ~/.local/bin
        mkdir -p "$HOME/.local/bin" 2>/dev/null && INSTALL_DIR="$HOME/.local/bin" || INSTALL_DIR="/usr/local/bin"
    fi

    # 如果需要 sudo 权限
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] && [ ! -w "$INSTALL_DIR" ]; then
        warn "需要 sudo 权限安装到 $INSTALL_DIR"
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/ddz"
        sudo chmod +x "$INSTALL_DIR/ddz"
    else
        mv "$BINARY_NAME" "$INSTALL_DIR/ddz"
        chmod +x "$INSTALL_DIR/ddz"
    fi

    info "已安装到: $INSTALL_DIR/ddz"

    # 检查 PATH
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        warn "$INSTALL_DIR 不在 PATH 中"
        warn "请将以下内容添加到 ~/.bashrc 或 ~/.zshrc:"
        echo ""
        echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
    fi

    # 清理临时文件
    cd - > /dev/null
    rm -rf "$TMP_DIR"
}

# 主函数
main() {
    echo ""
    echo "🎮 欢乐斗地主 - 客户端安装"
    echo ""

    detect_platform
    get_latest_version
    download_binary
    install_binary

    echo ""
    info "✅ 安装完成！"
    echo ""
    echo "🎮 开始游戏："
    echo "    ddz"
    echo ""
    echo "💡 提示：直接运行即可，已自动连接到官方服务器"
    echo ""
}

main
