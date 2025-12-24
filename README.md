# Fight the Landlord

启动服务端

```
# 确保 Redis 在运行
redis-server
# 启动服务端 (默认端口 1780)
go run ./cmd/server
# 或使用自定义配置
go run ./cmd/server -config configs/config.yaml
```

启动客户端

```
# 连接本地服务器
go run ./cmd/client
# 连接远程服务器
go run ./cmd/client -server 192.168.1.100:1780
```

```mermaid
graph TD
    %% --- 样式定义 ---
    %% 绿色系代表用户端
    classDef green fill:#e9f5e9,stroke:#81c784,stroke-width:2px,color:#2e7d32;
    %% 黄色系代表处理逻辑
    classDef yellow fill:#fff8e1,stroke:#ffd54f,stroke-width:2px,color:#f57f17;
    %% 灰色/蓝色系代表数据库
    classDef grey fill:#eceff1,stroke:#90a4ae,stroke-width:2px,color:#37474f;

    %% --- 图表结构 ---
    subgraph Clients [Clients]
        C1[TUI Client 1]:::green
        C2[TUI Client 2]:::green
        C3[TUI Client 3]:::green
    end

    subgraph Server [Server Layer]
        WH[WebSocket Handler]:::yellow
        RM[Room Manager]:::yellow
        GM[Game Manager]:::yellow
        MM[Match Maker]:::yellow
    end

    subgraph External [Storage]
        Redis[(Redis)]:::grey
    end

    %% --- 连接关系 ---
    C1 <--> WH
    C2 <--> WH
    C3 <--> WH

    WH --> RM
    WH --> GM
    WH --> MM

    RM --> Redis
    MM --> Redis

    %% --- 背景框样式 ---
    style Clients fill:#ffffff,stroke:#ccc,stroke-dasharray: 5 5
    style Server fill:#ffffff,stroke:#ccc,stroke-dasharray: 5 5
    style External fill:#ffffff,stroke:#ccc,stroke-dasharray: 5 5
```

```mermaid
stateDiagram-v2
    direction TB

    %% --- 样式 ---
    classDef waitState fill:#f5f5f5,stroke:#d7ccc8,stroke-width:2px,color:#5d4037
    classDef playState fill:#e3f2fd,stroke:#90caf9,stroke-width:2px,color:#1565c0
    classDef endState fill:#ffebee,stroke:#ef9a9a,stroke-width:2px,color:#c62828

    %% --- 状态 ---
    state "Waiting<br/>(等待玩家)<br/>人数不足3人持续等待" as WaitingStr
    state "Ready<br/>(准备就绪)" as ReadyStr
    state "Bidding<br/>(叫地主)" as BiddingStr
    state "Playing<br/>(游戏中)" as PlayingStr
    state "GameOver<br/>(结算)" as GameOverStr

    %% --- 主流程 ---
    [*] --> WaitingStr : 创建房间
    WaitingStr --> ReadyStr : 3人齐且全部准备
    ReadyStr --> BiddingStr : 发牌完成
    BiddingStr --> PlayingStr : 叫地主完成
    PlayingStr --> GameOverStr : 有人出完牌
    GameOverStr --> [*] : 解散房间
    GameOverStr --> WaitingStr : 再来一局

    %% --- 样式 ---
    class WaitingStr,ReadyStr waitState
    class BiddingStr,PlayingStr playState
    class GameOverStr endState
```

```mermaid
graph LR
    %% --- 莫兰迪配色定义 ---
    %% 启动阶段：淡雅的鼠尾草绿
    classDef setup fill:#E0F2F1,stroke:#80CBC4,stroke-width:2px,color:#00695C;
    %% 决策节点：柔和的杏色
    classDef decision fill:#FFF3E0,stroke:#FFCC80,stroke-width:2px,color:#E65100;
    %% 中间过程：淡紫灰
    classDef process fill:#F3E5F5,stroke:#CE93D8,stroke-width:2px,color:#6A1B9A;
    %% 游戏核心：静谧的雾霾蓝
    classDef game fill:#E8EAF6,stroke:#9FA8DA,stroke-width:2px,color:#283593;

    %% --- 节点定义 ---
    Start(启动):::setup
    Connect(连接服务器):::setup
    Lobby(大厅界面):::setup

    UserSelect{用户选择}:::decision

    Input(输入房间号):::process
    Matching(匹配中...):::process

    Waiting(等待界面):::process
    GameScreen(游戏界面):::game

    GameOver{游戏结束}:::decision

    %% --- 连接流程 ---
    %% 启动流程
    Start --> Connect --> Lobby --> UserSelect

    %% 分支流程
    UserSelect -- 创建房间 --> Waiting
    UserSelect -- 加入房间 --> Input --> Waiting
    UserSelect -- 快速匹配 --> Matching --> Waiting

    %% 游戏流程
    Waiting --> GameScreen --> GameOver

    %% 循环/返回流程
    GameOver -- 再来一局 --> Waiting
    GameOver -- 返回大厅 --> Lobby

    %% --- 线条样式优化 ---
    linkStyle default stroke:#90A4AE,stroke-width:2px,fill:none;
```
