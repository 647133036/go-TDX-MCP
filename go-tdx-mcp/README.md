# TDX Finance MCP

通达信金融数据 MCP 服务器，提供 A 股/港股/美股实时行情、K 线、技术指标、缠论分析、量化回测、基金持仓、加密货币等金融数据服务。

支持两种运行模式：

- **MCP 模式**：标准 MCP 协议，通过 stdin/stdout 与 Claude Desktop、Cursor 等 AI 工具集成
- **Web API 模式**：RESTful HTTP API + WebSocket 实时推送，可直接浏览器访问

## 特性

- **实时行情**：A 股/港股/美股/期货实时报价
- **K 线数据**：日线/周线/月线/5/15/30/60 分钟线，支持前复权/后复权
- **技术指标**：34 种指标（MACD/KDJ/RSI/BOLL/DMI/ATR/WR/CCI/OBV/VR 等），一键批量计算
- **缠论分析**：自动识别笔、线段、中枢、买卖点
- **量化回测**：内置 12 种策略（均线交叉、海龟、布林带等）
- **财务数据**：利润表/资产负债表/现金流量表
- **资金流向**：主力资金/北向资金/南向资金
- **板块分析**：行业/概念/地域板块排行与成分股
- **基金数据**：基金净值/持仓/搜索
- **加密货币**：BTC/ETH 等主流币种行情
- **WebSocket 推送**：实时行情每 3 秒推送

## 安装

```bash
# 克隆仓库
git clone https://github.com/647133036/go-TDX-MCP.git
cd go-TDX-MCP

# 编译
go build -o go-tdx-mcp .

# 或下载预编译版本
# https://github.com/647133036/go-TDX-MCP/releases
```

## 运行

### Web API 模式

```bash
# 默认端口 8000
./go-tdx-mcp --web

# 自定义端口
./go-tdx-mcp --web --port=9000

# 使用配置文件
./go-tdx-mcp config.json --web

# 环境变量配置
TDX_TOKEN=your_token ./go-tdx-mcp --web --port=8000
```

启动后可通过浏览器访问：
- API 文档首页：http://localhost:8000/
- 健康检查：http://localhost:8000/api/v1/health

### MCP 模式

```bash
# 标准 MCP 模式（stdin/stdout）
./go-tdx-mcp
```

在 Claude Desktop 中配置：

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

## 配置

创建 `config.json` 文件：

```json
{
  "token": "your_tdx_token",
  "timeout": 30,
  "web_port": 8000,
  "tdx_host": "",
  "tdx_port": 0
}
```

环境变量优先级高于配置文件：

| 变量 | 说明 |
|------|------|
| `TDX_TOKEN` | 通达信 HTTP API Token |
| `TDX_HOST` | 通达信服务器地址 |
| `TDX_PORT` | 通达信服务器端口 |

## API 端点

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
│   ├── tools_expanded.go      # 扩展工具集
│   ├── tools_new.go           # 新增工具集
│   └── v3_tools.go            # V3 工具集
├── indicator/                 # 技术指标计算
├── backtest/                  # 量化回测引擎
├── chanlun/                   # 缠论分析引擎
├── finance/                   # 财务报表解析
├── offline/                   # 离线数据读写（TDX vipdoc）
├── scraper/                   # 网页数据爬虫
└── config.json                # 配置文件
```

## 数据源

- **TDX TCP**：通达信行情服务器（实时行情、K 线、财务等）
- **东方财富 API**：push2his / push2delay / datacenter（K 线、财务、板块等）
- **新浪财经**：行情数据补充
- **本地 TDX 数据**：vipdoc 目录下的日线/分钟线/板块文件

## 技术栈

- **语言**：Go 1.22+
- **框架**：gorilla/mux（HTTP）、mark3labs/mcp-go（MCP 协议）
- **TDX 协议**：gotdx（通达信二进制协议）
- **爬虫**：chromedp（Chrome 无头浏览器）、net/http

## 致谢

感谢以下开源项目为本项目提供技术支持：

- [gotdx](https://github.com/bensema/gotdx) — 通达信 TCP 协议 Go 实现
- [mcp-go](https://github.com/mark3labs/mcp-go) — Model Context Protocol SDK
- [gorilla/websocket](https://github.com/gorilla/websocket) — WebSocket 库
- [chromedp](https://github.com/chromedp/chromedp) — Chrome 无头浏览器自动化
- [goquery](https://github.com/PuerkitoBio/goquery) — jQuery 风格 HTML 解析

数据来源：

- 通达信行情服务器
- 东方财富 push2 API
- 新浪财经

## 许可证

MIT
