# TDX Finance MCP

通达信金融数据 MCP 服务器，提供 A 股/港股/美股/加密货币等多市场金融数据服务。

212 个工具，覆盖实时行情、K 线、技术指标、缠论分析、量化回测、资金流向、板块分析、基金数据、宏观数据等场景。

支持三种运行模式：

- **MCP Stdio 模式**：标准 MCP 协议，通过 stdin/stdout 与 Claude Desktop、Cursor、Windsurf 等 AI 工具集成
- **MCP Web 模式**：SSE / Streamable HTTP / 混合模式，通过 HTTP 提供 MCP 协议接口
- **REST API 模式**：RESTful HTTP API + WebSocket 实时推送，可直接浏览器或任意 HTTP 客户端访问

## 工具总览

| 类别 | 数量 | 说明 |
|------|------|------|
| Core | 6 | 基础行情：实时报价、K 线、股票信息、选股、指标选择、API 数据 |
| Expanded | 64 | 扩展工具：行情、板块、资金流、缠论、回测、财务、公告、离线数据等 |
| V3 | 8 | 高级工具：市场概览、板块资金流、涨跌停、财务指标、宏观数据、新闻舆情、表格爬虫 |
| New | 134 | 新增工具：加密货币、基金净值、融资融券、龙虎榜、可转债、期货、因子计算、选股扫描、回测组合等 |

## 快速开始

### 安装

```bash
git clone https://github.com/647133036/go-TDX-MCP.git
cd go-TDX-MCP
go build -o go-tdx-mcp .
```

### 配置

创建 `config.json`：

```json
{
  "token": "your_tdx_token",
  "timeout": 30,
  "web_port": 8000,
  "tdx_host": "",
  "tdx_port": 0
}
```

或通过环境变量：

| 变量 | 说明 |
|------|------|
| `TDX_TOKEN` | 通达信 HTTP API Token |
| `TDX_HOST` | 通达信服务器地址 |
| `TDX_PORT` | 通达信服务器端口 |

环境变量优先级高于配置文件。

### 运行

```bash
# MCP Stdio 模式（默认）
./go-tdx-mcp

# MCP SSE 模式
./go-tdx-mcp --sse --port=8000

# MCP Streamable HTTP 模式
./go-tdx-mcp --streamable-http --port=8000

# 混合模式（MCP + REST API + WebSocket）
./go-tdx-mcp --web --port=8000
```

### 集成 Claude Desktop

```json
{
  "mcpServers": {
    "tdx-finance": {
      "command": "/path/to/go-tdx-mcp",
      "args": []
    }
  }
}
```

## 运行模式详解

### MCP Stdio 模式

标准 MCP 协议，适合与 AI 编程工具集成。

```bash
./go-tdx-mcp
```

支持的客户端：Claude Desktop、Cursor、Windsurf、VS Code MCP 插件等。

### MCP Web 模式

通过 HTTP 提供 MCP 协议接口，适合远程调用。

```bash
# SSE 模式
./go-tdx-mcp --sse --port=8000
# SSE 端点: http://localhost:8000/sse
# 消息端点: http://localhost:8000/message

# Streamable HTTP 模式
./go-tdx-mcp --streamable-http --port=8000
# 端点: http://localhost:8000/mcp
```

### 混合模式

同时提供 MCP Streamable HTTP、REST API 文档首页和 WebSocket 实时推送。

```bash
./go-tdx-mcp --web --port=8000
```

启动后可通过浏览器访问：
- API 文档首页：http://localhost:8000/
- MCP Streamable HTTP：http://localhost:8000/mcp
- 健康检查：http://localhost:8000/api/v1/health

## REST API 端点

### 行情数据

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/quotes` | GET | `codes` | 实时报价（支持多代码） |
| `/api/v1/bars` | GET | `code`, `market`, `period`, `count`, `fq_type` | K 线数据 |
| `/api/v1/symbol-info` | GET | `code`, `market` | 标的基本信息 |
| `/api/v1/quote-list` | GET | `count`, `sort_type` | 行情列表 |
| `/api/v1/security-count` | GET | `market` | 涨跌停统计 |
| `/api/v1/market-stat` | GET | - | 市场统计 |
| `/api/v1/market-overview` | GET | - | 市场概览 |

### 技术指标

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/indicator/list` | GET | - | 指标列表 |
| `/api/v1/indicator/compute` | POST | data, indicators | 指标计算 |
| `/api/v1/indicator/compute_all` | GET | `code`, `market`, `indicators` | 批量计算 |

### 缠论与回测

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/chanlun/analyze` | GET | `code`, `market`, `period`, `count` | 缠论分析 |
| `/api/v1/backtest/run` | GET | `code`, `market`, `strategy`, `count` | 量化回测 |

### 财务与公告

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/financial/report` | GET | `code`, `type` | 财务报表（lrb/fzb/llb） |
| `/api/v1/announcements` | GET | `code`, `count` | 公告列表 |

