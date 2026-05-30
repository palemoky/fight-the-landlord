<div align="center">
    <img src="https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/docs/logo.png" alt="Logo" height="100px" />

# 🎮 欢乐斗地主

**一个真正公平的斗地主游戏 - 无控牌、无算法操控、纯粹的运气与技巧**

基于 Go 语言实现的斗地主游戏，支持联网对战、断线重连、排行榜等功能。

[![Docker Image Size](https://img.shields.io/docker/image-size/palemoky/fight-the-landlord/latest)](https://hub.docker.com/r/palemoky/fight-the-landlord)
[![Test](https://github.com/palemoky/fight-the-landlord/actions/workflows/test.yml/badge.svg)](https://github.com/palemoky/fight-the-landlord/actions/workflows/test.yml)
[![Release](https://github.com/palemoky/fight-the-landlord/actions/workflows/release.yml/badge.svg)](https://github.com/palemoky/fight-the-landlord/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/palemoky/fight-the-landlord)](https://goreportcard.com/report/github.com/palemoky/fight-the-landlord)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

</div>

## 💡 项目初衷

厌倦了商业斗地主游戏的控牌机制？我也是。

在某些知名斗地主游戏中，新手或回归玩家刚开始会获得好牌，匹配豆子少的对手，营造"连胜"的错觉。但随着游戏时间增长，牌质量明显下降，且频繁匹配高段位玩家，导致快速输光豆子。这种算法操控严重破坏了游戏的公平性和纯粹性。

**本项目承诺**：

- ✅ **真随机发牌**：每局洗牌完全随机，无任何控牌算法
- ✅ **公平匹配**：不考虑胜率、段位、游戏时长，纯随机或房间匹配
- ✅ **开源透明**：所有代码公开，欢迎审计和贡献
- ✅ **无内购无广告**：纯粹的游戏体验，技巧决定胜负

> **核心理念**：斗地主应该是运气与技巧的博弈，而不是算法与钱包的较量。

## 📸 游戏截图

<div align="center">
  <img src="https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/docs/lobby.png" alt="Lobby" width="45%" />
  <img src="https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/docs/in-game.png" alt="In Game" width="45%" />
</div>

## 🤖 DouZero 机器人出牌演示

[DouZero](https://github.com/kwai/DouZero) 是快手开源的基于深度强化学习的斗地主 AI。相比于经常出非法牌型的 LLM 而言，DouZero 能展现出更丰富的高级策略：自由出牌时主动组合复杂牌型、农民间默契配合、精准顶牌与拆牌等，对局体验更接近真人对手。

<div align="center">
  <img src="https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/docs/douzero-game.png" alt="Game" width="45%" />
  <img src="https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/docs/douzero-log.png" alt="Log" width="45%" />
</div>

## ✨ 功能特性

| 功能         | 说明                                                                 |
| ------------ | -------------------------------------------------------------------- |
| 🎯 实时对战  | WebSocket 实时通信，支持大规模并发对战（每局 3 人）                  |
| 🏠 房间系统  | 创建房间、加入房间、快速匹配                                         |
| 🤖 AI 机器人 | 凑人不足时DouZero 神经网络引擎自动填充空位；内置规则启发式机器人兜底 |
| 🔄 断线重连  | 网络波动时自动重连，游戏状态完整恢复                                 |
| ⏸️ 离线等待  | 对手掉线时暂停计时，等待重连                                         |
| 🏆 排行榜    | 积分系统、胜率统计、实时排名                                         |
| 📲 聊天系统  | 支持大厅聊天、房间快捷消息，互动更便捷                               |
| 🔒 安全防护  | 来源验证、速率限制、IP 过滤                                          |
| 🐳 容器部署  | Docker Compose 一键部署，Docker Hardened Images (DHI) 基础镜像加固   |
| 🔄 优雅升级  | 维护模式 + 零停机发版，等待游戏结束后自动关闭                        |
| ⚡ 流量优化  | Protocol Buffers ~~+WebSocket~~ 压缩，节省 60-80% 流量               |
| 🚀 性能优化  | `sync.Pool` 对象池复用，降低 GC 压力，提升并发性能                   |
| 📝 日志记录  | 文件记录日志，便于调试和问题追踪                                     |

## 🚀 快速开始

### 客户端安装

**macOS / Linux**：

```bash
curl -fsSL https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/install.sh | bash
```

**Windows (PowerShell)**：

```powershell
irm https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/install.ps1 | iex
```

**运行客户端**：

```bash
ddz
```

### 服务端部署

**使用 Docker Compose（推荐）**：

```bash
# 1. 创建项目目录
mkdir fight-the-landlord && cd fight-the-landlord

# 2. 下载配置文件
curl -fsSL https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/docker-compose.yml -o docker-compose.yml
curl -fsSL https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/.env.example -o .env

# 3. 修改配置（可选）
vim .env

# 4. 启动服务
docker compose up -d

# 5. 停止服务
docker compose down
```

💡 推荐使用 [lazydocker](https://github.com/jesseduffield/lazydocker) 管理服务

### 本地开发

```bash
# 1. 启动 Redis
redis-server

# 2. 启动服务端
go run ./cmd/server

# 3. 启动客户端（开 3 个终端）
go run ./cmd/client
```

## 🎲 游戏规则

与常见的斗地主相同，开局叫地主后，两位农民需配合击败地主，地主则需要阻击两个农民，率先出完手牌的一方获胜。

### 牌型示例

```
单张: 3, K, 2
对子: 33, KK
三张: 333
三带一: 3334
三带二: 33344
顺子: 34567 (5张+)
连对: 334455 (3对+)
飞机: 333444 (两个连三+)
飞机带单: 33344456
飞机带对: 3334445566
四带二: 333345
四带两对: 33334455
炸弹: 3333
王炸: 小王大王
```

### 叫抢地主与倍数

**叫抢流程**：

1. 发牌后随机一名玩家开始，依次选择叫地主 / 不叫
2. 有人叫地主后进入抢地主阶段，其余玩家可抢，叫地主者可反抢——**每抢一次倍数翻倍**
3. 连续两人放弃后，最后抢到的人成为地主，叫抢累计的倍数即为底倍
4. 一圈无人叫则流局重新发牌；连续 3 次流局后随机强制指定地主，避免无限流局

**倍数来源**，最终倍数 = 底倍 × 各项翻倍：

| 来源        | 翻倍                            |
| ----------- | ------------------------------- |
| 抢 / 反抢   | 每次 ×2                         |
| 炸弹 / 王炸 | 每打出一个 ×2                   |
| 春天        | 地主获胜且农民一张未出 ×2       |
| 反春天      | 农民获胜且地主仅首攻出过一手 ×2 |

**本局结算**（地主独自对抗两名农民，与排行榜积分相互独立）：地主获胜得 `+2 ×倍数`、两农民各 `-倍数`；农民获胜则地主 `-2 ×倍数`、两农民各 `+倍数`。

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

---

<div align="center">

**让斗地主回归纯粹 - 无控牌，真公平**

Made with ❤️ by [palemoky](https://github.com/palemoky)

</div>
