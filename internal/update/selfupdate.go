package update

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/minio/selfupdate"
)

// downloadTimeout 是下载新版本二进制的最大耗时。
const downloadTimeout = 60 * time.Second

// AssetName 返回当前平台对应的 Release 资源文件名，与 release 工作流产物命名保持一致。
func AssetName(goos, goarch string) string {
	name := fmt.Sprintf("fight-the-landlord-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name
}

// downloadURL 拼接指定 tag、资源文件的下载地址。
func downloadURL(tag, asset string) string {
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", Repo, tag, asset)
}

// Apply 下载 res.LatestVersion 对应的二进制并原子替换当前可执行文件。
//
// 下载后会用 .sha256 校验文件完整性，校验失败或写入失败时返回 error 且不会破坏当前已安装的可执行文件（selfupdate 在失败时回滚）。
func Apply(ctx context.Context, res *Result) error {
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	asset := AssetName(runtime.GOOS, runtime.GOARCH)
	binURL := downloadURL(res.LatestVersion, asset)

	// 下载新二进制到内存
	data, err := fetch(ctx, binURL)
	if err != nil {
		return fmt.Errorf("下载新版本失败: %w", err)
	}

	// 下载并校验 sha256（校验文件缺失时跳过，与安装脚本行为一致）
	if sum, err := fetch(ctx, binURL+".sha256"); err == nil {
		if err := verifyChecksum(data, sum); err != nil {
			return err
		}
	}

	if err := selfupdate.Apply(bytes.NewReader(data), selfupdate.Options{}); err != nil {
		// 尝试回滚到替换前的可执行文件
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("升级失败且回滚失败: %w", rerr)
		}
		return fmt.Errorf("升级失败（已回滚）: %w", err)
	}
	return nil
}

// Relaunch 以新版本的可执行文件重新启动客户端，沿用当前进程的参数、环境与终端
//
// 该调用会阻塞直到子进程退出，随后以子进程的退出码结束当前进程，从而把控制权平滑交接给新版本
func Relaunch() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	args := os.Args[1:]
	// 防止子进程再次触发升级导致循环
	args = append(args, "-no-update-check")

	// #nosec G204,G702 -- exePath 来自 os.Executable，且不经过 shell，参数直接传递给目标进程。
	cmd := exec.Command(exePath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	os.Exit(0)
	return nil
}

// fetch 下载 url 的全部内容
func fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// verifyChecksum 用 .sha256 文件内容校验二进制数据完整性
//
// 校验文件格式为 `<hex>  <filename>`，仅取首段十六进制摘要比较
func verifyChecksum(data, checksumFile []byte) error {
	fields := strings.Fields(string(checksumFile))
	if len(fields) == 0 {
		return fmt.Errorf("校验文件为空")
	}
	want := strings.ToLower(fields[0])

	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if got != want {
		return fmt.Errorf("文件校验失败: 期望 %s, 实际 %s", want, got)
	}
	return nil
}