### 资金与板块

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/capital-flow` | GET | `code`, `market` | 资金流向 |
| `/api/v1/auction` | GET | `code`, `market` | 集合竞价 |
| `/api/v1/unusual` | GET | `market`, `count` | 异动监控 |
| `/api/v1/board/list` | GET | `board_type`, `top_n` | 板块列表 |
| `/api/v1/board/members` | GET | `board_symbol`, `count` | 板块成分股 |
| `/api/v1/board/ranking` | GET | `board_type`, `top_n` | 板块排行 |
| `/api/v1/belong-board` | GET | `code`, `market` | 标的所属板块 |

### 扩展市场

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/ex/markets` | GET | - | 扩展市场列表 |
| `/api/v1/ex/quote` | GET | `ex_market`, `code` | 扩展市场报价 |
| `/api/v1/ex/bars` | GET | `ex_market`, `code` | 扩展市场 K 线 |

### 数据爬虫

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/scraper` | GET | - | 全量股票列表 |
| `/api/v1/scraper/sector-boards` | GET | `board_type` | 板块数据 |
| `/api/v1/scraper/northbound-flow` | GET | - | 北向资金 |
| `/api/v1/scraper/northbound-stocks` | GET | `date` | 北向持股股票 |
| `/api/v1/scraper/northbound-holders` | GET | `date` | 北向持股机构 |
| `/api/v1/scraper/fund-nav` | GET | `code`, `count` | 基金净值 |
| `/api/v1/scraper/margin-trade` | GET | `code`, `type` | 融资融券 |
| `/api/v1/scraper/fund-holding` | GET | `fund_code`, `period` | 基金持仓 |
| `/api/v1/scraper/fund-search` | GET | `keyword` | 基金搜索 |
| `/api/v1/scraper/hkus-quote` | GET | - | 港股报价 |
| `/api/v1/scraper/crypto` | GET | `symbols` | 加密货币 |

### 离线数据

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/offline/home` | GET | - | 检测 TDX 数据目录 |
| `/api/v1/offline/daily` | GET | `code`, `market` | 日线数据 |
| `/api/v1/offline/min` | GET | `code`, `market` | 分钟线数据 |
| `/api/v1/offline/gbbq` | GET | - | 股本变迁 |
| `/api/v1/offline/blocks` | GET | `filename` | 板块文件 |

### 宏观与资讯

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/macro-data` | GET | `indicator`, `count` | 宏观经济数据 |
| `/api/v1/news-sentiment` | GET | `code`, `count` | 新闻情感 |

### WebSocket

| 端点 | 协议 | 说明 |
|------|------|------|
| `/ws/realtime/{symbol}` | WebSocket | 实时行情推送 |

### 系统信息

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/health` | GET | 健康检查 |
| `/api/v1/server-info` | GET | 服务器信息 |
| `/` | GET | API 文档首页 |

## 项目结构

```
go-TDX-MCP/
├── main.go                    # 入口：MCP/Web 双模式
├── web/
│   └── server.go              # Web API 服务器
├── tdx/
│   ├── tcp_client.go          # TDX TCP 客户端
│   ├── unified_client.go      # 统一客户端（TCP + HTTP）
│   ├── tools_expanded.go      # 扩展工具集（64 个）
│   ├── tools_new.go           # 新增工具集（134 个）
│   ├── tools_v3.go            # V3 工具集（8 个）
│   └── tools.go               # 核心工具集（6 个）
├── indicator/                 # 技术指标计算
├── backtest/                  # 量化回测引擎
├── chanlun/                   # 缠论分析引擎
├── finance/                   # 财务报表解析
├── offline/                   # 离线数据读写（TDX vipdoc）
├── scraper/                   # 网页数据爬虫
├── portfolio/                 # 投资组合优化
├── screen/                    # 选股扫描
└── config.json                # 配置文件
```

## 数据源

- **TDX TCP**：通达信行情服务器（实时行情、K 线、财务等）
- **东方财富 API**：push2his / push2delay / datacenter（K 线、财务、板块等）
- **新浪财经**：行情数据补充
- **腾讯证券**：融资融券数据
- **本地 TDX 数据**：vipdoc 目录下的日线/分钟线/板块文件
- **Binance API**：加密货币行情（免费）
- **CoinGecko API**：加密货币数据备用源

## 技术栈

- **语言**：Go 1.26+
- **MCP**：mark3labs/mcp-go v0.55.0
- **HTTP**：gorilla/mux + gorilla/websocket
- **TDX 协议**：gotdx（通达信二进制协议）
- **爬虫**：chromedp（Chrome 无头浏览器）、net/http、goquery

## 测试

```bash
go test ./...
```

## 许可证

MIT
