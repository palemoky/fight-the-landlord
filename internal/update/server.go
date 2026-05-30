package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// requirementTimeout 是查询服务端版本要求的最大耗时
const requirementTimeout = 2 * time.Second

// ServerRequirement 描述服务端对客户端版本的要求
//
// 由服务端 /version 接口返回，使升级策略由服务端集中控制：服务端只在确有不兼容变更时抬高 MinClientVersion，而非每次发版都强制所有人升级
type ServerRequirement struct {
	ServerVersion    string `json:"server_version"`     // 服务端自身的版本
	MinClientVersion string `json:"min_client_version"` // 服务端要求的最低客户端版本，空字符串表示不限制
}

// RequiresUpgrade 判断 currentVersion 是否低于服务端要求的最低版本。
//
// 当服务端未设置 MinClientVersion（空）时永远返回 false，即不强制升级
func (r *ServerRequirement) RequiresUpgrade(currentVersion string) bool {
	if r == nil || r.MinClientVersion == "" {
		return false
	}
	return CompareVersions(currentVersion, r.MinClientVersion) < 0
}

// FetchServerRequirement 从 serverURL 对应的服务端查询版本要求
//
// serverURL 为客户端连接所用的 WebSocket 地址（ws:// 或 wss://），本函数会据此推导出同源的 http(s)://host/version 地址。任何网络或解析错误都以 error 返回，调用方应将其视为「无法判断」而非「需要升级」。
func FetchServerRequirement(ctx context.Context, serverURL string) (*ServerRequirement, error) {
	endpoint, err := versionEndpoint(serverURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, requirementTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update: version endpoint returned status %d", resp.StatusCode)
	}

	var out ServerRequirement
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// versionEndpoint 把 WebSocket 服务地址转换为同源的 /version HTTP 地址
func versionEndpoint(serverURL string) (string, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return "", err
	}

	switch u.Scheme {
	case "wss", "https":
		u.Scheme = "https"
	case "ws", "http":
		u.Scheme = "http"
	default:
		return "", fmt.Errorf("update: unsupported server scheme %q", u.Scheme)
	}

	u.Path = "/version"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}
