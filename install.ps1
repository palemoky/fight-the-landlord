# 斗地主客户端一键安装脚本 (Windows)
# 使用方法: irm https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

# 颜色输出
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    exit 1
}

# 获取最新版本
function Get-LatestVersion {
    Write-Info "获取最新版本..."
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/palemoky/fight-the-landlord/releases/latest"
        $version = $response.tag_name
        Write-Info "最新版本: $version"
        return $version
    }
    catch {
        Write-Error-Custom "无法获取最新版本: $_"
    }
}

# 下载二进制文件
function Download-Binary {
    param([string]$Version)

    $binaryName = "fight-the-landlord-windows-amd64.exe"
    $downloadUrl = "https://github.com/palemoky/fight-the-landlord/releases/download/$Version/$binaryName"

    Write-Info "下载客户端..."

    $tempDir = Join-Path $env:TEMP "fight-the-landlord-install"
    New-Item -ItemType Directory -Force -Path $tempDir | Out-Null

    $outputPath = Join-Path $tempDir $binaryName

    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $outputPath -UseBasicParsing
        Write-Info "下载完成"
        return $outputPath
    }
    catch {
        Write-Error-Custom "下载失败: $_"
    }
}

# 安装二进制文件
function Install-Binary {
    param([string]$BinaryPath)

    Write-Info "安装客户端..."

    # 安装到用户目录
    $installDir = Join-Path $env:USERPROFILE ".fight-the-landlord"
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null

    $targetPath = Join-Path $installDir "fight-the-landlord.exe"
    Copy-Item -Path $BinaryPath -Destination $targetPath -Force

    Write-Info "已安装到: $targetPath"

    # 添加到 PATH
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$installDir*") {
        Write-Info "添加到 PATH..."
        [Environment]::SetEnvironmentVariable(
            "Path",
            "$userPath;$installDir",
            "User"
        )
        $env:Path = "$env:Path;$installDir"
        Write-Info "已添加到 PATH (需要重启终端生效)"
    }

    # 清理临时文件
    Remove-Item -Path (Split-Path $BinaryPath) -Recurse -Force
}

# 主函数
function Main {
    Write-Host ""
    Write-Host "🎮 欢乐斗地主 - 客户端安装" -ForegroundColor Cyan
    Write-Host ""

    $version = Get-LatestVersion
    $binaryPath = Download-Binary -Version $version
    Install-Binary -BinaryPath $binaryPath

    Write-Host ""
    Write-Info "✅ 安装完成！"
    Write-Host ""
    Write-Host "🎮 开始游戏：" -ForegroundColor Cyan
    Write-Host "    fight-the-landlord" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "💡 提示：直接运行即可，已自动连接到官方服务器" -ForegroundColor Gray
    Write-Host ""
    Write-Host "注意: 如果命令未找到，请重启终端" -ForegroundColor Yellow
    Write-Host ""
}

Main
